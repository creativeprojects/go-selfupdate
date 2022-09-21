package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/xanzy/go-gitlab"
)

// GitLabConfig is an object to pass to NewGitLabSource
type GitLabConfig struct {
	// APIToken represents GitLab API token. If it's not empty, it will be used for authentication for the API
	APIToken string
	// BaseURL is a base URL of your private GitLab instance
	BaseURL string
}

// GitLabSource is used to load release information from GitLab
type GitLabSource struct {
	api *gitlab.Client
}

// NewGitLabSource creates a new GitLabSource from a config object.
// It initializes a GitLab API client.
// If you set your API token to the $GITLAB_TOKEN environment variable, the client will use it.
// You can pass an empty GitLabSource{} to use the default configuration
// The function will return an error if the GitLab Enterprise URLs in the config object cannot be parsed
func NewGitLabSource(config GitLabConfig) (*GitLabSource, error) {
	token := config.APIToken
	if token == "" {
		// try the environment variable
		token = os.Getenv("GITLAB_TOKEN")
	}
	option := make([]gitlab.ClientOptionFunc, 0, 1)
	if config.BaseURL != "" {
		option = append(option, gitlab.WithBaseURL(config.BaseURL))
	}
	client, err := gitlab.NewClient(token, option...)
	if err != nil {
		return nil, fmt.Errorf("cannot create GitLab client: %w", err)
	}
	return &GitLabSource{
		api: client,
	}, nil
}

// ListReleases returns all available releases
func (s *GitLabSource) ListReleases(ctx context.Context, owner, repo string) ([]SourceRelease, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	rels, _, err := s.api.Releases.ListReleases(url.PathEscape(owner+"/"+repo), nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	releases := make([]SourceRelease, len(rels))
	for i, rel := range rels {
		releases[i] = NewGitLabRelease(rel)
	}
	return releases, nil
}

// DownloadReleaseAsset downloads an asset from its ID.
// It returns an io.ReadCloser: it is your responsability to Close it.
// Please note releaseID is not used by GitLabSource.
func (s *GitLabSource) DownloadReleaseAsset(ctx context.Context, owner, repo string, releaseID, id int64) (io.ReadCloser, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Verify interface
var _ Source = &GitLabSource{}
