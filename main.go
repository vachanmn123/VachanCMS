package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/config"
	"github.com/vachanmn123/vachancms/routes"
)

func main() {
	cfg := config.Load()

	router := gin.Default()
	routes.SetupRoutes(router)

	log.Printf("Server starting on port %s", cfg.Port)
	router.Run(":" + cfg.Port)
}
