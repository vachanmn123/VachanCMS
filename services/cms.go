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

func UpdateRepoConfig(access_token, owner, repo string, config *models.ConfigFile, commitMsgs ...string) error {
	fileContent, err := json.Marshal(config)
	if err != nil {
		return err
	}

	var commitMsg string
	if len(commitMsgs) > 0 {
		commitMsg = commitMsgs[0]
	} else {
		commitMsg = "Update config"
	}

	return CreateOrUpdateFile(access_token, owner, repo, "config/config.json", commitMsg, string(fileContent))
}
