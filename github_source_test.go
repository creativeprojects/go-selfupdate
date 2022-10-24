package selfupdate

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubTokenEnv(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("because $GITHUB_TOKEN is not set")
	}

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
	if _, err := NewGitHubSource(GitHubConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitHub source with API token config")
	}
}

func TestGitHubTokenIsNotSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
}

func TestGitHubEnterpriseClientInvalidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

func TestGitHubEnterpriseClientValidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitHub source with valid URL")
	}
}

func TestGitHubListReleasesContextCancelled(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &Release{repository: ParseSlug("creativeprojects/resticprofile")}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, ErrInvalidRelease)
}
