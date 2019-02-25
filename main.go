package main

import (
	"database/sql"
	"flag"
	"fmt"

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
	for _, p := range gorncs.Postes {
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

func scan() {
	database, err := sql.Open("sqlite3", db)
	if err != nil {
		panic(err)
	}
	_, err = database.Exec("PRAGMA journal_mode = OFF")
	_, err = database.Exec("PRAGMA synchronous = OFF")

	fmt.Println(err)
	queryString := gorncs.GetQueryString()

	tx, _ := database.Begin()
	stmt, err := tx.Prepare(queryString)
	if err != nil {
		panic(err)
	}

	for bilan := range gorncs.BilanWorker(path) {
		stmt.Exec(bilan.ToQueryParams()...)
	}

	stmt.Close()
	tx.Commit()
	database.Close()
}
