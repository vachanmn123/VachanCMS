package services

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/go-github/v62/github"
)

var (
	// httpClient is shared across all calls to maintain TCP/TLS connections
	httpClient *http.Client
	once       sync.Once
)

// getClient returns a github.Client using the global pooled http.Client
func getClient(token string) *github.Client {
	once.Do(func() {
		// 1. Create the base pooled transport
		baseTransport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		}

		httpClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: baseTransport,
		}
	})
	return github.NewClient(httpClient).WithAuthToken(token)
}

func ListRepos(token string) ([]*github.Repository, error) {
	ctx := context.Background()
	gh_client := getClient(token)
	repos, res, err := gh_client.Repositories.ListByAuthenticatedUser(ctx, &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		Sort:       "pushed_at",
	})

	if err != nil || res.StatusCode != 200 {
		return nil, err
	}

	return repos, nil
}

type FileNotFoundError struct{}

func (e *FileNotFoundError) Error() string {
	return "file not found"
}

func GetFileContents(token, user, repo, path string, branch ...string) (string, error) {
	ctx := context.Background()
	gh_client := getClient(token)
	var refString string
	if len(branch) == 0 {
		refString = ""
	} else {
		refString = branch[0]
	}

	fileContent, _, res, err := gh_client.Repositories.GetContents(ctx, user, repo, path, &github.RepositoryContentGetOptions{
		Ref: refString,
	})
	if err != nil || res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return "", &FileNotFoundError{}
		}
		return "", err
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", err
	}

	return content, nil
}

func CreateOrUpdateFile(token, user, repo, path, message, content string, branch ...string) error {
	ctx := context.Background()
	gh_client := getClient(token)
	var refString string
	if len(branch) == 0 {
		refString = ""
	} else {
		refString = branch[0]
	}

	fileContent, _, _, err := gh_client.Repositories.GetContents(ctx, user, repo, path, &github.RepositoryContentGetOptions{
		Ref: refString,
	})
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
		Branch:  &refString,
	})
	if err != nil {
		return err
	}
	return nil
}

func UploadFile(token, user, repo, path, message string, content []byte, branch ...string) error {
	ctx := context.Background()
	gh_client := getClient(token)
	var refString string
	if len(branch) == 0 {
		refString = ""
	} else {
		refString = branch[0]
	}

	fileContent, _, _, err := gh_client.Repositories.GetContents(ctx, user, repo, path, &github.RepositoryContentGetOptions{
		Ref: refString,
	})
	var sha *string
	if err == nil {
		s := fileContent.GetSHA()
		sha = &s
	}
	// For both create and update, use CreateFile
	_, _, err = gh_client.Repositories.CreateFile(ctx, user, repo, path, &github.RepositoryContentFileOptions{
		Message: &message,
		Content: content,
		SHA:     sha,
		Branch:  &refString,
	})
	if err != nil {
		return err
	}
	return nil
}

type PageConfig struct {
	Initialized bool
	URL         string
}

func GetPagesConfig(token, user, repo string) (*PageConfig, error) {
	ctx := context.Background()
	gh_client := getClient(token)

	pagesConfig, res, err := gh_client.Repositories.GetPagesInfo(ctx, user, repo)
	if err != nil || res.StatusCode != 200 {
		if res.StatusCode == 404 {
			return &PageConfig{
				Initialized: false,
			}, nil
		}
		return nil, err
	}

	return &PageConfig{
		Initialized: true,
		URL:         pagesConfig.GetHTMLURL(),
	}, nil
}

func CreateBranch(token, user, repo, newBranch string, srcBranch ...string) error {
	ctx := context.Background()
	gh_client := getClient(token)

	gh_repo, _, err := gh_client.Repositories.Get(ctx, user, repo)
	if err != nil {
		return err
	}

	var srcRef string
	if len(srcBranch) == 0 || srcBranch[0] == "" {
		// Use "HEAD" as an alias for the default branch to save an API call
		srcRef = "refs/heads/" + *gh_repo.DefaultBranch
	} else {
		srcRef = "refs/heads/" + srcBranch[0]
	}

	// 1. Get the SHA from the source
	ref, _, err := gh_client.Git.GetRef(ctx, user, repo, srcRef)
	if err != nil {
		return err
	}

	// 2. Create the new reference
	// MUST be "refs/heads/"
	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + newBranch),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}

	_, _, err = gh_client.Git.CreateRef(ctx, user, repo, newRef)
	return err
}

func MergeBranch(token, user, repo, fromBranch, message string, toBranch ...string) error {
	ctx := context.Background()
	gh_client := getClient(token)
	var baseBranchName string
	if len(toBranch) == 0 {
		gh_repo, _, err := gh_client.Repositories.Get(ctx, user, repo)
		if err != nil {
			return err
		}
		baseBranchName = *gh_repo.DefaultBranch
	} else {
		baseBranchName = toBranch[0]
	}

	mergeRequest := &github.RepositoryMergeRequest{
		Base:          github.String(baseBranchName), // The name of the branch you want to merge into
		Head:          github.String(fromBranch),     // The name of the branch you want to merge from
		CommitMessage: github.String(message),        // Optional commit message
	}

	_, _, err := gh_client.Repositories.Merge(ctx, user, repo, mergeRequest)
	if err != nil {
		return err
	}

	_, err = gh_client.Git.DeleteRef(ctx, user, repo, "refs/heads/"+fromBranch)
	return err
}

func IsRepoEmpty(token, user, repo string) (bool, string, error) {
	ctx := context.Background()
	gh_client := getClient(token)

	// Get repository details to find default branch
	gh_repo, _, err := gh_client.Repositories.Get(ctx, user, repo)
	if err != nil {
		return false, "", err
	}

	defaultBranch := *gh_repo.DefaultBranch

	// Try to get commits on the default branch
	// If the repo is empty, this will fail because there are no commits
	_, _, err = gh_client.Repositories.ListCommits(ctx, user, repo, &github.CommitsListOptions{
		SHA: defaultBranch,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})

	if err != nil {
		// If we get an error, the repo is likely empty (no commits)
		return true, defaultBranch, nil
	}

	// If we successfully got commits, the repo is not empty
	return false, defaultBranch, nil
}
