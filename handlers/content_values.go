package handlers

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vachanmn123/vachancms/models"
	"github.com/vachanmn123/vachancms/services"
)

// slugRegex validates that slug is lowercase alphanumeric with hyphens
var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateSlug(slug string) bool {
	if slug == "" {
		return true // empty slug is valid (optional field)
	}
	return slugRegex.MatchString(slug)
}

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

	// Validate slug format if provided
	if newValue.Slug != "" {
		if !validateSlug(newValue.Slug) {
			c.JSON(400, gin.H{"error": "Invalid slug format. Slug must be lowercase alphanumeric with hyphens (e.g., 'my-blog-post')"})
			return
		}
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
		case "text", "textarea":
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
		case "media":
			isMultiple := slices.Contains(fieldDef.Options, "multiple")
			var mediaIds []string

			if isMultiple {
				arr, ok := value.([]interface{})
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be an array of media IDs", key)})
					return
				}
				for _, item := range arr {
					strVal, ok := item.(string)
					if !ok {
						c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should contain string media IDs", key)})
						return
					}
					mediaIds = append(mediaIds, strVal)
				}
			} else {
				strVal, ok := value.(string)
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string media ID", key)})
					return
				}
				if strVal != "" {
					mediaIds = []string{strVal}
				}
			}

			if len(mediaIds) > 0 {
				invalidIds, err := services.ValidateMediaIds(access_token, owner, repo, mediaIds)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to validate media references"})
					return
				}
				if len(invalidIds) > 0 {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid media ID(s) for field %s: %v", key, invalidIds)})
					return
				}
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

	// Create the main id.json file
	err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, newValue.Id), fmt.Sprintf("Add new content value to %s", ctSlug), string(newValueJson), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new content value file"})
		return
	}

	// Fetch config to check slug uniqueness
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

	// Initialize slugs map if nil
	if config.Slugs == nil {
		config.Slugs = make(map[string]string)
	}

	// Check slug uniqueness if provided
	if newValue.Slug != "" {
		if _, exists := config.Slugs[newValue.Slug]; exists {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Slug '%s' is already in use", newValue.Slug)})
			return
		}

		// Create slug.json file
		err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, newValue.Slug), fmt.Sprintf("Add slug file for content value %s", newValue.Id), string(newValueJson), newBranchName)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create slug file"})
			return
		}

		// Add slug to config
		config.Slugs[newValue.Slug] = newValue.Id
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

	// Validate slug format if provided
	if updatedValue.Slug != "" {
		if !validateSlug(updatedValue.Slug) {
			c.JSON(400, gin.H{"error": "Invalid slug format. Slug must be lowercase alphanumeric with hyphens (e.g., 'my-blog-post')"})
			return
		}
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
		case "text", "textarea":
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
		case "media":
			isMultiple := slices.Contains(fieldDef.Options, "multiple")
			var mediaIds []string

			if isMultiple {
				arr, ok := value.([]interface{})
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be an array of media IDs", key)})
					return
				}
				for _, item := range arr {
					strVal, ok := item.(string)
					if !ok {
						c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should contain string media IDs", key)})
						return
					}
					mediaIds = append(mediaIds, strVal)
				}
			} else {
				strVal, ok := value.(string)
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string media ID", key)})
					return
				}
				if strVal != "" {
					mediaIds = []string{strVal}
				}
			}

			if len(mediaIds) > 0 {
				invalidIds, err := services.ValidateMediaIds(access_token, owner, repo, mediaIds)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to validate media references"})
					return
				}
				if len(invalidIds) > 0 {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid media ID(s) for field %s: %v", key, invalidIds)})
					return
				}
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

	// Update the main id.json file
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

	// Initialize slugs map if nil
	if config.Slugs == nil {
		config.Slugs = make(map[string]string)
	}

	// Find the old slug for this ID (if any)
	var oldSlug string
	for slug, valueId := range config.Slugs {
		if valueId == id {
			oldSlug = slug
			break
		}
	}

	// Handle slug changes
	configChanged := false
	if oldSlug != updatedValue.Slug {
		// Delete old slug file if it existed
		if oldSlug != "" {
			err = services.DeleteFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, oldSlug), fmt.Sprintf("Remove old slug file for content value %s", id), newBranchName)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to delete old slug file"})
				return
			}
			delete(config.Slugs, oldSlug)
			configChanged = true
		}

		// Create new slug file if slug is provided
		if updatedValue.Slug != "" {
			// Check slug uniqueness
			if existingId, exists := config.Slugs[updatedValue.Slug]; exists && existingId != id {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Slug '%s' is already in use", updatedValue.Slug)})
				return
			}

			err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, updatedValue.Slug), fmt.Sprintf("Add slug file for content value %s", id), string(updatedValueJson), newBranchName)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to create slug file"})
				return
			}
			config.Slugs[updatedValue.Slug] = id
			configChanged = true
		}
	} else if updatedValue.Slug != "" {
		// Slug hasn't changed but we still need to update the slug file with new content
		err = services.CreateOrUpdateFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, updatedValue.Slug), fmt.Sprintf("Update slug file for content value %s", id), string(updatedValueJson), newBranchName)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update slug file"})
			return
		}
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

	// Update config if slugs changed
	if configChanged {
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
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Edit content value - %s/%s", ctSlug, id))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(200, updatedValue)
}
