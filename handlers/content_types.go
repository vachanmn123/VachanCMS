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

	// Update the config file
	err = services.UpdateRepoConfig(access_token, owner, repo, configFile, fmt.Sprintf("Create Content Type: %s", contentType.Name))
	if err != nil {
		fmt.Println("Error updating config:", err)
		c.JSON(500, gin.H{"error": "Failed to update config"})
		return
	}

	contentTypeConfigFile := models.ContentValueConfigFile{
		TotalPages:   1,
		TotalItems:   0,
		ItemsPerPage: 10,
		Items:        map[string]int{},
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

	// Create a new folder under the data/ directory for this content type
	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", contentType.Slug), fmt.Sprintf("Create data folder for content type: %s", contentType.Name), string(contentTypeConfigFileJson))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create data folder and config file for content type"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/index-1.json", contentType.Slug), fmt.Sprintf("Create index file for content type: %s", contentType.Name), string(contentValueIndexFileJson))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create index file for content type"})
		return
	}

	c.JSON(201, contentType)
}
