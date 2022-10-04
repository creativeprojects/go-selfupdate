package selfupdate

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGiteaTokenEnv(t *testing.T) {
	token := os.Getenv("GITEA_TOKEN")
	if token == "" {
		t.Skip("because $GITEA_TOKEN is not set")
	}

	if _, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"}); err != nil {
		t.Error("Failed to initialize Gitea source with URL")
	}
	if _, err := NewGiteaSource(GiteaConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize Gitea source with API token config")
	}
}

func TestGiteaTokenIsNotSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	if _, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"}); err != nil {
		t.Error("Failed to initialize Gitea source with URL")
	}
}

func TestGiteaLatestReleaseNotSupported(t *testing.T) {
	source, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"})
	require.NoError(t, err)

	_, err = source.LatestRelease(context.Background(), ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, ErrNotSupported)
}

func TestGiteaListReleasesContextCancelled(t *testing.T) {
	source, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGiteaDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &Release{repository: ParseSlug("creativeprojects/resticprofile")}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGiteaDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewGiteaSource(GiteaConfig{BaseURL: "https://git.lbsfilm.at"})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, ErrInvalidRelease)
}
