package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/MegalLink/my/service"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

const scraperServiceKey = "scraperServiceKey"

// ApiMiddleware will add the db connection to the context
func ApiMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		collector := colly.NewCollector(
			colly.AllowedDomains("es.besoccer.com", "www.es.besoccer.com"),
		)
		scraperService := service.NewScraperService(collector)
		c.Set(scraperServiceKey, scraperService)
		c.Next()
	}
}

func getTeamInformation(c *gin.Context) {
	url := c.Query("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "'url' QUERY PARAMETER CANT NOT BE EMPTY, make sure your request looks like this localhost:8080/team?url=https://es.besoccer.com/equipo/real-madrid",
		})
		return
	}
	scraperService, ok := c.MustGet(scraperServiceKey).(service.ScraperService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "INTERNAL SERVICE ERROR",
		})
		return
	}

	response := scraperService.ScrapTeamInformation(url)
	c.JSON(http.StatusOK, response)
}

func main() {
	r := gin.Default()
	r.Use(ApiMiddleware())
	r.GET("/team", getTeamInformation)
	r.Run()
}

func writeJSON(data interface{}) []byte {
	byteData, err := json.Marshal(data)
	if err != nil {
		log.Println("Unable to create json file")
		return nil
	}

	return byteData
}
