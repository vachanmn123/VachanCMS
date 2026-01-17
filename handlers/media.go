package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vachanmn123/vachancms/models"
	"github.com/vachanmn123/vachancms/services"
)

func ListMedia(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err != nil || parsedPage < 1 {
			c.JSON(400, gin.H{"error": "Invalid page parameter"})
			return
		} else {
			page = parsedPage
		}
	}

	configContents, err := services.GetFileContents(access_token, owner, repo, "media/config.json")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch media config"})
		return
	}

	var config models.MediaConfigFile
	err = json.Unmarshal([]byte(configContents), &config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse media config"})
		return
	}

	if page > config.TotalPages {
		c.JSON(400, gin.H{"error": "Page exceeds total pages"})
		return
	}

	indexContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("media/index-%d.json", page))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch media index"})
		return
	}

	var index models.MediaIndexFile
	err = json.Unmarshal([]byte(indexContents), &index)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse media index"})
		return
	}

	c.JSON(200, index)
}

func UploadMedia(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	access_token := c.GetString("user_access_token")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to get file from request"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read file content"})
		return
	}

	id := uuid.New().String()
	fileName := header.Filename
	fileType := header.Header.Get("Content-Type")
	if fileType == "" {
		fileType = "application/octet-stream"
	}

	newBranchName := uuid.New().String()
	err = services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch."})
		return
	}

	// Upload the file
	err = services.UploadFile(access_token, owner, repo, fmt.Sprintf("media/%s", id), fmt.Sprintf("Upload media file: %s", fileName), content, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to upload file"})
		return
	}

	mediaFile := models.MediaFile{
		Id:       id,
		FileName: fileName,
		FileType: fileType,
	}

	// Handle config and index
	configContents, err := services.GetFileContents(access_token, owner, repo, "media/config.json", newBranchName)
	var config models.MediaConfigFile
	if err != nil {
		// Create initial config
		config = models.MediaConfigFile{
			TotalPages:   1,
			TotalItems:   0,
			ItemsPerPage: 10,
			Items:        map[string]int{},
		}
		configJson, _ := json.Marshal(config)
		services.CreateOrUpdateFile(access_token, owner, repo, "media/config.json", "Initialize media config", string(configJson), newBranchName)

		// Create initial index
		indexFile := models.MediaIndexFile{
			Page:  1,
			Media: []models.MediaFile{},
		}
		indexJson, _ := json.Marshal(indexFile)
		services.CreateOrUpdateFile(access_token, owner, repo, "media/index-1.json", "Initialize media index", string(indexJson), newBranchName)
	} else {
		json.Unmarshal([]byte(configContents), &config)
	}

	targetPage := (config.TotalItems + config.ItemsPerPage) / config.ItemsPerPage
	if targetPage > config.TotalPages {
		// Create new page
		newIndexFile := models.MediaIndexFile{
			Page:  targetPage,
			Media: []models.MediaFile{},
		}
		newIndexJson, _ := json.Marshal(newIndexFile)
		services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("media/index-%d.json", targetPage), fmt.Sprintf("Create media index page %d", targetPage), string(newIndexJson), newBranchName)
		config.TotalPages = targetPage
	}

	// Read and update the target page's index
	indexContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("media/index-%d.json", targetPage), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch target index"})
		return
	}

	var indexFile models.MediaIndexFile
	json.Unmarshal([]byte(indexContents), &indexFile)

	indexFile.Media = append(indexFile.Media, mediaFile)
	updatedIndexJson, _ := json.Marshal(indexFile)
	services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("media/index-%d.json", targetPage), fmt.Sprintf("Update media index page %d", targetPage), string(updatedIndexJson), newBranchName)

	// Update config
	config.TotalItems++
	config.Items[mediaFile.Id] = targetPage
	updatedConfigJson, _ := json.Marshal(config)
	services.CreateOrUpdateFile(access_token, owner, repo, "media/config.json", "Update media config", string(updatedConfigJson), newBranchName)

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Added new media - %s", mediaFile.Id))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(201, mediaFile)
}

func GetMediaById(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	id := c.Param("id")
	access_token := c.GetString("user_access_token")

	// Get file content
	content, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("media/%s", id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Media file not found"})
		return
	}

	// Get metadata from config
	configContents, err := services.GetFileContents(access_token, owner, repo, "media/config.json")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch media config"})
		return
	}

	var config models.MediaConfigFile
	json.Unmarshal([]byte(configContents), &config)

	page, exists := config.Items[id]
	if !exists {
		c.JSON(404, gin.H{"error": "Media file not found"})
		return
	}

	indexContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("media/index-%d.json", page))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch media index"})
		return
	}

	var index models.MediaIndexFile
	json.Unmarshal([]byte(indexContents), &index)

	var mediaFile models.MediaFile
	found := false
	for _, mf := range index.Media {
		if mf.Id == id {
			mediaFile = mf
			found = true
			break
		}
	}
	if !found {
		c.JSON(404, gin.H{"error": "Media file not found"})
		return
	}

	// Return file content with appropriate headers
	c.Header("Content-Type", mediaFile.FileType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", mediaFile.FileName))
	c.Data(200, mediaFile.FileType, []byte(content))
}
