package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"code.gitea.io/sdk/gitea"
)

// GiteaConfig is an object to pass to NewGiteaSource
type GiteaConfig struct {
	// APIToken represents Gitea API token. If it's not empty, it will be used for authentication for the API
	APIToken string
	// BaseURL is a base URL of your gitea instance
	BaseURL string
	// Context used by the http client (default to context.Background)
	Context context.Context
}

// GiteaSource is used to load release information from Gitea
type GiteaSource struct {
	api   *gitea.Client
	ctx   context.Context
	token string
}

// NewGiteaSource creates a new NewGiteaSource from a config object.
// It initializes a Gitea API Client.
// If you set your API token to the $GITEA_TOKEN environment variable, the client will use it.
// You can pass an empty GiteaSource{} to use the default configuration
func NewGiteaSource(config GiteaConfig) (*GiteaSource, error) {
	token := config.APIToken
	if token == "" {
		// try the environment variable
		token = os.Getenv("GITEA_TOKEN")
	}
	if config.BaseURL == "" {
		return nil, fmt.Errorf("gitea base url must be set")
	}

	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}

	client, err := gitea.NewClient(config.BaseURL, gitea.SetContext(ctx), gitea.SetToken(token))
	if err != nil {
		return nil, fmt.Errorf("Error connecting to gitea: %w", err)
	}

	return &GiteaSource{
		api:   client,
		ctx:   ctx,
		token: token,
	}, nil
}

// ListReleases returns all available releases
func (s *GiteaSource) ListReleases(owner, repo string) ([]SourceRelease, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}

	rels, res, err := s.api.ListReleases(owner, repo, gitea.ListReleasesOptions{})
	if err != nil {
		log.Printf("API returned an error response: %s", err)
		if res != nil && res.StatusCode == 404 {
			// 404 means repository not found or release not found. It's not an error here.
			log.Print("API returned 404. Repository or release not found")
			return nil, nil
		}
		return nil, err
	}
	releases := make([]SourceRelease, len(rels))
	for i, rel := range rels {
		releases[i] = NewGiteaRelease(rel)
	}
	return releases, nil
}

// DownloadReleaseAsset downloads an asset from its ID.
// It returns an io.ReadCloser: it is your responsability to Close it.
func (s *GiteaSource) DownloadReleaseAsset(owner, repo string, releaseID, id int64) (io.ReadCloser, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	// create a new http client so the GitHub library can download the redirected file (if any)
	// don't pass the "default" one as it could be the one it's already using
	attachment, _, err := s.api.GetReleaseAttachment(owner, repo, releaseID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to call Gitea Releases API for getting the asset ID %d on repository '%s/%s': %w", id, owner, repo, err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", attachment.DownloadURL, nil)
	if err != nil {
		return nil, err
	}
	log.Print(err)

	req.Header.Set("Authorization", "token "+s.token)
	rc, err := client.Do(req)
	log.Print(err)

	if err != nil {
		return nil, err
	}

	return rc.Body, nil
}

// Verify interface
var _ Source = &GiteaSource{}
