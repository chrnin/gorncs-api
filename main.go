package main

import (
	"fmt"
	"os"

	"github.com/gin-contrib/cors"

	"github.com/chrnin/gorncs"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/", greetings)
	r.GET("/config", config)
	r.GET("/reindex", reindex)
	r.GET("/getBilan/:siren", getBilan)

	r.Run(":3000") // listen and serve on 0.0.0.0:8080
}

type query struct {
	Siren string `json:"siren"`
	Path  string `json:"path"`
}

func config(c *gin.Context) {
	version := "0.1a"
	dataPath := os.Args[1]
	workingDirectory, _ := os.Getwd()
	response := gin.H{
		"path":              dataPath,
		"version":           version,
		"working directory": workingDirectory,
	}
	c.JSON(200, response)
}

func greetings(c *gin.Context) {
	c.JSON(200, gin.H{
		"greetings": "Votre installation Fonctionne",
	})
}

func reindex(c *gin.Context) {
	err := gorncs.Index(os.Args[1])
	if err != nil {
		c.JSON(500, err)
	} else {
		c.JSON(200, "done")
	}
}
func getBilan(c *gin.Context) {
	siren := c.Params.ByName("siren")
	bilan, err := gorncs.GetBilan("../../inpiRoot", siren)

	if err != nil {
		fmt.Println(err)
		c.JSON(500, err)
	} else {
		c.JSON(200, bilan)
	}
}
