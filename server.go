package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strings"
)

type data struct {
	worldData map[int][][]int
	quickData map[int][]coords
}

func getSeeds() []string {
	files, err := os.ReadDir("./seeds")
	if err != nil {
		log.Fatal(err)
	}
	var seeds []string
	for _, f := range files {
		seed := strings.Split(f.Name(), ".")[0]
		seeds = append(seeds, seed)
	}
	return seeds
}

func main() {
	r := gin.Default()
	r.Use(CORSMiddleware())

	r.GET("/api_seed", func(c *gin.Context) {
		seeds := getSeeds()
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"seeds":  seeds,
		})
	})

	r.POST("/api_rle", func(c *gin.Context) {
		var requestData map[string]interface{}
		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		seed := requestData["seed"].(string)

		data := rleDecoder(seed)
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   data.initialiseData,
			"sizeX":  data.sizeX,
			"sizeY":  data.sizeY,
			"info":   data.info,
		})
	})

	r.POST("/api_gol", func(c *gin.Context) {
		var requestData map[string]interface{}
		if err := c.BindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		turnStart := int(requestData["turn"].(float64))
		worldInterface := requestData["world"].([]interface{})
		world := make([][]int, len(worldInterface))
		for i, rowInterface := range worldInterface {
			row := rowInterface.([]interface{})
			world[i] = make([]int, len(row))
			for j, cellInterface := range row {
				cell := int(cellInterface.(float64)) // Assuming the cells are integers
				world[i][j] = cell
			}
		}
		turns := 10
		data := GOLRunner(world, turnStart, turns)
		worldData := data.worldData
		quickData := data.quickData

		// Marshal the map to JSON
		jsonWorldData, err := json.Marshal(worldData)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}
		jsonQuickData, err := json.Marshal(quickData)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"data":      string(jsonWorldData),
			"quickData": string(jsonQuickData),
			"turns":     turns + turnStart,
		})
	})

	//r.Run(":5000")
	err := http.ListenAndServeTLS(":8443", "localhost+1.pem", "localhost+1-key.pem", r)
	if err != nil {
		panic(err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
