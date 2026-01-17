package services

import (
	"encoding/json"

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
