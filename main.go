package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strconv"

	"github.com/chrnin/gorncs"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

var dial string
var db string
var collection string
var path string
var scanner bool
var initdb bool
var bind string

func init() {
	flag.StringVar(&db, "DB", "./bilan.db", "sqlite3 database path")
	flag.StringVar(&path, "path", ".", "RNCS root path")
	flag.StringVar(&bind, "bind", "127.0.0.1:3000", "Listen and serve on")
	flag.BoolVar(&scanner, "scanner", false, "Scan and import everything below the root path, doesn't run API endpoint")
	flag.BoolVar(&initdb, "initdb", false, "initialize a fresh new sqlite database")
}

func main() {
	flag.Parse()
	if scanner {
		scan()
	} else if initdb {
		initDB()
	} else {
		fmt.Println("gorncs-api listening on: " + bind)
		fmt.Println("for more information: gorncs-api --help")
		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.Use(cors.Default())
		r.GET("/:siren", search)
		r.Run(bind)
	}

}

type query struct {
	Siren string `json:"siren"`
}

func search(c *gin.Context) {
	session, err := mgo.Dial(dial)
	if err != nil {
		c.JSON(500, err.Error())
	}
	db := session.DB(db)
	var bilans []interface{}

	siren := c.Params.ByName("siren")

	db.C(collection).Find(bson.M{"_id.siren": siren}).All(&bilans)

	c.JSON(200, bilans)
}

func getPostes() []string {
	var dbSchema = make(map[string]struct{})
	for codePoste := range gorncs.Kb {
		for codeBilan := range gorncs.Kb[codePoste] {
			key := gorncs.Key{CodeBilan: codeBilan, CodePoste: codePoste}
			schema, _ := gorncs.GetSchema(key)
			if schema[0] != "" {
				dbSchema[schema[0]] = struct{}{}
			}
			if schema[1] != "" {
				dbSchema[schema[1]] = struct{}{}
			}
			if schema[2] != "" {
				dbSchema[schema[2]] = struct{}{}
			}
			if schema[3] != "" {
				dbSchema[schema[3]] = struct{}{}
			}
		}
	}
	var postes []string
	for k := range dbSchema {
		postes = append(postes, k)
	}
	sort.Slice(postes, func(a int, b int) bool { return postes[a] < postes[b] })
	return postes
}

func initDB() {
	createTableQuery := `create table bilan (
		id integer primary key,
		siren text,
		date_cloture_exercice datetime,
		code_greffe text,
		num_depot text,
		num_gestion text,
		code_activite text,
		date_cloture_exercice_precedent datetime,
		duree_exercice text,
		duree_exercice_precedent text,
		date_depot datetime,
		code_motif text,
		code_type_bilan text,
		code_devise text,
		code_origine_devise text,
		code_confidentialite text,
		denomination text,
		adresse text,
		rapport_integration text`
	for _, p := range getPostes() {
		createTableQuery = createTableQuery + `,
		` + p + ` integer`
	}
	createTableQuery = createTableQuery + ");"
	database, err := sql.Open("sqlite3", db)
	if err != nil {
		panic(err)
	}
	_, err = database.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Erreur lors de la création de la table: " + err.Error())
	} else {
		fmt.Println("Table bilan créée dans la base " + db)
	}
}

var insert sql.Stmt

func insertBilan(database *sql.DB, bilan gorncs.Bilan) error {
	requete := `insert into bilan (
		siren, date_cloture_exercice, code_greffe, num_depot, num_gestion, code_activite, date_cloture_exercice_precedent,
		duree_exercice, duree_exercice_precedent,	date_depot, code_motif, code_type_bilan, code_devise, code_origine_devise,
		code_confidentialite, denomination, adresse, rapport_integration
		`
	values := `) values (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18`

	// Siren                        string          `json:"siren" bson:"siren"`
	// DateClotureExercice          time.Time       `json:"dateClotureExercice" bson:"dateClotureExercice"`
	// CodeGreffe                   string          `json:"codeGreffe" bson:"codeGreffe"`
	// NumDepot                     string          `json:"numDepot" bson:"numDepot"`
	// NumGestion                   string          `json:"numGestion" bson:"numGestion"`
	// CodeActivite                 string          `json:"codeActivite" bson:"codeActivite"`
	// DateClotureExercicePrecedent time.Time       `json:"dateClotureExercicePrecedent" bson:"dateClotureExercicePrecedent"`
	// DureeExercice                string          `json:"dureeExercice" bson:"dureeExercice"`
	// DureeExercicePrecedent       string          `json:"dureeExercicePrecedent" bson:"dureeExercicePrecedent"`
	// DateDepot                    time.Time       `json:"dateDepot" bson:"dateDepot"`
	// CodeMotif                    string          `json:"codeMotif" bson:"codeMotif"`
	// CodeTypeBilan                string          `json:"codeTypeBilan" bson:"codeTypeBilan"`
	// CodeDevise                   string          `json:"codeDevise" bson:"codeDevise"`
	// CodeOrigineDevise            string          `json:"codeOrigineDevise" bson:"codeOrigineDevise"`
	// CodeConfidentialite          string          `json:"codeConfidentialite" bson:"codeConfidentialite"`
	// Denomination                 string          `json:"denomination" bson:"denomination"`
	// Adresse                      string          `json:"adresse" bson:"adresse"`
	// XMLSource                    string          `json:"XMLSource" bson:"XMLSource"`
	if bilan.Siren == "" {
		fmt.Println("Siren non renseigné, xml source: " + bilan.XMLSource)
		return nil
	}

	if len(bilan.Lignes) > 0 {
		for k, v := range bilan.Lignes {
			requete = requete + ", " + k
			values = values + ", " + strconv.Itoa(*v)
		}

		requete = requete[:len(requete)]
		rapportIntegration, _ := json.Marshal(bilan.Report)
		values = values[:len(values)] + ");"
		_, err := database.Exec(requete+values, bilan.Siren, bilan.DateClotureExercice, bilan.CodeGreffe, bilan.NumDepot,
			bilan.NumGestion, bilan.CodeActivite, bilan.DateClotureExercicePrecedent, bilan.DureeExercice, bilan.DureeExercicePrecedent,
			bilan.DateDepot, bilan.CodeMotif, bilan.CodeTypeBilan, bilan.CodeDevise, bilan.CodeOrigineDevise, bilan.CodeConfidentialite,
			bilan.Denomination, bilan.Adresse, string(rapportIntegration))
		return err
	}
	return nil
}

func scan() {
	database, err := sql.Open("sqlite3", db)

	if err != nil {
		panic(err)
	}
	for bilan := range gorncs.BilanWorker(path) {
		err := insertBilan(database, bilan)
		if err != nil {
			panic(err)
		}
	}
}
