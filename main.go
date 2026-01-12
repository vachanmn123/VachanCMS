package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/config"
	"github.com/vachanmn123/vachancms/routes"
)

func main() {
	cfg := config.Load()

	router := gin.Default()
	routes.SetupRoutes(router.Group("/api"))

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "API not found"})
		} else {
			// Find the file in frontend/dist
			// If it exists, send it
			// Else, send index.html
			filePath := filepath.Join("./frontend/dist", c.Request.URL.Path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
			} else {
				c.File("./frontend/dist/index.html")
			}
		}
	})

	log.Printf("Server starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
