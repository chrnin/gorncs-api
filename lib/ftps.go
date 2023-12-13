package gorncs

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	curl "github.com/andelf/go-curl"
)

var re = regexp.MustCompilePOSIX(`^([-d]{1})[-rwxd]{9}.*ftp.*ftp[ ]*([0-9]*).{14}(.*)$`)

func filenameFromURL(url string) string {
	segments := strings.Split(url, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}
	return ""
}

// CheckMD5 compare les MD5 du fichier de contrôle et du fichier téléchargé
func CheckMD5(path string) (bool, error) {
	var MD5String string
	zipFile, err := os.Open(path[:len(path)-3] + "zip")
	if err != nil {
		return true, err
	}

	md5File, err := os.Open(path[:len(path)-3] + "md5")
	if err != nil {
		return true, err
	}

	var MD5 = make([]byte, 32)
	_, err = md5File.Read(MD5)

	hash := md5.New()
	if _, err := io.Copy(hash, zipFile); err != nil {
		return true, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	MD5String = hex.EncodeToString(hashInBytes)
	if string(MD5) == MD5String {
		return true, nil
	}
	return false, nil
}

// DownloadFile télécharge un fichier si nécessaire et
// essaye de vérifier le checksum md5 s'il est fourni
func DownloadFile(url string, user string, password string, path string, size int) error {
	easy := curl.EasyInit()
	defer easy.Cleanup()
	if _, err := os.Stat(path); err == nil {
		log.Print(url + ": le fichier existe déjà")
		if path[len(path)-3:len(path)] == "zip" {
			isGoodMD5, err := CheckMD5(path)
			if err == nil {
				if isGoodMD5 {
					log.Print(url + ": MD5 valide")
				} else {
					log.Print(url + ": MD5 erroné, fichier supprimé, nouvelle tentative")
					os.Remove(path[:len(path)-3] + "zip")
					os.Remove(path[:len(path)-3] + "md5")
					os.Exit(0)

					DownloadFile(url, user, password, path, size)
				}
			}
		}
	} else if os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}

		handleFile := func(buf []byte, userdata interface{}) bool {
			_, err := f.Write(buf)
			if err != nil {
				log.Print(path + ": echec du téléchargement")
				os.Remove(path[:len(path)-3] + "zip")
				os.Remove(path[:len(path)-3] + "md5")
				return false
			}
			return true
		}

		easy.Setopt(curl.OPT_USE_SSL, true)
		easy.Setopt(curl.OPT_USERNAME, user)
		easy.Setopt(curl.OPT_PASSWORD, password)
		easy.Setopt(curl.OPT_URL, url)
		easy.Setopt(curl.OPT_WRITEFUNCTION, handleFile)

		if err := easy.Perform(); err != nil {
			return err
		}
		f.Close()

		log.Print(url + ": téléchargement effectué")
		isGoodMD5, err := CheckMD5(path)
		if err == nil {
			if isGoodMD5 {
				log.Print(url + ": MD5 valide")
			} else {
				log.Print(url + ": MD5 invalide, le fichier est supprimé. Pour l'obtenir relancez la procédure.")
				os.Remove(path[:len(path)-3] + "zip")
				os.Remove(path[:len(path)-3] + "md5")
			}
		}

	} else {
		// Schrodinger: file may or may not exist. See err for details.
		log.Print(url + ": problème d'accès au fichier (" + err.Error() + ") passe")
	}
	return nil
}

// DownloadFolder liste les fichiers présents dans un répertoire du ftp
func DownloadFolder(url string, user string, password string, path string) error {
	log.Print(url + "parcours du dossier ")

	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_USE_SSL, true)
	easy.Setopt(curl.OPT_USERNAME, user)
	easy.Setopt(curl.OPT_PASSWORD, password)
	easy.Setopt(curl.OPT_URL, url)
	var listing [][][]byte

	var handleData = func(buf []byte, userdata interface{}) bool {
		listing = re.FindAllSubmatch(buf, -1)
		return true
	}

	easy.Setopt(curl.OPT_WRITEFUNCTION, handleData)

	if err := easy.Perform(); err != nil {
		return err
	}

	for _, line := range listing {
		if string(line[1]) == "d" {
			os.Mkdir(path+"/"+string(line[3]), 0777)
			err := DownloadFolder(url+string(line[3])+"/", user, password, path+"/"+string(line[3]))
			if err != nil {
				return err
			}
		} else {
			size, err := strconv.Atoi(string(line[2]))
			if err != nil {
				return err
			}
			err = DownloadFile(url+string(line[3]), user, password, path+"/"+string(line[3]), size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
