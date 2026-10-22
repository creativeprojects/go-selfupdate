// Package gitlab provides a selfupdate.Source loading releases from gitlab.com
// or a self-managed GitLab instance.
//
// An API token is read from Config.APIToken or from the $GITLAB_TOKEN environment variable.
// GitLab releases have no ID, draft or prerelease flag, and release links have no size:
// the source reports zero values for them.
package gitlab

import (
	"context"
	"fmt"
	"io"
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
	// Logger receives the messages describing what the source is doing. Default is to discard them.
	Logger selfupdate.Logger
}

// Source is used to load release information from GitLab
type Source struct {
	api     *gitlab.Client
	token   string
	baseURL string
	logger  selfupdate.Logger
}

// NewSource creates a new gitlab.Source from a config object.
// It initializes a GitLab API client.
// If you set your API token to the $GITLAB_TOKEN environment variable, the client will use it.
// You can pass an empty Config{} to use the default configuration (gitlab.com).
// The function will return an error if the base URL in the config object cannot be parsed
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
	logger := config.Logger
	if logger == nil {
		logger = selfupdate.DiscardLogger
	}
	return &Source{
		api:     client,
		token:   token,
		baseURL: config.BaseURL,
		logger:  logger,
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
			return nil, s.notFoundError(ctx, pid)
		}
		return nil, err
	}
	return NewRelease(rel), nil
}

// notFoundError tells a project without any release from a missing project:
// the API answers 404 to both when asked for the latest release.
func (s *Source) notFoundError(ctx context.Context, pid any) error {
	_, resp, err := s.api.Projects.GetProject(pid, nil, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return selfupdate.ErrRepositoryNotFound
		}
		return err
	}
	return selfupdate.ErrReleaseNotFound
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

	s.logger.Printf("downloading %q", downloadURL)
	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
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
		return nil, err
	}

	return response.Body, nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
