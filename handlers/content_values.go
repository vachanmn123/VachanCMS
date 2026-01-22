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

	c.JSON(200, gin.H{
		"page":        values.Page,
		"items":       values.Items,
		"total_pages": config.TotalPages,
		"total_items": config.TotalItems,
	})
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

	contentType := services.GetContentTypeFromConfig(configFile, ctSlug)
	if contentType == nil {
		c.JSON(400, gin.H{"error": "Content type not found"})
		return
	}

	// Validate newValue fields
	if err := validateContentValueFields(c, &newValue, contentType, access_token, owner, repo); err != nil {
		return // Error response already sent by validateContentValueFields
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

	// Fetch config
	config, err := services.GetContentValueConfig(access_token, owner, repo, ctSlug, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	// Migrate to Order if needed (for existing content types without Order)
	if err := services.MigrateConfigToOrder(access_token, owner, repo, ctSlug, newBranchName, config); err != nil {
		c.JSON(500, gin.H{"error": "Failed to migrate config"})
		return
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

	// Add to Order based on AddTo setting
	addTo := contentType.AddTo
	if addTo == "" {
		addTo = "bottom" // Default
	}

	if addTo == "top" {
		// Prepend to Order
		config.Order = append([]string{newValue.Id}, config.Order...)
	} else {
		// Append to Order (bottom)
		config.Order = append(config.Order, newValue.Id)
	}

	// Regenerate indexes
	if addTo == "top" {
		// If adding to top, regenerate from page 1
		err = services.RegenerateIndexes(access_token, owner, repo, ctSlug, newBranchName, config)
	} else {
		// If adding to bottom, only regenerate from the last page
		lastPage := config.TotalPages
		if lastPage < 1 {
			lastPage = 1
		}
		err = services.RegenerateIndexesFromPage(access_token, owner, repo, ctSlug, newBranchName, config, lastPage)
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to regenerate indexes"})
		return
	}

	config.TotalPages = (len(config.Order)-1)/config.ItemsPerPage + 1
	config.TotalItems = len(config.Order)

	// Save config
	err = services.SaveContentValueConfig(access_token, owner, repo, ctSlug, newBranchName, config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save config"})
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

	contentType := services.GetContentTypeFromConfig(configFile, ctSlug)
	if contentType == nil {
		c.JSON(400, gin.H{"error": "Content type not found"})
		return
	}

	// Validate fields
	if err := validateContentValueFields(c, &updatedValue, contentType, access_token, owner, repo); err != nil {
		return
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

	config, err := services.GetContentValueConfig(access_token, owner, repo, ctSlug, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	// Migrate if needed
	if err := services.MigrateConfigToOrder(access_token, owner, repo, ctSlug, newBranchName, config); err != nil {
		c.JSON(500, gin.H{"error": "Failed to migrate config"})
		return
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

	config.TotalPages = (len(config.Order)-1)/config.ItemsPerPage + 1
	config.TotalItems = len(config.Order)

	// Update config if slugs changed
	if configChanged {
		err = services.SaveContentValueConfig(access_token, owner, repo, ctSlug, newBranchName, config)
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

func DeleteValueById(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
	id := c.Param("id")
	access_token := c.GetString("user_access_token")

	// Create branch for changes
	newBranchName := uuid.New().String()
	err := services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch"})
		return
	}

	// Fetch config
	config, err := services.GetContentValueConfig(access_token, owner, repo, ctSlug, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	// Migrate if needed
	if err := services.MigrateConfigToOrder(access_token, owner, repo, ctSlug, newBranchName, config); err != nil {
		c.JSON(500, gin.H{"error": "Failed to migrate config"})
		return
	}

	// Check if ID exists in Order
	idIndex := -1
	for i, orderId := range config.Order {
		if orderId == id {
			idIndex = i
			break
		}
	}
	if idIndex == -1 {
		c.JSON(404, gin.H{"error": "Content value not found"})
		return
	}

	// Find the page where this item is located
	affectedPage := config.Items[id]

	// Delete the main id.json file
	err = services.DeleteFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, id), fmt.Sprintf("Delete content value %s from %s", id, ctSlug), newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete content value file"})
		return
	}

	// Delete slug file if exists
	for slug, valueId := range config.Slugs {
		if valueId == id {
			err = services.DeleteFile(access_token, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, slug), fmt.Sprintf("Delete slug file for content value %s", id), newBranchName)
			if err != nil {
				// Log but continue - slug file may already be deleted
				fmt.Println("[WARN] Failed to delete slug file:", err)
			}
			delete(config.Slugs, slug)
			break
		}
	}

	// Remove from Order
	config.Order = append(config.Order[:idIndex], config.Order[idIndex+1:]...)

	// Remove from Items map
	delete(config.Items, id)

	// Regenerate indexes from affected page onward
	err = services.RegenerateIndexesFromPage(access_token, owner, repo, ctSlug, newBranchName, config, affectedPage)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to regenerate indexes"})
		return
	}

	config.TotalPages = (len(config.Order)-1)/config.ItemsPerPage + 1
	config.TotalItems = len(config.Order)

	// Save config
	err = services.SaveContentValueConfig(access_token, owner, repo, ctSlug, newBranchName, config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save config"})
		return
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Deleted content value - %s/%s", ctSlug, id))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(200, gin.H{"message": "Content value deleted successfully"})
}

// ReorderValueRequest is the request body for reordering a content value
type ReorderValueRequest struct {
	Position int `json:"position" binding:"required"`
}

func ReorderValueById(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctSlug := c.Param("ctSlug")
	id := c.Param("id")
	access_token := c.GetString("user_access_token")

	var req ReorderValueRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body. Position is required."})
		return
	}

	// Create branch for changes
	newBranchName := uuid.New().String()
	err := services.CreateBranch(access_token, owner, repo, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new branch"})
		return
	}

	// Fetch config
	config, err := services.GetContentValueConfig(access_token, owner, repo, ctSlug, newBranchName)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch content values config"})
		return
	}

	// Migrate if needed
	if err := services.MigrateConfigToOrder(access_token, owner, repo, ctSlug, newBranchName, config); err != nil {
		c.JSON(500, gin.H{"error": "Failed to migrate config"})
		return
	}

	// Find current position of the item (1-based)
	currentIndex := -1
	for i, orderId := range config.Order {
		if orderId == id {
			currentIndex = i
			break
		}
	}
	if currentIndex == -1 {
		c.JSON(404, gin.H{"error": "Content value not found"})
		return
	}

	totalItems := len(config.Order)

	// Validate position (1-based)
	if req.Position < 1 || req.Position > totalItems {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Position must be between 1 and %d", totalItems)})
		return
	}

	// Convert to 0-based index
	newIndex := req.Position - 1

	// If no change needed
	if newIndex == currentIndex {
		c.JSON(200, gin.H{"message": "No change needed", "position": req.Position})
		return
	}

	// Calculate affected pages for partial regeneration
	currentPage := (currentIndex / config.ItemsPerPage) + 1
	newPage := (newIndex / config.ItemsPerPage) + 1
	affectedFromPage := currentPage
	if newPage < affectedFromPage {
		affectedFromPage = newPage
	}

	// Remove from current position
	config.Order = append(config.Order[:currentIndex], config.Order[currentIndex+1:]...)

	// Insert at new position
	// After removal, if newIndex >= currentIndex, we need to adjust
	if newIndex > currentIndex {
		newIndex-- // Adjust for the removal
	}
	if newIndex >= len(config.Order) {
		config.Order = append(config.Order, id)
	} else {
		config.Order = append(config.Order[:newIndex], append([]string{id}, config.Order[newIndex:]...)...)
	}

	// Regenerate indexes from the earliest affected page
	err = services.RegenerateIndexesFromPage(access_token, owner, repo, ctSlug, newBranchName, config, affectedFromPage)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to regenerate indexes"})
		return
	}

	config.TotalPages = (len(config.Order)-1)/config.ItemsPerPage + 1
	config.TotalItems = len(config.Order)

	// Save config
	err = services.SaveContentValueConfig(access_token, owner, repo, ctSlug, newBranchName, config)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to save config"})
		return
	}

	err = services.MergeBranch(access_token, owner, repo, newBranchName, fmt.Sprintf("Reordered content value %s to position %d in %s", id, req.Position, ctSlug))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to merge branch"})
		return
	}

	c.JSON(200, gin.H{"message": "Content value reordered successfully", "position": req.Position})
}

// validateContentValueFields validates the fields of a content value against its content type definition
func validateContentValueFields(c *gin.Context, value *models.ContentValue, contentType *models.ContentType, accessToken, owner, repo string) error {
	for key, fieldValue := range value.Value {
		fieldIndex := slices.IndexFunc(contentType.Fields, func(f models.ContentTypeField) bool {
			return f.FieldName == key
		})
		if fieldIndex == -1 {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s is not defined in content type", key)})
			return fmt.Errorf("field not defined")
		}
		fieldDef := &contentType.Fields[fieldIndex]

		switch fieldDef.FieldType {
		case "text", "textarea":
			_, ok := fieldValue.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return fmt.Errorf("invalid field type")
			}
		case "number":
			_, ok := fieldValue.(float64)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a number", key)})
				return fmt.Errorf("invalid field type")
			}
		case "boolean":
			_, ok := fieldValue.(bool)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a boolean", key)})
				return fmt.Errorf("invalid field type")
			}
		case "select":
			strVal, ok := fieldValue.(string)
			if !ok {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string", key)})
				return fmt.Errorf("invalid field type")
			}
			validOption := slices.Contains(fieldDef.Options, strVal)
			if !validOption {
				c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s has invalid option %s", key, strVal)})
				return fmt.Errorf("invalid option")
			}
		case "media":
			isMultiple := slices.Contains(fieldDef.Options, "multiple")
			var mediaIds []string

			if isMultiple {
				arr, ok := fieldValue.([]interface{})
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be an array of media IDs", key)})
					return fmt.Errorf("invalid field type")
				}
				for _, item := range arr {
					strVal, ok := item.(string)
					if !ok {
						c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should contain string media IDs", key)})
						return fmt.Errorf("invalid field type")
					}
					mediaIds = append(mediaIds, strVal)
				}
			} else {
				strVal, ok := fieldValue.(string)
				if !ok {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Field %s should be a string media ID", key)})
					return fmt.Errorf("invalid field type")
				}
				if strVal != "" {
					mediaIds = []string{strVal}
				}
			}

			if len(mediaIds) > 0 {
				invalidIds, err := services.ValidateMediaIds(accessToken, owner, repo, mediaIds)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to validate media references"})
					return fmt.Errorf("validation error")
				}
				if len(invalidIds) > 0 {
					c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid media ID(s) for field %s: %v", key, invalidIds)})
					return fmt.Errorf("invalid media ids")
				}
			}
		default:
			c.JSON(400, gin.H{"error": fmt.Sprintf("Unsupported field type %s for field %s", fieldDef.FieldType, key)})
			return fmt.Errorf("unsupported field type")
		}
	}
	return nil
}
