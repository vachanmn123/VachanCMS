package handlers

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vachanmn123/vachancms/models"
	"github.com/vachanmn123/vachancms/services"
)

func ListRepositoriesHandler(c *gin.Context) {
	access_token := c.GetString("user_access_token")

	repos, err := services.ListRepos(access_token)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch repositories"})
		return
	}

	c.JSON(200, repos)
}

func GetRepoConfig(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	configFile, err := services.GetRepoConfig(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch or parse config"})
		return
	}

	c.JSON(200, *configFile)
}

func InitializeRepo(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	type InitRequest struct {
		SiteName string `json:"site_name"`
	}
	var initReq InitRequest

	err := c.BindJSON(&initReq)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Create a new config/config.json file with the default values
	cfg := models.ConfigFile{
		SiteName:           initReq.SiteName,
		ContentTypes:       []models.ContentType{},
		InitializationDate: time.Now().String(),
	}

	fileContent, err := json.Marshal(cfg)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal config content"})
		return
	}

	commitMsg := "Initalize VachanCMS Repository"

	err = services.CreateOrUpdateFile(access_token, owner, repo, "config/config.json", commitMsg, string(fileContent))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create config file"})
		return
	}

	// Create a new content/.gitkeep file to initialize the content directory
	err = services.CreateOrUpdateFile(access_token, owner, repo, "content/.gitkeep", commitMsg, "")

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create data directory"})
		return
	}

	// Create a new media/.gitkeep file to initialize the media directory
	err = services.CreateOrUpdateFile(access_token, owner, repo, "media/.gitkeep", commitMsg, "")

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create media directory"})
		return
	}

	c.JSON(200, gin.H{"message": "Repository initialized successfully"})
}
