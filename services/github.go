package services

import (
	"context"

	"github.com/google/go-github/v62/github"
)

func ListRepos(token string) ([]*github.Repository, error) {
	ctx := context.Background()
	gh_client := github.NewTokenClient(ctx, token)
	repos, res, err := gh_client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		Sort:       "pushed_at",
	})

	if err != nil || res.StatusCode != 200 {
		return nil, err
	}

	return repos, nil
}

func GetFileContents(token, user, repo, path string) (string, error) {
	ctx := context.Background()
	gh_client := github.NewTokenClient(ctx, token)

	fileContent, _, res, err := gh_client.Repositories.GetContents(ctx, user, repo, path, &github.RepositoryContentGetOptions{})
	if err != nil || res.StatusCode != 200 {
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}

func CreateOrUpdateFile(token, user, repo, path, message, content string) error {
	ctx := context.Background()
	gh_client := github.NewTokenClient(ctx, token)

	fileContent, _, _, err := gh_client.Repositories.GetContents(ctx, user, repo, path, nil)
	var sha *string
	if err == nil {
		s := fileContent.GetSHA()
		sha = &s
	}
	// For both create and update, use CreateFile
	_, _, err = gh_client.Repositories.CreateFile(ctx, user, repo, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: []byte(content),
		SHA:     sha,
	})
	if err != nil {
		return err
	}
	return nil
}
