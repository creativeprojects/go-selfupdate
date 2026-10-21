package github

import (
	"context"
	"os"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	"github.com/creativeprojects/go-selfupdate/v2/internal/sourcetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubTokenEnv(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("because $GITHUB_TOKEN is not set")
	}

	if _, err := NewSource(Config{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
	if _, err := NewSource(Config{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitHub source with API token config")
	}
}

func TestGitHubTokenIsNotSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	if _, err := NewSource(Config{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
}

func TestGitHubEnterpriseClientInvalidURL(t *testing.T) {
	_, err := NewSource(Config{APIToken: "my_token", EnterpriseBaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

func TestGitHubEnterpriseClientValidURL(t *testing.T) {
	_, err := NewSource(Config{APIToken: "my_token", EnterpriseBaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitHub source with valid URL")
	}
}

func TestGitHubListReleasesContextCancelled(t *testing.T) {
	source, err := NewSource(Config{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, selfupdate.ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewSource(Config{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &selfupdate.Release{Repository: selfupdate.ParseSlug("creativeprojects/resticprofile")}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewSource(Config{})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, selfupdate.ErrInvalidRelease)
}

// TestGitHubSource runs the shared source test suite against a fake GitHub API.
func TestGitHubSource(t *testing.T) {
	sourcetest.Run(t, func(t *testing.T, data *memory.Source) selfupdate.Source {
		server := sourcetest.NewGitHubServer(t, data)
		source, err := NewSource(Config{EnterpriseBaseURL: server.URL})
		require.NoError(t, err)
		return source
	}, sourcetest.Options{})
}
