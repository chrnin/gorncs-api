package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-contrib/cors"
	gorncs "github.com/signaux-faibles/gorncs-api/lib"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

var db string
var path string
var scanner bool
var initdb bool
var bind string
var download bool
var user string
var password string
var verbose bool
var limit int
var siren string
var database *sql.DB

func init() {
	flag.StringVar(&db, "DB", "./bilan.db", "chemin de la base sqlite3")
	flag.StringVar(&path, "path", ".", "chemin où sont stockés les fichiers RNCS")
	flag.StringVar(&bind, "bind", "127.0.0.1:3000", "port d'écoute de l'api")
	flag.BoolVar(&scanner, "scan", false, "importer les fichiers")
	flag.BoolVar(&initdb, "initdb", false, "créer une nouvelle base sqlite")
	flag.BoolVar(&verbose, "verbose", false, "afficher les informations d'importation")
	flag.BoolVar(&download, "download", false, "synchroniser le dépôt RNCS dans (voir -path, -user et -password)")
	flag.StringVar(&user, "user", "", "utilisateur FTPS RNCS/inpi")
	flag.StringVar(&password, "password", "", "mot de passe FTPS RNCS/inpi")
	flag.IntVar(&limit, "limit", 0, "limiter l'import à n bilans")
	flag.StringVar(&siren, "siren", "", "restreint l'importation au siren")
}

func main() {
	flag.Parse()
	if scanner {
		scan()
	} else if initdb {
		initDB()
	} else if download {
		err := gorncs.DownloadFolder(
			"ftp://opendata-rncs.inpi.fr/public/Bilans_Donnees_Saisies/",
			user,
			password,
			path)
		if err != nil {
			log.Print("Interruption du téléchargement: " + err.Error())
		}
	} else {
		var err error
		database, err = sql.Open("sqlite3", db)
		if err != nil {
			panic(err)
		}

		fmt.Println("gorncs-api écoute " + bind)
		fmt.Println("Pour plus d'information: gorncs-api --help")
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.Use(cors.Default())
		r.GET("/bilan/:siren", search)
		r.GET("/fields", fields)

		r.Run(bind)
	}
}

func fields(c *gin.Context) {
	for _, p := range gorncs.Postes {
		fmt.Println(p)
	}
}

type query struct {
	Siren string `json:"siren"`
}

func search(c *gin.Context) {
	siren := c.Params.ByName("siren")

	rows, err := database.Query("select * from bilan where siren = $1", siren)
	cols, _ := rows.Columns()

	if err != nil {
		c.JSON(500, "wtf")
	}

	var result []interface{}

	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			fmt.Println(err)
		}

		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if *val != nil {
				m[colName] = *val
			}
		}

		result = append(result, m)
	}
	c.JSON(200, result)
}

func initDB() {
	log.Print("initialisation de la base de données Sqlite pour gorncs: " + db)
	database, err := sql.Open("sqlite3", db)
	if err != nil {
		log.Fatal("Erreur d'accès au fichier " + db + ": " + err.Error())
	}

	createTableQuery := gorncs.GetCreateTableQuery()
	_, err = database.Exec(createTableQuery)
	if err != nil {
		log.Fatal("interruption lors de la création de la table: " + err.Error())
	} else {
		log.Print("creation de la table bilan (" + strconv.Itoa(len(gorncs.Postes)) + " champs): ok")
	}
	_, err = database.Exec("create unique index idx_lookup_bilan on bilan (nom_fichier, siren, date_cloture_exercice, code_activite, date_depot, denomination);")
	if err != nil {
		log.Print("creation index: " + err.Error())
	} else {
		log.Print("creation index: ok")
	}

	// log.Print("creation vue synthetique: not yet implemented")
	// log.Print("creation vue actif: not yet implemented")
	// log.Print("creation vue passif: not yet implemented")
	// log.Print("creation vue compte_de_resultat: not yet implemented")
	// log.Print("creation vue ratio: not yet implemented")
}

func scan() {
	log.Print("gorncs - analyse de l'arborescence INPI dans " + path)
	database, err := sql.Open("sqlite3", db)
	if err != nil {
		panic(err)
	}

	// options d'optimisation dangereuses
	// _, err = database.Exec("PRAGMA journal_mode = OFF")
	// _, err = database.Exec("PRAGMA synchronous = OFF")

	queryString := gorncs.GetQueryString()

	tx, _ := database.Begin()
	stmt, err := tx.Prepare(queryString)
	if err != nil {
		panic(err)
	}

	n := 0
	for bilan := range gorncs.BilanWorker(path) {
		if bilan.Siren == siren || siren == "" {
			if bilan.Siren != "" && len(bilan.Lignes) > 0 {
				_, err := stmt.Exec(bilan.ToQueryParams()...)
				if err != nil {

					if verbose {
						if err.Error()[0:6] == "UNIQUE" {
							log.Print("bilan déjà présent " + bilan.NomFichier)
						} else {
							log.Print("probleme à l'insert de " + bilan.NomFichier + ": " + err.Error())
						}
					}
				} else {
					n++
				}
				if verbose {
					log.Print("scan: " + bilan.NomFichier)
				}
			} else {
				if verbose {
					log.Print("aucune donnée: " + bilan.NomFichier)
				}
			}
		}
		if n == limit && limit != 0 {
			break
		}

	}

	stmt.Close()
	tx.Commit()
	database.Close()
	log.Print("Bilans importés: " + strconv.Itoa(n))
}
