package gitea

import (
	"context"
	"os"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/sourcetest"
	"github.com/creativeprojects/go-selfupdate/v2/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGiteaTokenEnv(t *testing.T) {
	token := os.Getenv("GITEA_TOKEN")
	if token == "" {
		t.Skip("because $GITEA_TOKEN is not set")
	}

	if _, err := NewSource(context.Background(), Config{BaseURL: sourcetest.NewGiteaServer(t, sourcetest.Fixture()).URL}); err != nil {
		t.Error("Failed to initialize Gitea source with URL")
	}
	if _, err := NewSource(context.Background(), Config{APIToken: token}); err != nil {
		t.Error("Failed to initialize Gitea source with API token config")
	}
}

func TestGiteaTokenIsNotSet(t *testing.T) {
	t.Setenv("GITEA_TOKEN", "")

	if _, err := NewSource(context.Background(), Config{BaseURL: sourcetest.NewGiteaServer(t, sourcetest.Fixture()).URL}); err != nil {
		t.Error("Failed to initialize Gitea source with URL")
	}
}

func TestGiteaListReleasesContextCancelled(t *testing.T) {
	source, err := NewSource(context.Background(), Config{BaseURL: sourcetest.NewGiteaServer(t, sourcetest.Fixture()).URL})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, selfupdate.ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGiteaDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewSource(context.Background(), Config{BaseURL: sourcetest.NewGiteaServer(t, sourcetest.Fixture()).URL})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &selfupdate.Release{Repository: selfupdate.ParseSlug("creativeprojects/resticprofile")}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGiteaDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewSource(context.Background(), Config{BaseURL: sourcetest.NewGiteaServer(t, sourcetest.Fixture()).URL})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, selfupdate.ErrInvalidRelease)
}

// TestGiteaSource runs the shared source test suite against a fake Gitea API.
func TestGiteaSource(t *testing.T) {
	sourcetest.Run(t, func(t *testing.T, data *memory.Source) selfupdate.Source {
		server := sourcetest.NewGiteaServer(t, data)
		source, err := NewSource(context.Background(), Config{BaseURL: server.URL})
		require.NoError(t, err)
		return source
	}, sourcetest.Options{
		// the Gitea release has no URL (see FIXME in NewRelease)
		NoReleaseURL: true,
	})
}
