// Package sourcetest provides a conformance test suite for selfupdate.Source
// implementations, and fake GitHub, GitLab, Gitea and HTTP servers so the real
// sources can run it offline.
package sourcetest

import (
	"context"
	"io"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unknownAssetID is not used by any asset of the fixture
const unknownAssetID = 9999

// Factory returns a Source serving the releases and files held in data.
type Factory func(t *testing.T, data *memory.Source) selfupdate.Source

// Options lists the release information a source cannot report:
// the suite then expects the zero value instead of the fixture value.
type Options struct {
	NoReleaseID  bool
	NoDraft      bool
	NoPrerelease bool
	NoReleaseURL bool
	NoAssetSize  bool
}

// Run executes the conformance suite against sources created by factory.
func Run(t *testing.T, factory Factory, options Options) {
	t.Helper()

	unknownRepository := selfupdate.NewRepositorySlug("owner", "unknown")
	invalidSlug := selfupdate.ParseSlug("invalid")

	t.Run("ListReleases", func(t *testing.T) {
		data := Fixture()
		source := factory(t, data)

		releases, err := source.ListReleases(context.Background(), Repository)
		require.NoError(t, err)
		require.Len(t, releases, len(data.Releases))
		for i, expected := range data.Releases {
			assertRelease(t, expected, releases[i], options)
		}
	})

	t.Run("ListReleasesEmptyRepository", func(t *testing.T) {
		source := factory(t, &memory.Source{Repository: Repository})

		releases, err := source.ListReleases(context.Background(), Repository)
		require.NoError(t, err)
		assert.Empty(t, releases)
	})

	t.Run("ListReleasesUnknownRepository", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.ListReleases(context.Background(), unknownRepository)
		assert.ErrorIs(t, err, selfupdate.ErrRepositoryNotFound)
	})

	t.Run("ListReleasesInvalidSlug", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.ListReleases(context.Background(), invalidSlug)
		assert.ErrorIs(t, err, selfupdate.ErrInvalidSlug)
	})

	t.Run("ListReleasesCancelledContext", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.ListReleases(cancelledContext(), Repository)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("GetLatestRelease", func(t *testing.T) {
		data := Fixture()
		source := factory(t, data)

		release, err := source.GetLatestRelease(context.Background(), Repository)
		require.NoError(t, err)
		assertRelease(t, data.LatestRelease, release, options)
	})

	t.Run("GetLatestReleaseEmptyRepository", func(t *testing.T) {
		source := factory(t, &memory.Source{Repository: Repository})

		_, err := source.GetLatestRelease(context.Background(), Repository)
		assert.ErrorIs(t, err, selfupdate.ErrReleaseNotFound)
	})

	t.Run("GetLatestReleaseUnknownRepository", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.GetLatestRelease(context.Background(), unknownRepository)
		assert.ErrorIs(t, err, selfupdate.ErrRepositoryNotFound)
	})

	t.Run("GetLatestReleaseInvalidSlug", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.GetLatestRelease(context.Background(), invalidSlug)
		assert.ErrorIs(t, err, selfupdate.ErrInvalidSlug)
	})

	t.Run("GetLatestReleaseCancelledContext", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.GetLatestRelease(cancelledContext(), Repository)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("DownloadReleaseAsset", func(t *testing.T) {
		data := Fixture()
		source := factory(t, data)

		// use the releases as seen by the source: download URLs are platform specific
		releases, err := source.ListReleases(context.Background(), Repository)
		require.NoError(t, err)
		for _, release := range releases {
			for _, asset := range release.GetAssets() {
				content := download(t, source, newRelease(release, asset, nil), asset.GetID())
				assert.Equal(t, data.Files[asset.GetID()], content, "%s from %s", asset.GetName(), release.GetTagName())
			}
		}
	})

	t.Run("DownloadValidationAsset", func(t *testing.T) {
		data := Fixture()
		source := factory(t, data)

		release, err := source.GetLatestRelease(context.Background(), Repository)
		require.NoError(t, err)
		assets := release.GetAssets()
		require.Len(t, assets, len(data.LatestRelease.GetAssets()))
		validation := assets[len(assets)-1]

		content := download(t, source, newRelease(release, assets[0], validation), validation.GetID())
		assert.Equal(t, data.Files[validation.GetID()], content)
	})

	t.Run("DownloadReleaseAssetUnknown", func(t *testing.T) {
		source := factory(t, Fixture())

		release, err := source.GetLatestRelease(context.Background(), Repository)
		require.NoError(t, err)

		_, err = source.DownloadReleaseAsset(context.Background(), newRelease(release, release.GetAssets()[0], nil), unknownAssetID)
		assert.ErrorIs(t, err, selfupdate.ErrAssetNotFound)
	})

	t.Run("DownloadReleaseAssetNilRelease", func(t *testing.T) {
		source := factory(t, Fixture())

		_, err := source.DownloadReleaseAsset(context.Background(), nil, 1)
		assert.ErrorIs(t, err, selfupdate.ErrInvalidRelease)
	})

	t.Run("DownloadReleaseAssetCancelledContext", func(t *testing.T) {
		source := factory(t, Fixture())

		release, err := source.GetLatestRelease(context.Background(), Repository)
		require.NoError(t, err)
		asset := release.GetAssets()[0]

		_, err = source.DownloadReleaseAsset(cancelledContext(), newRelease(release, asset, nil), asset.GetID())
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// assertRelease checks a release returned by the source against the fixture it was built from.
func assertRelease(t *testing.T, expected, actual selfupdate.SourceRelease, options Options) {
	t.Helper()

	require.NotNil(t, actual)
	assert.Equal(t, expected.GetTagName(), actual.GetTagName(), "tag name")
	assert.Equal(t, expected.GetName(), actual.GetName(), "name")
	assert.Equal(t, expected.GetReleaseNotes(), actual.GetReleaseNotes(), "release notes")
	assert.WithinDuration(t, expected.GetPublishedAt(), actual.GetPublishedAt(), 0, "published at")
	assertField(t, options.NoReleaseID, expected.GetID(), actual.GetID(), "release ID")
	assertField(t, options.NoDraft, expected.GetDraft(), actual.GetDraft(), "draft")
	assertField(t, options.NoPrerelease, expected.GetPrerelease(), actual.GetPrerelease(), "prerelease")
	assertField(t, options.NoReleaseURL, expected.GetURL(), actual.GetURL(), "URL")

	assets := actual.GetAssets()
	require.Len(t, assets, len(expected.GetAssets()), "assets of %s", expected.GetTagName())
	for i, expectedAsset := range expected.GetAssets() {
		assert.Equal(t, expectedAsset.GetID(), assets[i].GetID(), "asset ID")
		assert.Equal(t, expectedAsset.GetName(), assets[i].GetName(), "asset name")
		assertField(t, options.NoAssetSize, expectedAsset.GetSize(), assets[i].GetSize(), "asset size")
		assert.NotEmpty(t, assets[i].GetBrowserDownloadURL(), "asset download URL")
	}
}

// assertField compares actual to expected, or to the zero value when the source cannot report the field.
func assertField[T any](t *testing.T, unsupported bool, expected, actual T, name string) {
	t.Helper()

	if unsupported {
		assert.Zero(t, actual, name)
		return
	}
	assert.Equal(t, expected, actual, name)
}

// newRelease builds the Release the updater hands to DownloadReleaseAsset for asset,
// with validation pointing at validationAsset when not nil.
func newRelease(release selfupdate.SourceRelease, asset, validationAsset selfupdate.SourceAsset) *selfupdate.Release {
	result := &selfupdate.Release{
		Repository:        Repository,
		AssetURL:          asset.GetBrowserDownloadURL(),
		AssetByteSize:     asset.GetSize(),
		AssetID:           asset.GetID(),
		AssetName:         asset.GetName(),
		ValidationAssetID: -1,
		URL:               release.GetURL(),
		ReleaseID:         release.GetID(),
		ReleaseNotes:      release.GetReleaseNotes(),
		Name:              release.GetName(),
		PublishedAt:       release.GetPublishedAt(),
		Prerelease:        release.GetPrerelease(),
	}
	if validationAsset != nil {
		result.ValidationAssetID = validationAsset.GetID()
		result.ValidationAssetURL = validationAsset.GetBrowserDownloadURL()
	}
	return result
}

// download returns the content of an asset
func download(t *testing.T, source selfupdate.Source, release *selfupdate.Release, assetID int64) []byte {
	t.Helper()

	reader, err := source.DownloadReleaseAsset(context.Background(), release, assetID)
	require.NoError(t, err)
	defer reader.Close()

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	return content
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
