package gitlab

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Config is an object to pass to gitlab.NewSource
type Config struct {
	// APIToken represents GitLab API token. If it's not empty, it will be used for authentication for the API
	APIToken string
	// BaseURL is a base URL of your private GitLab instance
	BaseURL string
}

// Source is used to load release information from GitLab
type Source struct {
	api     *gitlab.Client
	token   string
	baseURL string
}

// NewSource creates a new GitLabSource from a config object.
// It initializes a GitLab API client.
// If you set your API token to the $GITLAB_TOKEN environment variable, the client will use it.
// You can pass an empty GitLabSource{} to use the default configuration
// The function will return an error if the GitLab Enterprise URLs in the config object cannot be parsed
func NewSource(config Config) (*Source, error) {
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
	return &Source{
		api:     client,
		token:   token,
		baseURL: config.BaseURL,
	}, nil
}

// ListReleases returns all available releases
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	pid, err := repository.Get()
	if err != nil {
		return nil, err
	}

	rels, resp, err := s.api.Releases.ListReleases(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, selfupdate.ErrRepositoryNotFound
		}
		return nil, err
	}
	releases := make([]selfupdate.SourceRelease, len(rels))
	for i, rel := range rels {
		releases[i] = NewRelease(rel)
	}
	return releases, nil
}

func (s *Source) GetLatestRelease(ctx context.Context, repository selfupdate.Repository) (selfupdate.SourceRelease, error) {
	pid, err := repository.Get()
	if err != nil {
		return nil, err
	}

	rel, resp, err := s.api.Releases.GetLatestRelease(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, selfupdate.ErrRepositoryNotFound
		}
		return nil, err
	}
	return NewRelease(rel), nil
}

// DownloadReleaseAsset downloads an asset from a release.
// It returns an io.ReadCloser: it is your responsibility to Close it.
func (s *Source) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if rel == nil {
		return nil, selfupdate.ErrInvalidRelease
	}
	var downloadURL string
	if rel.AssetID == assetID {
		downloadURL = rel.AssetURL
	} else if rel.ValidationAssetID == assetID {
		downloadURL = rel.ValidationAssetURL
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("asset ID %d: %w", assetID, selfupdate.ErrAssetNotFound)
	}

	log.Printf("downloading %q", downloadURL)
	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if s.token != "" {
		// verify request is from same domain not to leak token
		ok, err := internal.CanUseTokenForDomain(s.baseURL, downloadURL)
		if err != nil {
			return nil, err
		}
		if ok {
			req.Header.Set("PRIVATE-TOKEN", s.token)
		}
	}
	response, err := client.Do(req)

	if err != nil {
		log.Print(err)
		return nil, err
	}

	return response.Body, nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
