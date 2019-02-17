package main

import (
	"flag"
	"fmt"

	"github.com/chrnin/gorncs"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
)

var dial string
var db string
var collection string
var path string
var scanner bool

func init() {

	flag.StringVar(&dial, "dial", "localhost", "MongoDB dial URL")
	flag.StringVar(&db, "DB", "inpi", "MongoDB database")
	flag.StringVar(&collection, "C", "bilan", "MongoDB collection")
	flag.StringVar(&path, "path", ".", "RNCS root path")
	flag.BoolVar(&scanner, "scanner", false, "Scan and import the root directory")
}

func main() {
	flag.Parse()
	if scanner {
		scan()
	} else {
		r := gin.Default()
		r.Use(cors.Default())
		r.GET("/:siren", search)
		r.Run(":3000") // listen and serve on 0.0.0.0:8080
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

func scan() {
	session, err := mgo.Dial(dial)
	if err != nil {
		panic(err)
	}
	db := session.DB(db)

	var bs []interface{}
	for bilan := range gorncs.BilanWorker(path) {
		_, err := db.C(collection).Upsert(bson.M{"_id": bilan.ID}, bilan)
		if err != nil {
			fmt.Println(err)
		}
	}
	db.C(collection).Insert(bs...)

}
