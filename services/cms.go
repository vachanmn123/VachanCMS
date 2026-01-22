package services

import (
	"encoding/json"
	"fmt"

	"github.com/vachanmn123/vachancms/models"
)

func GetRepoConfig(access_token, owner, repo string) (*models.ConfigFile, error) {
	configContent, err := GetFileContents(access_token, owner, repo, "config/config.json")
	if err != nil {
		return nil, err
	}

	var configFile models.ConfigFile
	err = json.Unmarshal([]byte(configContent), &configFile)
	if err != nil {
		return nil, err
	}

	return &configFile, nil
}

// ValidateMediaIds checks if the given media IDs exist in the repo's media config.
// Returns a list of invalid IDs (empty slice if all valid).
func ValidateMediaIds(access_token, owner, repo string, mediaIds []string) ([]string, error) {
	if len(mediaIds) == 0 {
		return []string{}, nil
	}

	configContent, err := GetFileContents(access_token, owner, repo, "media/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch media config: %w", err)
	}

	var config models.MediaConfigFile
	if err := json.Unmarshal([]byte(configContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse media config: %w", err)
	}

	var invalidIds []string
	for _, id := range mediaIds {
		if _, exists := config.Items[id]; !exists {
			invalidIds = append(invalidIds, id)
		}
	}

	return invalidIds, nil
}

// GetContentValueConfig fetches and parses the content value config for a content type
func GetContentValueConfig(accessToken, owner, repo, ctSlug string, branch ...string) (*models.ContentValueConfigFile, error) {
	var configContent string
	var err error

	if len(branch) > 0 && branch[0] != "" {
		configContent, err = GetFileContents(accessToken, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug), branch[0])
	} else {
		configContent, err = GetFileContents(accessToken, owner, repo, fmt.Sprintf("data/%s/config.json", ctSlug))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch content value config: %w", err)
	}

	var config models.ContentValueConfigFile
	if err := json.Unmarshal([]byte(configContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse content value config: %w", err)
	}

	// Initialize maps if nil
	if config.Items == nil {
		config.Items = make(map[string]int)
	}
	if config.Slugs == nil {
		config.Slugs = make(map[string]string)
	}
	if config.Order == nil {
		config.Order = []string{}
	}

	return &config, nil
}

// GetContentValue fetches a single content value by ID
func GetContentValue(accessToken, owner, repo, ctSlug, id string, branch ...string) (*models.ContentValue, error) {
	var content string
	var err error

	if len(branch) > 0 && branch[0] != "" {
		content, err = GetFileContents(accessToken, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, id), branch[0])
	} else {
		content, err = GetFileContents(accessToken, owner, repo, fmt.Sprintf("data/%s/%s.json", ctSlug, id))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch content value: %w", err)
	}

	var value models.ContentValue
	if err := json.Unmarshal([]byte(content), &value); err != nil {
		return nil, fmt.Errorf("failed to parse content value: %w", err)
	}

	return &value, nil
}

// RegenerateIndexes rebuilds all index files from the Order array in config.
// This is the source of truth for content ordering.
// It updates the Items map, TotalItems, TotalPages, and regenerates all index-*.json files.
// It also deletes any extra index files that are no longer needed.
func RegenerateIndexes(accessToken, owner, repo, ctSlug, branch string, config *models.ContentValueConfigFile) error {
	if config.ItemsPerPage <= 0 {
		config.ItemsPerPage = 10 // Default
	}

	// Calculate total pages
	totalItems := len(config.Order)
	totalPages := 1
	if totalItems > 0 {
		totalPages = (totalItems + config.ItemsPerPage - 1) / config.ItemsPerPage
	}

	oldTotalPages := config.TotalPages

	// Reset and rebuild Items map
	config.Items = make(map[string]int)

	// Generate index files page by page
	for page := 1; page <= totalPages; page++ {
		startIdx := (page - 1) * config.ItemsPerPage
		endIdx := startIdx + config.ItemsPerPage
		if endIdx > totalItems {
			endIdx = totalItems
		}

		pageItems := []models.ContentValue{}

		for i := startIdx; i < endIdx; i++ {
			id := config.Order[i]
			config.Items[id] = page

			// Fetch the content value
			value, err := GetContentValue(accessToken, owner, repo, ctSlug, id, branch)
			if err != nil {
				// If item doesn't exist, skip it (it may have been deleted)
				continue
			}
			pageItems = append(pageItems, *value)
		}

		indexFile := models.ContentValueIndexFile{
			Page:  page,
			Items: pageItems,
		}

		indexJson, err := json.Marshal(indexFile)
		if err != nil {
			return fmt.Errorf("failed to marshal index file for page %d: %w", page, err)
		}

		err = CreateOrUpdateFile(accessToken, owner, repo,
			fmt.Sprintf("data/%s/index-%d.json", ctSlug, page),
			fmt.Sprintf("Regenerate index page %d for %s", page, ctSlug),
			string(indexJson), branch)
		if err != nil {
			return fmt.Errorf("failed to update index file for page %d: %w", page, err)
		}
	}

	// Delete extra index files if pages decreased
	for page := totalPages + 1; page <= oldTotalPages; page++ {
		err := DeleteFile(accessToken, owner, repo,
			fmt.Sprintf("data/%s/index-%d.json", ctSlug, page),
			fmt.Sprintf("Remove extra index page %d for %s", page, ctSlug),
			branch)
		if err != nil {
			// Log but don't fail - file might not exist
			continue
		}
	}

	// Update config
	config.TotalItems = totalItems
	config.TotalPages = totalPages

	return nil
}

// RegenerateIndexesFromPage regenerates index files starting from a specific page.
// This is more efficient when you know which page was affected.
// Use this when an item is added/removed/moved and you know the affected page.
func RegenerateIndexesFromPage(accessToken, owner, repo, ctSlug, branch string, config *models.ContentValueConfigFile, fromPage int) error {
	if config.ItemsPerPage <= 0 {
		config.ItemsPerPage = 10
	}

	totalItems := len(config.Order)
	totalPages := 1
	if totalItems > 0 {
		totalPages = (totalItems + config.ItemsPerPage - 1) / config.ItemsPerPage
	}

	oldTotalPages := config.TotalPages

	// Regenerate from fromPage to totalPages
	for page := fromPage; page <= totalPages; page++ {
		startIdx := (page - 1) * config.ItemsPerPage
		endIdx := startIdx + config.ItemsPerPage
		if endIdx > totalItems {
			endIdx = totalItems
		}

		pageItems := []models.ContentValue{}

		for i := startIdx; i < endIdx; i++ {
			id := config.Order[i]
			config.Items[id] = page

			value, err := GetContentValue(accessToken, owner, repo, ctSlug, id, branch)
			if err != nil {
				continue
			}
			pageItems = append(pageItems, *value)
		}

		indexFile := models.ContentValueIndexFile{
			Page:  page,
			Items: pageItems,
		}

		indexJson, err := json.Marshal(indexFile)
		if err != nil {
			return fmt.Errorf("failed to marshal index file for page %d: %w", page, err)
		}

		err = CreateOrUpdateFile(accessToken, owner, repo,
			fmt.Sprintf("data/%s/index-%d.json", ctSlug, page),
			fmt.Sprintf("Regenerate index page %d for %s", page, ctSlug),
			string(indexJson), branch)
		if err != nil {
			return fmt.Errorf("failed to update index file for page %d: %w", page, err)
		}
	}

	// Delete extra index files
	for page := totalPages + 1; page <= oldTotalPages; page++ {
		err := DeleteFile(accessToken, owner, repo,
			fmt.Sprintf("data/%s/index-%d.json", ctSlug, page),
			fmt.Sprintf("Remove extra index page %d for %s", page, ctSlug),
			branch)
		if err != nil {
			continue
		}
	}

	// Rebuild Items map for pages before fromPage (they weren't updated above)
	for page := 1; page < fromPage; page++ {
		startIdx := (page - 1) * config.ItemsPerPage
		endIdx := startIdx + config.ItemsPerPage
		if endIdx > totalItems {
			endIdx = totalItems
		}
		for i := startIdx; i < endIdx; i++ {
			config.Items[config.Order[i]] = page
		}
	}

	config.TotalItems = totalItems
	config.TotalPages = totalPages

	return nil
}

// SaveContentValueConfig saves the content value config to the repo
func SaveContentValueConfig(accessToken, owner, repo, ctSlug, branch string, config *models.ContentValueConfigFile) error {
	configJson, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = CreateOrUpdateFile(accessToken, owner, repo,
		fmt.Sprintf("data/%s/config.json", ctSlug),
		fmt.Sprintf("Update config for %s", ctSlug),
		string(configJson), branch)
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// MigrateConfigToOrder migrates an existing config that doesn't have Order
// by building the Order array from existing index files
func MigrateConfigToOrder(accessToken, owner, repo, ctSlug, branch string, config *models.ContentValueConfigFile) error {
	if len(config.Order) > 0 {
		// Already has order, no migration needed
		return nil
	}

	// Build order from existing index files
	order := []string{}

	for page := 1; page <= config.TotalPages; page++ {
		indexContent, err := GetFileContents(accessToken, owner, repo,
			fmt.Sprintf("data/%s/index-%d.json", ctSlug, page), branch)
		if err != nil {
			continue
		}

		var indexFile models.ContentValueIndexFile
		if err := json.Unmarshal([]byte(indexContent), &indexFile); err != nil {
			continue
		}

		for _, item := range indexFile.Items {
			order = append(order, item.Id)
		}
	}

	config.Order = order
	return nil
}

// GetContentTypeFromConfig finds a content type by slug from the repo config
func GetContentTypeFromConfig(configFile *models.ConfigFile, ctSlug string) *models.ContentType {
	for _, ct := range configFile.ContentTypes {
		if ct.Slug == ctSlug {
			return &ct
		}
	}
	return nil
}
