package selfupdate

import (
	"fmt"
	stdlog "log"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func skipRateLimitExceeded(t *testing.T, err error) {
	if err == nil {
		return
	}
	if strings.Contains(err.Error(), "403 API rate limit") {
		t.Skip("Test skipped because of GitHub API rate limit exceeded")
		runtime.Goexit()
	}
}

func TestDetectReleaseWithVersionPrefix(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: mockSourceRepository(t)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, ok, err := testItem.updater.DetectLatest("creativeprojects/resticprofile")
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.True(t, ok, "Failed to detect latest")
			assert.NotNil(t, r, "No release returned")
			if r.LessThan("0.10.0") {
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
	testVersion := "v0.10.0"
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: mockSourceRepository(t)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, ok, err := testItem.updater.DetectVersion("creativeprojects/resticprofile", testVersion)
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.Truef(t, ok, "Failed to detect %s", testVersion)
			assert.NotNil(t, r, "No release returned")
		})
	}
}

func TestDetectVersionNotExisting(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: mockSourceRepository(t)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			r, ok, err := testItem.updater.DetectVersion("creativeprojects/resticprofile", "foobar")
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.False(t, ok, "Failed to correctly detect foobar")
			assert.Nil(t, r, "Release not detected but got a returned value for it")
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
			r, ok, err := DetectLatest(tc.slug)
			skipRateLimitExceeded(t, err)
			if err != nil {
				t.Fatal("Fetch failed:", err)
			}
			if !ok {
				t.Fatal(tc.slug, "not found")
			}
			if r == nil {
				t.Fatal("Release not detected")
			}
			if !r.Equal("1.2.3") {
				t.Error("")
			}
			url := fmt.Sprintf("https://github.com/%s/releases/tag/%s1.2.3", tc.slug, tc.prefix)
			if r.URL != url {
				t.Error("URL is not correct. Want", url, "but got", r.URL)
			}
			if r.ReleaseNotes == "" {
				t.Error("Release note is unexpectedly empty")
			}
			if !strings.HasPrefix(r.AssetURL, fmt.Sprintf("https://github.com/%s/releases/download/%s1.2.3/", tc.slug, tc.prefix)) {
				t.Error("Unexpected asset URL:", r.AssetURL)
			}
			if r.Name == "" {
				t.Error("Release name is unexpectedly empty")
			}
			if r.AssetByteSize == 0 {
				t.Error("Asset's size is unexpectedly zero")
			}
			if r.AssetID == 0 {
				t.Error("Asset's ID is unexpectedly zero")
			}
			if r.PublishedAt.IsZero() {
				t.Error("Release time is unexpectedly zero")
			}
		})
	}
}

func TestDetectReleaseButNoAsset(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: NewMockSource(
			[]SourceRelease{
				&GitHubRelease{
					name:    "first",
					tagName: "v1.0",
					assets:  nil,
				},
				&GitHubRelease{
					name:    "second",
					tagName: "v2.0",
					assets:  nil,
				},
			},
			nil,
		)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			_, ok, err := testItem.updater.DetectLatest("rhysd/clever-f.vim")
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.False(t, ok, "When no asset found, result should be marked as 'not found'")
		})
	}
}

func TestInvalidSlug(t *testing.T) {
	up := DefaultUpdater()

	for _, slug := range []string{
		"foo",
		"/",
		"foo/",
		"/bar",
		"foo/bar/piyo",
	} {
		_, _, err := up.DetectLatest(slug)
		assert.EqualError(t, err, ErrInvalidSlug.Error())
	}
}

func TestNonExistingRepo(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: NewMockSource(nil, nil)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			_, ok, err := testItem.updater.DetectLatest("rhysd/non-existing-repo")
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.False(t, ok, "Release for non-existing repo should not be found")
		})
	}
}

func TestNoReleaseFound(t *testing.T) {
	testData := []struct {
		run     bool
		name    string
		updater *Updater
	}{
		{true, "Mock", newMockUpdater(t, Config{Source: NewMockSource(nil, nil)})},
		{!testing.Short(), "GitHub", DefaultUpdater()},
	}

	for _, testItem := range testData {
		if !testItem.run {
			continue
		}
		t.Run(testItem.name, func(t *testing.T) {
			_, ok, err := testItem.updater.DetectLatest("rhysd/misc")
			skipRateLimitExceeded(t, err)
			require.NoError(t, err)
			assert.False(t, ok, "Repo having no release should not be found")
		})
	}
}

func TestFindAssetFromRelease(t *testing.T) {
	type findReleaseAndAssetFixture struct {
		name            string
		config          Config
		release         SourceRelease
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
			config:        Config{},
			release:       nil,
			targetVersion: "",
			expectedFound: false,
		},
		{
			name: "find asset, no filters",
			release: &GitHubRelease{
				name:    rel1,
				tagName: v1,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url1},
				},
			},
			targetVersion:   "1.0.0",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "find asset, no target version",
			release: &GitHubRelease{
				name:    rel1,
				tagName: v1,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url1},
				},
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "don't find prerelease",
			release: &GitHubRelease{
				name:    rel1,
				tagName: v1,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url1},
				},
				prerelease: true,
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   false,
		},
		{
			name: "find named prerelease",
			release: &GitHubRelease{
				name:    rel1,
				tagName: v1,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url1},
				},
				prerelease: true,
			},
			targetVersion:   "1.0.0",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name:   "find prerelease",
			config: Config{Prerelease: true},
			release: &GitHubRelease{
				name:    rel1,
				tagName: v1,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url1},
				},
				prerelease: true,
			},
			targetVersion:   "",
			expectedAsset:   asset1,
			expectedVersion: "1.0.0",
			expectedFound:   true,
		},
		{
			name: "don't find asset with wrong extension, no filters",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: wrongAsset1, url: url11},
				},
			},
			targetVersion: "1.1.0",
			expectedFound: false,
		},
		{
			name: "find asset with different name, no filters",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset1, url: url11},
				},
			},
			targetVersion:   "1.1.0",
			expectedAsset:   asset1,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, no filters (2)",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset11, url: url11},
				},
			},
			targetVersion:   "1.1.0",
			expectedAsset:   asset11,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match filter",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset11, url: url11},
					&GitHubAsset{name: asset1, url: url1},
				},
			},
			targetVersion:   "1.1.0",
			config:          Config{Filters: []string{"11"}},
			expectedAsset:   asset11,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match another filter",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset11, url: url11},
					&GitHubAsset{name: asset1, url: url1},
				},
			},
			targetVersion:   "1.1.0",
			config:          Config{Filters: []string{"([^1])1{1}([^1])"}},
			expectedAsset:   asset1,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match any filter",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset11, url: url11},
					&GitHubAsset{name: asset2, url: url2},
				},
			},
			targetVersion:   "1.1.0",
			config:          Config{Filters: []string{"([^1])1{1}([^1])", "([^1])2{1}([^1])"}},
			expectedAsset:   asset2,
			expectedVersion: "1.1.0",
			expectedFound:   true,
		},
		{
			name: "find asset, match no filter",
			release: &GitHubRelease{
				name:    rel11,
				tagName: v11,
				assets: []SourceAsset{
					&GitHubAsset{name: asset11, url: url11},
					&GitHubAsset{name: asset2, url: url2},
				},
			},
			targetVersion: "",
			config:        Config{Filters: []string{"another", "binary"}},
			expectedFound: false,
		},
	}

	for _, fixture := range testData {
		t.Run(fixture.name, func(t *testing.T) {
			updater := newMockUpdater(t, fixture.config)
			asset, ver, found := updater.findAssetFromRelease(fixture.release, []string{".gz"}, fixture.targetVersion)
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
	SetLogger(stdlog.New(os.Stderr, "", 0))
	defer SetLogger(&emptyLogger{})

	// stupid library using pointer to strings everywhere
	tag2 := "v2.0.0"
	rel2 := "rel2"
	assetLinuxI32 := "asset_linux_386.tgz"
	assetLinuxI64 := "asset_linux_amd64.tgz"
	assetLinuxARM := "asset_linux_arm.tgz"
	assetLinuxARMv5 := "asset_linux_armv5.tgz"
	assetLinuxARMv6 := "asset_linux_armv6.tgz"
	assetLinuxARMv7 := "asset_linux_armv7.tgz"
	assetLinuxARM64 := "asset_linux_arm64.tgz"
	testData := []struct {
		name              string
		os                string
		arch              string
		arm               uint8
		releases          []SourceRelease
		version           string
		filters           []string
		found             bool
		expectedAssetName *string
	}{
		{
			name: "no match",
			os:   "darwin",
			arch: "amd64",
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxI32,
						},
						&GitHubAsset{
							name: assetLinuxI64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             false,
			expectedAssetName: &assetLinuxI64,
		},
		{
			name: "simple match",
			os:   "linux",
			arch: "amd64",
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxI32,
						},
						&GitHubAsset{
							name: assetLinuxI64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: &assetLinuxI64,
		},
		{
			name: "match default arm",
			os:   "linux",
			arch: "arm",
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxARM,
						},
						&GitHubAsset{
							name: assetLinuxARM64,
						},
						&GitHubAsset{
							name: assetLinuxARMv5,
						},
						&GitHubAsset{
							name: assetLinuxARMv6,
						},
						&GitHubAsset{
							name: assetLinuxARMv7,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: &assetLinuxARM,
		},
		{
			name: "match armv6",
			os:   "linux",
			arch: "arm",
			arm:  6,
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxARM,
						},
						&GitHubAsset{
							name: assetLinuxARM64,
						},
						&GitHubAsset{
							name: assetLinuxARMv5,
						},
						&GitHubAsset{
							name: assetLinuxARMv6,
						},
						&GitHubAsset{
							name: assetLinuxARMv7,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: &assetLinuxARMv6,
		},
		{
			name: "fallback to armv5",
			os:   "linux",
			arch: "arm",
			arm:  7,
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxARM,
						},
						&GitHubAsset{
							name: assetLinuxARM64,
						},
						&GitHubAsset{
							name: assetLinuxARMv5,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: &assetLinuxARMv5,
		},
		{
			name: "fallback to arm",
			os:   "linux",
			arch: "arm",
			arm:  5,
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxARM,
						},
						&GitHubAsset{
							name: assetLinuxARM64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             true,
			expectedAssetName: &assetLinuxARM,
		},
		{
			name: "arm not found",
			os:   "linux",
			arch: "arm",
			arm:  6,
			releases: []SourceRelease{
				&GitHubRelease{
					name:    rel2,
					tagName: tag2,
					assets: []SourceAsset{
						&GitHubAsset{
							name: assetLinuxARMv7,
						},
						&GitHubAsset{
							name: assetLinuxARM64,
						},
					},
				},
			},
			version:           "v2.0.0",
			filters:           nil,
			found:             false,
			expectedAssetName: &assetLinuxARM,
		},
	}

	for _, testItem := range testData {
		t.Run(testItem.name, func(t *testing.T) {
			updater, err := NewUpdater(Config{
				Filters: testItem.filters,
				OS:      testItem.os,
				Arch:    testItem.arch,
				Arm:     testItem.arm,
			})
			require.NoError(t, err)
			_, asset, _, found := updater.findReleaseAndAsset(testItem.releases, testItem.version)
			assert.Equal(t, testItem.found, found)
			if found {
				assert.Equal(t, *testItem.expectedAssetName, asset.GetName())
			}
		})
	}
}

func newMockUpdater(t *testing.T, config Config) *Updater {
	updater, err := NewUpdater(config)
	require.NoError(t, err)
	return updater
}
