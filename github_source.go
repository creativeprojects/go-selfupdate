package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

// GitHubConfig is an object to pass to NewGitHubSource
type GitHubConfig struct {
	// APIToken represents GitHub API token. If it's not empty, it will be used for authentication of GitHub API
	APIToken string
	// EnterpriseBaseURL is a base URL of GitHub API. If you want to use this library with GitHub Enterprise,
	// please set "https://{your-organization-address}/api/v3/" to this field.
	EnterpriseBaseURL string
	// EnterpriseUploadURL is a URL to upload stuffs to GitHub Enterprise instance. This is often the same as an API base URL.
	// So if this field is not set and EnterpriseBaseURL is set, EnterpriseBaseURL is also set to this field.
	EnterpriseUploadURL string
	// Context used by the http client (default to context.Background)
	Context context.Context
}

// GitHubSource is used to load release information from GitHub
type GitHubSource struct {
	api *github.Client
	ctx context.Context
}

// NewGitHubSource creates a new GitHubSource from a config object.
// It initializes a GitHub API client.
// If you set your API token to the $GITHUB_TOKEN environment variable, the client will use it.
// You can pass an empty GitHubSource{} to use the default configuration
// The function will return an error if the GitHub Entreprise URLs in the config object cannot be parsed
func NewGitHubSource(config GitHubConfig) (*GitHubSource, error) {
	token := config.APIToken
	if token == "" {
		// try the environment variable
		token = os.Getenv("GITHUB_TOKEN")
	}
	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}
	hc := newHTTPClient(ctx, token)

	if config.EnterpriseBaseURL == "" {
		// public (or private) repository on standard GitHub offering
		client := github.NewClient(hc)
		return &GitHubSource{
			api: client,
			ctx: ctx,
		}, nil
	}

	u := config.EnterpriseUploadURL
	if u == "" {
		u = config.EnterpriseBaseURL
	}
	client, err := github.NewEnterpriseClient(config.EnterpriseBaseURL, u, hc)
	if err != nil {
		return nil, fmt.Errorf("cannot parse GitHub entreprise URL: %w", err)
	}
	return &GitHubSource{
		api: client,
		ctx: ctx,
	}, nil
}

// ListReleases returns all available releases
func (s *GitHubSource) ListReleases(owner, repo string) ([]SourceRelease, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	rels, res, err := s.api.Repositories.ListReleases(s.ctx, owner, repo, nil)
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
		releases[i] = NewGitHubRelease(rel)
	}
	return releases, nil
}

// DownloadReleaseAsset downloads an asset from its ID.
// It returns an io.ReadCloser: it is your responsability to Close it.
// Please note releaseID is not used by GitHubSource.
func (s *GitHubSource) DownloadReleaseAsset(owner, repo string, releaseID, id int64) (io.ReadCloser, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	// create a new http client so the GitHub library can download the redirected file (if any)
	// don't pass the "default" one as it could be the one it's already using
	client := &http.Client{}
	rc, _, err := s.api.Repositories.DownloadReleaseAsset(s.ctx, owner, repo, id, client)
	if err != nil {
		return nil, fmt.Errorf("failed to call GitHub Releases API for getting the asset ID %d on repository '%s/%s': %w", id, owner, repo, err)
	}
	return rc, nil
}

func newHTTPClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return &http.Client{}
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}

// Verify interface
var _ Source = &GitHubSource{}
