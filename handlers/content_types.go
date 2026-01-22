package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vachanmn123/vachancms/models"
	"github.com/vachanmn123/vachancms/services"
)

func ListContentTypes(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	configFile, err := services.GetRepoConfig(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch or parse config"})
		return
	}

	c.JSON(200, configFile.ContentTypes)
}

func CreateContentType(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	var contentType models.ContentType
	if err := c.BindJSON(&contentType); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	contentType.Id = uuid.New().String()

	// Validate and set defaults for ItemsPerPage
	if contentType.ItemsPerPage <= 0 {
		contentType.ItemsPerPage = 10 // Default
	} else if contentType.ItemsPerPage > 100 {
		c.JSON(400, gin.H{"error": "ItemsPerPage must be between 1 and 100"})
		return
	}

	// Validate and set defaults for AddTo
	if contentType.AddTo == "" {
		contentType.AddTo = "bottom" // Default
	} else if contentType.AddTo != "top" && contentType.AddTo != "bottom" {
		c.JSON(400, gin.H{"error": "AddTo must be 'top' or 'bottom'"})
		return
	}

	configFile, err := services.GetRepoConfig(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch or parse config"})
		return
	}

	// Find if a content type with the same slug already exists
	for _, ct := range configFile.ContentTypes {
		if ct.Slug == contentType.Slug {
			c.JSON(400, gin.H{"error": "Content type with the same slug already exists"})
			return
		}
	}

	configFile.ContentTypes = append(configFile.ContentTypes, contentType)

	fileContent, err := json.Marshal(configFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal config file"})
		return
	}

	newBranchName := uuid.New().String()
	err = services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch."})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, "config/config.json", fmt.Sprintf("Create Content Type: %s", contentType.Name), string(fileContent), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update config"})
		return
	}

	// Initialize the content value config with Order array and ItemsPerPage from content type
	contentTypeConfigFile := models.ContentValueConfigFile{
		TotalPages:   1,
		TotalItems:   0,
		ItemsPerPage: contentType.ItemsPerPage,
		Items:        map[string]int{},
		Slugs:        map[string]string{},
		Order:        []string{},
	}
	contentTypeConfigFileJson, err := json.Marshal(contentTypeConfigFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal content type config file"})
		return
	}

	contentValueIndexFile := models.ContentValueIndexFile{
		Page:  1,
		Items: []models.ContentValue{},
	}
	contentValueIndexFileJson, err := json.Marshal(contentValueIndexFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal content value index file"})
		return
	}

	// Create a new folder under the data/ directory for this content type
	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", contentType.Slug), fmt.Sprintf("Create data folder for content type: %s", contentType.Name), string(contentTypeConfigFileJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create data folder and config file for content type"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/index-1.json", contentType.Slug), fmt.Sprintf("Create index file for content type: %s", contentType.Name), string(contentValueIndexFileJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create index file for content type"})
		return
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Added new content type - %s", contentType.Name))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(201, contentType)
}
