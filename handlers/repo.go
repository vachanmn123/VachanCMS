package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

		if _, ok := err.(*services.FileNotFoundError); ok {
			c.JSON(404, gin.H{"error": "Config file not found"})
			return
		}

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

	// Create config content
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

	commitMsg := "Initialize VachanCMS Repository"

	// Check if repo is empty
	isEmpty, defaultBranch, err := services.IsRepoEmpty(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to check repository status"})
		return
	}

	var targetBranch string
	if isEmpty {
		// For empty repos, create files on default branch
		targetBranch = defaultBranch
	} else {
		// For non-empty repos, use branch workflow
		targetBranch = uuid.New().String()
		err = services.CreateBranch(access_token, owner, repo, targetBranch)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create a new branch"})
			return
		}
	}

	// Create files
	err = services.CreateOrUpdateFile(access_token, owner, repo, "config/config.json", commitMsg, string(fileContent), targetBranch)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create config file"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, "content/.gitkeep", commitMsg, "", targetBranch)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create data directory"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, "media/.gitkeep", commitMsg, "", targetBranch)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create media directory"})
		return
	}

	// Only merge if we created a separate branch
	if !isEmpty {
		err = services.MergeBranch(access_token, owner, repo, targetBranch, "Initialize repository for VachanCMS")
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to merge branch"})
			return
		}
	}

	c.JSON(200, gin.H{"message": "Repository initialized successfully"})
}
