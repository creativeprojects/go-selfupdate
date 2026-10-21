package selfupdate_test

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/github"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testGithubRepository = selfupdate.NewRepositorySlug("creativeprojects", "resticprofile")
)

func skipRateLimitExceeded(t *testing.T, err error) {
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "403 API rate limit") {
		t.Skip("Test skipped because of GitHub API rate limit exceeded")
	}
}

func TestDetectReleaseWithVersionPrefix(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: mockMemorySource(t)})},
		{!testing.Short(), "GitHub", newGitHubUpdater(t, selfupdate.Config{})},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, err := testItem.updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{testGithubRepository})
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			require.NotNil(t, r, "No release returned")
			if r.LessThan("0.14.0") {
				t.Error("Incorrect version:", r.Version())
			}
			if !strings.HasSuffix(r.AssetURL, ".zip") && !strings.HasSuffix(r.AssetURL, ".tar.gz") {
				t.Error("Incorrect URL for asset:", r.AssetURL)
			}
			assert.NotEmpty(t, r.URL, "Document URL should not be empty")
			assert.NotEmpty(t, r.ReleaseNotes, "Description should not be empty for this repo")
			assert.NotEmpty(t, r.Name, "Release name is unexpectedly empty")
			assert.NotEmpty(t, r.AssetByteSize, "Asset's size is unexpectedly zero")
			assert.NotEmpty(t, r.AssetID, "Asset's ID is unexpectedly zero")
			assert.NotZero(t, r.PublishedAt, "Release time is unexpectedly zero")
		})
	}
}

func TestDetectVersionExisting(t *testing.T) {
	testVersion := "v0.14.0"
	gitHub := newGitHubUpdater(t, selfupdate.Config{Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"}})
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: mockMemorySource(t), Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"}})},
		{!testing.Short(), "GitHub", gitHub},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, err := testItem.updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, testVersion})
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			require.NotNil(t, r, "No release returned")
			assert.Greater(t, r.ValidationAssetID, int64(-1))
		})
	}
}

func TestDetectVersionExistingWithNoValidationFile(t *testing.T) {
	testVersion := "v0.14.0"
	gitHub := newGitHubUpdater(t, selfupdate.Config{Validator: &selfupdate.ChecksumValidator{UniqueFilename: "notfound.txt"}})
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: mockMemorySource(t), Validator: &selfupdate.ChecksumValidator{UniqueFilename: "notfound.txt"}})},
		{!testing.Short(), "GitHub", gitHub},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			_, err := testItem.updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, testVersion})
			skipRateLimitExceeded(t, err)
			assert.ErrorIs(t, err, selfupdate.ErrValidationAssetNotFound)
		})
	}
}

func TestDetectVersionNotExisting(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: mockMemorySource(t)})},
		{!testing.Short(), "GitHub", newGitHubUpdater(t, selfupdate.Config{})},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, err := testItem.updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, "foobar"})
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.Nil(t, r, "Release not detected but got a returned value for it")
		})
	}
}

func TestDetectPrerelease(t *testing.T) {
	testFixtures := []struct {
		prerelease bool
		version    string
	}{
		{false, "1.0.0"},
		{true, "2.0.0-beta"},
	}
	for _, testFixture := range testFixtures {
		t.Run(strconv.FormatBool(testFixture.prerelease), func(t *testing.T) {
			updater := newMockUpdater(t, selfupdate.Config{
				Source:     mockMemorySource(t),
				Prerelease: testFixture.prerelease,
			})
			r, err := updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{selfupdate.NewRepositorySlug("owner", "repo")})
			require.NotNil(t, r)
			assert.NoError(t, err)

			assert.Equal(t, testFixture.prerelease, r.Prerelease)
			assert.Equal(t, testFixture.version, r.Version())
		})
	}
}

func TestDetectReleasesForVariousArchives(t *testing.T) {
	for _, tc := range []struct {
		slug   string
		prefix string
	}{
		{"rhysd-test/test-release-zip", "v"},
		{"rhysd-test/test-release-tar", "v"},
		{"rhysd-test/test-release-gzip", "v"},
		{"rhysd-test/test-release-xz", "release-v"},
		{"rhysd-test/test-release-tar-xz", "release-"},
	} {
		t.Run(tc.slug, func(t *testing.T) {
			source, err := github.NewSource(github.Config{})
			require.NoError(t, err, "failed to create source")
			updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source, Arch: "amd64"})
			require.NoError(t, err, "failed to create updater")
			r, err := updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{selfupdate.ParseSlug(tc.slug)})
			skipRateLimitExceeded(t, err)

			assert.NoError(t, err, "fetch failed")
			require.NotNil(t, r, "release not detected")
			assert.Truef(t, r.Equal("1.2.3"), "incorrect release: expected 1.2.3 but got %v", r.Version())

			url := fmt.Sprintf("https://github.com/%s/releases/tag/%s1.2.3", tc.slug, tc.prefix)
			assert.Equal(t, url, r.URL)
			assert.NotEmpty(t, r.ReleaseNotes, "Release note is unexpectedly empty")

			if !strings.HasPrefix(r.AssetURL, fmt.Sprintf("https://github.com/%s/releases/download/%s1.2.3/", tc.slug, tc.prefix)) {
				t.Error("Unexpected asset URL:", r.AssetURL)
			}

			assert.NotEmpty(t, r.Name, "Release name is unexpectedly empty")
			assert.NotEmpty(t, r.AssetByteSize, "Asset's size is unexpectedly zero")
			assert.NotEmpty(t, r.AssetID, "Asset's ID is unexpectedly zero")
			assert.NotZero(t, r.PublishedAt)
		})
	}
}

func TestDetectReleaseButNoAsset(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: newMemorySource(
			[]selfupdate.SourceRelease{
				&memory.Release{
					Name:    "first",
					TagName: "v1.0",
					Assets:  nil,
				},
				&memory.Release{
					Name:    "second",
					TagName: "v2.0",
					Assets:  nil,
				},
			},
			nil,
		)})},
		{!testing.Short(), "GitHub", newGitHubUpdater(t, selfupdate.Config{})},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			release, err := testItem.updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{selfupdate.ParseSlug("rhysd/clever-f.vim")})
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.NotNil(t, release)
		})
	}
}

func TestNonExistingRepo(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: newRepositoryNotFoundSource()})},
		{!testing.Short(), "GitHub", newGitHubUpdater(t, selfupdate.Config{})},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			_, err := testItem.updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{selfupdate.ParseSlug("creativeprojects/non-existing-repo")})
			skipRateLimitExceeded(t, err)
			require.ErrorIs(t, err, selfupdate.ErrRepositoryNotFound)
		})
	}
}

func TestNoReleaseFound(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *selfupdate.Updater
	}{
		{true, "Mock", newMockUpdater(t, selfupdate.Config{Source: newMemorySource(nil, nil)})},
		{!testing.Short(), "GitHub", newGitHubUpdater(t, selfupdate.Config{})},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			rel, err := testItem.updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{selfupdate.ParseSlug("rhysd/misc")})
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.Nil(t, rel, "Should have not found any release")
		})
	}
}

func TestFindAssetFromRelease(t *testing.T) {
	type findReleaseAndAssetFixture struct {
		name            string
		config          selfupdate.Config
		release         selfupdate.SourceRelease
		targetVersion   string
		expectedAsset   string
		expectedVersion string
		expectedFound   bool
	}

	rel1 := "rel1"
	v1 := "1.0.0"
	rel11 := "rel11"
	v11 := "1.1.0"
	asset1 := "asset1.gz"
	asset2 := "asset2.gz"
	wrongAsset1 := "asset1.yaml"
	asset11 := "asset11.gz"
	url1 := "https://asset1"
	url2 := "https://asset2"
	url11 := "https://asset11"

	testData := []findReleaseAndAssetFixture{
		{
			name:          "empty fixture",
			config:        selfupdate.Config{},
			release:       nil,
			targetVersion: "",
			expectedFound: false,
		},
		{
			name: "find asset, no filters",
			release: &memory.Release{
				Name:    rel1,
				TagName: v1,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url1},
				},
			},
			targetVersion:   "1.0.0",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "find asset, no target version",
			release: &memory.Release{
				Name:    rel1,
				TagName: v1,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url1},
				},
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "don't find prerelease",
			release: &memory.Release{
				Name:    rel1,
				TagName: v1,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url1},
				},
				Prerelease: true,
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   false,
		},
		{
			name: "find named prerelease",
			release: &memory.Release{
				Name:    rel1,
				TagName: v1,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url1},
				},
				Prerelease: true,
			},
			targetVersion:   "1.0.0",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name:   "find prerelease",
			config: selfupdate.Config{Prerelease: true},
			release: &memory.Release{
				Name:    rel1,
				TagName: v1,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url1},
				},
				Prerelease: true,
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "don't find asset with wrong extension, no filters",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: wrongAsset1, URL: url11},
				},
			},
			targetVersion: "1.1.0",
			expectedFound: false,
		},
		{
			name: "find asset with different name, no filters",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset1, URL: url11},
				},
			},
			targetVersion:   "1.1.0",
			expectedAsset:   asset1,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, no filters (2)",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset11, URL: url11},
				},
			},
			targetVersion:   "1.1.0",
			expectedAsset:   asset11,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match filter",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset11, URL: url11},
					&memory.Asset{Name: asset1, URL: url1},
				},
			},
			targetVersion:   "1.1.0",
			config:          selfupdate.Config{Filters: []string{"11"}},
			expectedAsset:   asset11,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match another filter",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset11, URL: url11},
					&memory.Asset{Name: asset1, URL: url1},
				},
			},
			targetVersion:   "1.1.0",
			config:          selfupdate.Config{Filters: []string{"([^1])1{1}([^1])"}},
			expectedAsset:   asset1,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match any filter",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset11, URL: url11},
					&memory.Asset{Name: asset2, URL: url2},
				},
			},
			targetVersion:   "1.1.0",
			config:          selfupdate.Config{Filters: []string{"([^1])1{1}([^1])", "([^1])2{1}([^1])"}},
			expectedAsset:   asset2,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match no filter",
			release: &memory.Release{
				Name:    rel11,
				TagName: v11,
				Assets: []selfupdate.SourceAsset{
					&memory.Asset{Name: asset11, URL: url11},
					&memory.Asset{Name: asset2, URL: url2},
				},
			},
			targetVersion: "",
			config:        selfupdate.Config{Filters: []string{"another", "binary"}},
			expectedFound: false,
		},
	}

	for _, fixture := range testData {
		t.Run(fixture.name, func(t *testing.T) {
			updater := newMockUpdater(t, fixture.config)
			asset, ver, found := updater.FindAssetFromRelease(fixture.release, []string{".gz"}, fixture.targetVersion)
			if fixture.expectedFound {
				if !found {
					t.Fatalf("expected to find an asset for this fixture: %q", fixture.name)
				}
				if asset.GetName() == "" {
					t.Fatalf("invalid asset struct returned from fixture: %q, got: %v", fixture.name, asset)
				}
				if asset.GetName() != fixture.expectedAsset {
					t.Fatalf("expected asset %q in fixture: %q, got: %s", fixture.expectedAsset, fixture.name, asset.GetName())
				}
				t.Logf("asset %v, %v", asset, ver)
			} else if found {
				t.Fatalf("expected not to find an asset for this fixture: %q, but got: %v", fixture.name, asset)
			}
		})
	}
}

func TestFindReleaseAndAsset(t *testing.T) {
	selfupdate.SetLogger(stdlog.New(os.Stderr, "", 0))
	defer selfupdate.SetLogger(stdlog.New(io.Discard, "", 0))

	tag2 := "v2.0.0"
	rel2 := "rel2"
	assetLinux386 := "asset_linux_386.tgz"
	assetLinuxAMD64 := "asset_linux_amd64.tgz"
	assetLinuxX86 := "asset_linux_x86_64.tgz"
	assetLinuxARM := "asset_linux_arm.tgz"
	assetLinuxARMv5 := "asset_linux_armv5.tgz"
	assetLinuxARMv6 := "asset_linux_armv6.tgz"
	assetLinuxARMv7 := "asset_linux_armv7.tgz"
	assetLinuxARM64 := "asset_linux_arm64.tgz"
	assetLinuxAll := "asset_linux_all.tgz"
	assetDarwinAMD64 := "asset_darwin_amd64.tgz"
	assetDarwinARM64 := "asset_darwin_arm64.tgz"
	assetDarwinAll := "asset_darwin_all.tgz"

	testData := []struct {
		name              string
		os                string
		arch              string
		arm               uint8
		universalArch     string
		releases          []selfupdate.SourceRelease
		version           string
		filters           []string
		found             bool
		expectedAssetName string
	}{
		{
			name: "no match",
			os:   "darwin",
			arch: "amd64",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinux386,
						},
						&memory.Asset{
							Name: assetLinuxAMD64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             false,
			expectedAssetName: assetLinuxAMD64,
		},
		{
			name: "simple match",
			os:   "linux",
			arch: "amd64",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinux386,
						},
						&memory.Asset{
							Name: assetLinuxAMD64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxAMD64,
		},
		{
			name: "simple match case insensitive",
			os:   "linux",
			arch: "amd64",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinux386,
						},
						&memory.Asset{
							Name: "asset_Linux_AMD64.tgz",
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: "asset_Linux_AMD64.tgz",
		},
		{
			name: "match default arm",
			os:   "linux",
			arch: "arm",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxARM,
						},
						&memory.Asset{
							Name: assetLinuxARM64,
						},
						&memory.Asset{
							Name: assetLinuxARMv5,
						},
						&memory.Asset{
							Name: assetLinuxARMv6,
						},
						&memory.Asset{
							Name: assetLinuxARMv7,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxARM,
		},
		{
			name: "match armv6",
			os:   "linux",
			arch: "arm",
			arm:  6,
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxARM,
						},
						&memory.Asset{
							Name: assetLinuxARM64,
						},
						&memory.Asset{
							Name: assetLinuxARMv5,
						},
						&memory.Asset{
							Name: assetLinuxARMv6,
						},
						&memory.Asset{
							Name: assetLinuxARMv7,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxARMv6,
		},
		{
			name: "fallback to armv5",
			os:   "linux",
			arch: "arm",
			arm:  7,
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxARM,
						},
						&memory.Asset{
							Name: assetLinuxARM64,
						},
						&memory.Asset{
							Name: assetLinuxARMv5,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxARMv5,
		},
		{
			name: "fallback to arm",
			os:   "linux",
			arch: "arm",
			arm:  5,
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxARM,
						},
						&memory.Asset{
							Name: assetLinuxARM64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxARM,
		},
		{
			name: "arm not found",
			os:   "linux",
			arch: "arm",
			arm:  6,
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxARMv7,
						},
						&memory.Asset{
							Name: assetLinuxARM64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             false,
			expectedAssetName: assetLinuxARM,
		},
		{
			name: "match x86_64 for adm64",
			os:   "linux",
			arch: "amd64",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinux386,
						},
						&memory.Asset{
							Name: assetLinuxX86,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetLinuxX86,
		},
		{
			name:          "universal binary ignored on linux",
			os:            "linux", // universal binary is for darwin only
			arch:          "amd64",
			universalArch: "all",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetLinuxAll,
						},
					},
				},
			},
			version: "v2.0.0",
			filters: nil,
			found:   false,
		},
		{
			name:          "match amd64 instead of universal binary",
			os:            "darwin", // universal binary is for darwin only
			arch:          "amd64",
			universalArch: "all",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetDarwinAMD64,
						},
						&memory.Asset{
							Name: assetDarwinARM64,
						},
						&memory.Asset{
							Name: assetDarwinAll,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetDarwinAMD64,
		},
		{
			name:          "match arm64 instead of universal binary",
			os:            "darwin", // universal binary is for darwin only
			arch:          "arm64",
			universalArch: "all",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetDarwinAMD64,
						},
						&memory.Asset{
							Name: assetDarwinARM64,
						},
						&memory.Asset{
							Name: assetDarwinAll,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetDarwinARM64,
		},
		{
			name:          "match universal binary",
			os:            "darwin", // universal binary is for darwin only
			arch:          "arm64",
			universalArch: "all",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetDarwinAll,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: assetDarwinAll,
		},
		{
			name:          "no match when universal binary not specified",
			os:            "darwin",
			arch:          "arm64",
			universalArch: "",
			releases: []selfupdate.SourceRelease{
				&memory.Release{
					Name:    rel2,
					TagName: tag2,
					Assets: []selfupdate.SourceAsset{
						&memory.Asset{
							Name: assetDarwinAll,
						},
					},
				},
			},
			version: "v2.0.0",
			filters: nil,
			found:   false,
		},
	}

	for _, testItem := range testData {
		t.Run(testItem.name, func(t *testing.T) {
			t.Parallel()

			updater, err := selfupdate.NewUpdater(selfupdate.Config{
				Source:        &memory.Source{},
				Filters:       testItem.filters,
				OS:            testItem.os,
				Arch:          testItem.arch,
				Arm:           testItem.arm,
				UniversalArch: testItem.universalArch,
			})
			require.NoError(t, err)
			_, asset, _, found := updater.FindReleaseAndAsset(testItem.releases, testItem.version)
			assert.Equal(t, testItem.found, found)
			if found {
				assert.Equal(t, testItem.expectedAssetName, asset.GetName())
			}
		})
	}
}

func TestBuildMultistepValidationChain(t *testing.T) {
	testVersion := "v0.14.0"
	source := mockSignedSourceRepository(t)
	checksumValidator := &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"}

	t.Run("ValidConfig", func(t *testing.T) {
		updater, _ := selfupdate.NewUpdater(selfupdate.Config{
			Source:    source,
			Validator: newChecksumWithSignatureValidator("checksums.txt"),
		})

		release, err := updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, testVersion})
		require.NotNil(t, release)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(release.ValidationChain))
		assert.Equal(t, "checksums.txt", release.ValidationChain[0].ValidationAssetName)
		assert.Equal(t, "checksums.txt.asc", release.ValidationChain[1].ValidationAssetName)
	})

	t.Run("LoopConfig", func(t *testing.T) {
		updater, _ := selfupdate.NewUpdater(selfupdate.Config{
			Source: source,
			Validator: new(selfupdate.PatternValidator).
				Add("*", checksumValidator),
		})

		_, err := updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, testVersion})
		assert.NoError(t, err)
	})

	t.Run("InvalidLoopConfig", func(t *testing.T) {
		updater, _ := selfupdate.NewUpdater(selfupdate.Config{
			Source: source,
			Validator: new(selfupdate.PatternValidator).
				Add("*.*z*", checksumValidator).
				Add("*", new(selfupdate.SHAValidator)),
		})

		_, err := updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{testGithubRepository, testVersion})
		assert.EqualError(t, err, "validation file not found: \"checksums.txt.sha256\"")
	})
}

func newMockUpdater(t *testing.T, config selfupdate.Config) *selfupdate.Updater {
	t.Helper()

	if config.Source == nil {
		config.Source = mockMemorySource(t)
	}
	updater, err := selfupdate.NewUpdater(config)
	require.NoError(t, err)
	return updater
}

// newGitHubUpdater creates an updater reading from the real GitHub API, for tests not running in short mode
func newGitHubUpdater(t *testing.T, config selfupdate.Config) *selfupdate.Updater {
	t.Helper()

	source, err := github.NewSource(github.Config{})
	require.NoError(t, err)
	config.Source = source
	updater, err := selfupdate.NewUpdater(config)
	require.NoError(t, err)
	return updater
}
