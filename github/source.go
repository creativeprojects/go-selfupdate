package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/google/go-github/v86/github"
)

// Config is an object to pass to github.NewSource
type Config struct {
	// APIToken represents GitHub API token. If it's not empty, it will be used for authentication of GitHub API
	APIToken string
	// EnterpriseBaseURL is a base URL of GitHub API. If you want to use this library with GitHub Enterprise,
	// please set "https://{your-organization-address}/api/v3/" to this field.
	EnterpriseBaseURL string
	// EnterpriseUploadURL is a URL to upload stuffs to GitHub Enterprise instance. This is often the same as an API base URL.
	// So if this field is not set and EnterpriseBaseURL is set, EnterpriseBaseURL is also set to this field.
	EnterpriseUploadURL string
}

// Source is used to load release information from GitHub
type Source struct {
	api *github.Client
}

// NewSource creates a new github.Source from a config object.
// It initializes a GitHub API client.
// If you set your API token to the $GITHUB_TOKEN environment variable, the client will use it.
// You can pass an empty GitHubSource{} to use the default configuration
// The function will return an error if the GitHub Enterprise URLs in the config object cannot be parsed
func NewSource(config Config) (*Source, error) {
	client := github.NewClient(nil)

	token := config.APIToken
	if token == "" {
		// try the environment variable
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token != "" {
		client = client.WithAuthToken(token)
	}

	if config.EnterpriseBaseURL != "" {
		uploadURL := config.EnterpriseUploadURL
		if uploadURL == "" {
			uploadURL = config.EnterpriseBaseURL
		}
		var err error
		client, err = client.WithEnterpriseURLs(config.EnterpriseBaseURL, uploadURL)
		if err != nil {
			return nil, fmt.Errorf("cannot parse GitHub enterprise URL: %w", err)
		}
	}
	return &Source{
		api: client,
	}, nil
}

// ListReleases returns all available releases
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}
	rels, res, err := s.api.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
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
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}
	rel, res, err := s.api.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
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
	// Check if the AssetURL contains more than one "https://"
	useGithubProxy := strings.Count(rel.AssetURL, "https://") > 1
	// If the AssetURL contains more than 2 "https://", it means it's using a GitHub Proxy service.
	// In this case, we should download the asset directly from the AssetURL instead of using the GitHub API.
	// This is a workaround for the issue that the GitHub API does not support downloading assets from GitHub Proxy services.
	if useGithubProxy {
		// Determine download url based on asset id.
		var downloadURL string
		if rel.AssetID == assetID {
			downloadURL = rel.AssetURL
		} else if rel.ValidationAssetID == assetID {
			downloadURL = rel.ValidationAssetURL
		}
		if downloadURL == "" {
			return nil, fmt.Errorf("asset ID %d: %w", assetID, selfupdate.ErrAssetNotFound)
		}
		// Download the asset directly from the AssetURL
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create download request:%w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("download failed:%w", err)
		}

		// The caller is responsible for closing resp.Body
		if resp.StatusCode != http.StatusOK {
			defer resp.Body.Close()
			return nil, fmt.Errorf("download failed, status code:%d", resp.StatusCode)
		}

		return resp.Body, nil
	}
	// continue with the normal GitHub API download
	owner, repo, err := rel.GetSlug()
	if err != nil {
		return nil, err
	}
	// create a new http client so the GitHub library can download the redirected file (if any)
	client := http.DefaultClient
	rc, _, err := s.api.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID, client)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub Releases API for getting the asset ID %d on repository '%s/%s': %w", assetID, owner, repo, err)
	}
	return rc, nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
