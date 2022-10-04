package selfupdate

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitLabTokenEnv(t *testing.T) {
	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		t.Skip("because $GITLAB_TOKEN is not set")
	}

	if _, err := NewGitLabSource(GitLabConfig{}); err != nil {
		t.Error("Failed to initialize GitLab source with empty config")
	}
	if _, err := NewGitLabSource(GitLabConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitLab source with API token config")
	}
}

func TestGitLabTokenIsNotSet(t *testing.T) {
	t.Setenv("GITLAB_TOKEN", "")

	if _, err := NewGitLabSource(GitLabConfig{}); err != nil {
		t.Error("Failed to initialize GitLab source with empty config")
	}
}

func TestGitLabEnterpriseClientInvalidURL(t *testing.T) {
	_, err := NewGitLabSource(GitLabConfig{APIToken: "my_token", BaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

func TestGitLabEnterpriseClientValidURL(t *testing.T) {
	_, err := NewGitLabSource(GitLabConfig{APIToken: "my_token", BaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitLab source with valid URL")
	}
}

func TestGitLabLatestReleaseContextCancelled(t *testing.T) {
	source, err := NewGitLabSource(GitLabConfig{})
	require.NoError(t, err)

	_, err = source.LatestRelease(context.Background(), ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, ErrNotSupported)
}

func TestGitLabListReleasesContextCancelled(t *testing.T) {
	source, err := NewGitLabSource(GitLabConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitLabDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewGitLabSource(GitLabConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &Release{
		AssetID:  11,
		AssetURL: "http://localhost/",
	}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitLabDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewGitLabSource(GitLabConfig{})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, ErrInvalidRelease)
}

func TestGitLabDownloadReleaseAssetNotFound(t *testing.T) {
	source, err := NewGitLabSource(GitLabConfig{})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), &Release{
		AssetID:           11,
		ValidationAssetID: 12,
	}, 13)
	assert.ErrorIs(t, err, ErrAssetNotFound)
}
