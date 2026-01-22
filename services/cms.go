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
