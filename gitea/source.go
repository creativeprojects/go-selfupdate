// Package gitea provides a selfupdate.Source loading releases from a Gitea instance.
//
// Config.BaseURL is mandatory. An API token is read from Config.APIToken
// or from the $GITEA_TOKEN environment variable.
package gitea

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"code.gitea.io/sdk/gitea"
	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal"
)

// Config is an object to pass to NewSource
type Config struct {
	// APIToken represents Gitea API token. If it's not empty, it will be used for authentication for the API
	APIToken string
	// BaseURL is a base URL of your gitea instance. This parameter has NO default value.
	BaseURL string
}

// Source is used to load release information from Gitea
type Source struct {
	api     *gitea.Client
	token   string
	baseURL string
}

// NewSource creates a new Source from a config object.
// It initializes a Gitea API Client.
// If you set your API token to the $GITEA_TOKEN environment variable,
// the client will automatically use it.
func NewSource(ctx context.Context, config Config) (*Source, error) {
	token := config.APIToken
	if token == "" {
		// try the environment variable
		token = os.Getenv("GITEA_TOKEN")
	}
	if config.BaseURL == "" {
		return nil, fmt.Errorf("gitea base url must be set")
	}

	client, err := gitea.NewClient(config.BaseURL, gitea.SetContext(ctx), gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("error connecting to gitea: %w", err)
	}

	return &Source{
		api:     client,
		token:   token,
		baseURL: config.BaseURL,
	}, nil
}

// ListReleases returns all available releases
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}

	s.api.SetContext(ctx)
	rels, res, err := s.api.ListReleases(owner, repo, gitea.ListReleasesOptions{})
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

	s.api.SetContext(ctx)
	rel, res, err := s.api.GetLatestRelease(owner, repo)
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
	owner, repo, err := rel.GetSlug()
	if err != nil {
		return nil, err
	}
	s.api.SetContext(ctx)
	attachment, res, err := s.api.GetReleaseAttachment(owner, repo, rel.ReleaseID, assetID)
	if err != nil {
		if res != nil && res.StatusCode == http.StatusNotFound {
			err = selfupdate.ErrAssetNotFound
		}
		return nil, fmt.Errorf("failed to call Gitea Releases API for getting the asset ID %d on repository '%s/%s': %w", assetID, owner, repo, err)
	}

	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, attachment.DownloadURL, http.NoBody)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	if s.token != "" {
		// verify request is from same domain not to leak token
		ok, err := internal.CanUseTokenForDomain(s.baseURL, attachment.DownloadURL)
		if err != nil {
			return nil, err
		}
		if ok {
			req.Header.Set("Authorization", "token "+s.token)
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
