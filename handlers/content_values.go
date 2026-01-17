package handlers

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vachanmn123/vachancms/models"
	"github.com/vachanmn123/vachancms/services"
)

func ListValuesByType(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
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

	configContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	var config models.ContentValueConfigFile
	err = json.Unmarshal([]byte(configContents), &config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse content values config"})
		return
	}

	if page > config.TotalPages {
		c.JSON(400, gin.H{"error": "Page exceeds total pages"})
		return
	}

	valuesIndex, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, page))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values index"})
		return
	}

	var values models.ContentValueIndexFile
	err = json.Unmarshal([]byte(valuesIndex), &values)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse content values index"})
		return
	}

	c.JSON(200, values)
}

func GetValueById(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
	id := c.Param("id")
	access_token := c.GetString("user_access_token")

	valueContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, id))
	if err != nil {
		c.JSON(404, gin.H{"error": "Content value not found"})
		return
	}

	var value models.ContentValue
	err = json.Unmarshal([]byte(valueContents), &value)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse content value"})
		return
	}

	c.JSON(200, value)
}

func CreateValueOfType(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
	access_token := c.GetString("user_access_token")

	var newValue models.ContentValue
	if err := c.BindJSON(&newValue); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	configFile, err := services.GetRepoConfig(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch or parse config"})
		return
	}

	var contentType *models.ContentType
	for _, ct := range configFile.ContentTypes {
		if ct.Slug == ctSlug {
			contentType = &ct
			break
		}
	}
	if contentType == nil {
		c.JSON(400, gin.H{"error": "Content type not found"})
		return
	}

	// Validate newValue fields here
	for key, value := range newValue.Value {
		fieldIndex := slices.IndexFunc(contentType.Fields, func(f models.ContentTypeField) bool {
			if f.FieldName == key {
				return true
			}
			return false
		})
		if fieldIndex == -1 {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s is not defined in content type", key)})
			return
		}
		fieldDef := &contentType.Fields[fieldIndex]

		switch fieldDef.FieldType {
		case "text":
			_, ok := value.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return
			}
		case "number":
			_, ok := value.(float64)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a number", key)})
				return
			}
		case "boolean":
			_, ok := value.(bool)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a boolean", key)})
				return
			}
		case "select":
			strVal, ok := value.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return
			}
			validOption := slices.Contains(fieldDef.Options, strVal)
			if !validOption {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s has invalid option %s", key, strVal)})
				return
			}
		default:
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unsupported field type %s for field %s", fieldDef.FieldType, key)})
			return
		}
	}

	newValue.Id = uuid.New().String()

	newValueJson, err := json.Marshal(newValue)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal new content value"})
		return
	}

	newBranchName := uuid.New().String()
	err = services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch."})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, newValue.Id), fmt.Sprintf("Add new content value to %s", ctSlug), string(newValueJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new content value file"})
		return
	}

	configContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	var config models.ContentValueConfigFile
	err = json.Unmarshal([]byte(configContents), &config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse content values config"})
		return
	}

	targetPage := (config.TotalItems + config.ItemsPerPage) / config.ItemsPerPage
	if targetPage > config.TotalPages {
		// Create new page
		newIndexFile := models.ContentValueIndexFile{
			Page:  targetPage,
			Items: []models.ContentValue{},
		}
		newIndexJson, err := json.Marshal(newIndexFile)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to marshal new index file"})
			return
		}
		err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, targetPage), fmt.Sprintf("Create new index page %d for %s", targetPage, ctSlug), string(newIndexJson), newBranchName)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create new index file"})
			return
		}
		config.TotalPages = targetPage
	}

	// Read and update the target page's index
	indexContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, targetPage), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch target index file"})
		return
	}

	var indexFile models.ContentValueIndexFile
	err = json.Unmarshal([]byte(indexContents), &indexFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse target index file"})
		return
	}

	indexFile.Items = append(indexFile.Items, newValue)
	updatedIndexJson, err := json.Marshal(indexFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal updated index file"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, targetPage), fmt.Sprintf("Update index page %d for %s", targetPage, ctSlug), string(updatedIndexJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update index file"})
		return
	}

	// Update config
	config.TotalItems++
	config.Items[newValue.Id] = targetPage
	updatedConfigJson, err := json.Marshal(config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal updated config"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug), fmt.Sprintf("Update config for %s", ctSlug), string(updatedConfigJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update config"})
		return
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Added new content value - %s/%s", ctSlug, newValue.Id))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(201, newValue)
}

func UpdateValueById(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
	id := c.Param("id")
	access_token := c.GetString("user_access_token")

	var updatedValue models.ContentValue
	if err := c.BindJSON(&updatedValue); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	updatedValue.Id = id

	configFile, err := services.GetRepoConfig(access_token, owner, repo)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch or parse config"})
		return
	}

	var contentType *models.ContentType
	for _, ct := range configFile.ContentTypes {
		if ct.Slug == ctSlug {
			contentType = &ct
			break
		}
	}
	if contentType == nil {
		c.JSON(400, gin.H{"error": "Content type not found"})
		return
	}

	// Validate newValue fields here
	for key, value := range updatedValue.Value {
		fieldIndex := slices.IndexFunc(contentType.Fields, func(f models.ContentTypeField) bool {
			if f.FieldName == key {
				return true
			}
			return false
		})
		if fieldIndex == -1 {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s is not defined in content type", key)})
			return
		}
		fieldDef := &contentType.Fields[fieldIndex]

		switch fieldDef.FieldType {
		case "text":
			_, ok := value.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return
			}
		case "number":
			_, ok := value.(float64)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a number", key)})
				return
			}
		case "boolean":
			_, ok := value.(bool)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a boolean", key)})
				return
			}
		case "select":
			strVal, ok := value.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return
			}
			validOption := slices.Contains(fieldDef.Options, strVal)
			if !validOption {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s has invalid option %s", key, strVal)})
				return
			}
		default:
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unsupported field type %s for field %s", fieldDef.FieldType, key)})
			return
		}
	}

	updatedValueJson, err := json.Marshal(updatedValue)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal updated content value"})
		return
	}

	newBranchName := uuid.New().String()
	err = services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch."})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, id), fmt.Sprintf("Update content value %s in %s", id, ctSlug), string(updatedValueJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update content value file"})
		return
	}

	configContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	var config models.ContentValueConfigFile
	err = json.Unmarshal([]byte(configContents), &config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse content values config"})
		return
	}

	page, exists := config.Items[id]
	if !exists {
		c.JSON(400, gin.H{"error": "Content value ID not found in config"})
		return
	}

	indexContents, err := services.GetFileContents(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, page), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch index file"})
		return
	}

	var indexFile models.ContentValueIndexFile
	err = json.Unmarshal([]byte(indexContents), &indexFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse index file"})
		return
	}

	found := false
	for i, item := range indexFile.Items {
		if item.Id == id {
			indexFile.Items[i] = updatedValue
			found = true
			break
		}
	}
	if !found {
		indexFile.Items = append(indexFile.Items, updatedValue)
	}

	updatedIndexJson, err := json.Marshal(indexFile)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to marshal updated index file"})
		return
	}

	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/index-%d.json", ctSlug, page), fmt.Sprintf("Update index page %d for %s", page, ctSlug), string(updatedIndexJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update index file"})
		return
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Edit content value - %s/%s", ctSlug, id))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(200, updatedValue)
}
