package selfupdate

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// mockSourceRepository creates a new *MockSource pre-populated with different versions and assets
func mockSourceRepository(t *testing.T) *MockSource {

	gzData, err := os.ReadFile("testdata/new_version.tar.gz")
	require.NoError(t, err)

	zipData, err := os.ReadFile("testdata/new_version.zip")
	require.NoError(t, err)

	releases := []SourceRelease{
		&GitHubRelease{
			name:         "v0.1.0",
			tagName:      "v0.1.0",
			url:          "v0.1.0",
			prerelease:   true,
			publishedAt:  time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "first stable",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   1,
					name: "resticprofile_0.1.0_linux_amd64.tar.gz",
					url:  "resticprofile_0.1.0_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   2,
					name: "resticprofile_0.1.0_darwin_amd64.tar.gz",
					url:  "resticprofile_0.1.0_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   3,
					name: "resticprofile_0.1.0_windows_amd64.zip",
					url:  "resticprofile_0.1.0_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
		&GitHubRelease{
			name:         "v0.10.0",
			tagName:      "v0.10.0",
			url:          "v0.10.0",
			prerelease:   false,
			publishedAt:  time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "latest stable",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   4,
					name: "resticprofile_0.10.0_linux_amd64.tar.gz",
					url:  "resticprofile_0.10.0_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   5,
					name: "resticprofile_0.10.0_darwin_amd64.tar.gz",
					url:  "resticprofile_0.10.0_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   6,
					name: "resticprofile_0.10.0_windows_amd64.zip",
					url:  "resticprofile_0.10.0_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
		&GitHubRelease{
			name:         "v1.0.0-rc",
			tagName:      "v1.0.0-rc",
			url:          "v1.0.0-rc",
			prerelease:   false,
			publishedAt:  time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "release candidate",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   11,
					name: "resticprofile_1.0.0-rc_linux_amd64.tar.gz",
					url:  "resticprofile_1.0.0-rc_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   12,
					name: "resticprofile_1.0.0-rc_darwin_amd64.tar.gz",
					url:  "resticprofile_1.0.0-rc_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   13,
					name: "resticprofile_1.0.0-rc_windows_amd64.zip",
					url:  "resticprofile_1.0.0-rc_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
		&GitHubRelease{
			name:         "v1.0.0",
			tagName:      "v1.0.0",
			url:          "v1.0.0",
			prerelease:   false,
			publishedAt:  time.Date(2011, 2, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "final v1",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   14,
					name: "resticprofile_1.0.0_linux_amd64.tar.gz",
					url:  "resticprofile_1.0.0_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   15,
					name: "resticprofile_1.0.0_darwin_amd64.tar.gz",
					url:  "resticprofile_1.0.0_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   16,
					name: "resticprofile_1.0.0_windows_amd64.zip",
					url:  "resticprofile_1.0.0_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
		&GitHubRelease{
			name:         "v2.0.0-beta",
			tagName:      "v2.0.0-beta",
			url:          "v2.0.0-beta",
			prerelease:   true,
			publishedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "beta",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   21,
					name: "resticprofile_2.0.0-beta_linux_amd64.tar.gz",
					url:  "resticprofile_2.0.0-beta_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   22,
					name: "resticprofile_2.0.0-beta_darwin_amd64.tar.gz",
					url:  "resticprofile_2.0.0-beta_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   23,
					name: "resticprofile_2.0.0-beta_windows_amd64.zip",
					url:  "resticprofile_2.0.0-beta_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
		&GitHubRelease{
			name:         "v2.0.0",
			tagName:      "v2.0.0",
			url:          "v2.0.0",
			draft:        true,
			publishedAt:  time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			releaseNotes: "almost there",
			assets: []SourceAsset{
				&GitHubAsset{
					id:   24,
					name: "resticprofile_2.0.0_linux_amd64.tar.gz",
					url:  "resticprofile_2.0.0_linux_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   25,
					name: "resticprofile_2.0.0_darwin_amd64.tar.gz",
					url:  "resticprofile_2.0.0_darwin_amd64.tar.gz",
					size: len(gzData),
				},
				&GitHubAsset{
					id:   26,
					name: "resticprofile_2.0.0_windows_amd64.zip",
					url:  "resticprofile_2.0.0_windows_amd64.zip",
					size: len(zipData),
				},
			},
		},
	}

	files := map[int64][]byte{
		1:  gzData,
		2:  gzData,
		3:  zipData,
		4:  gzData,
		5:  gzData,
		6:  zipData,
		11: gzData,
		12: gzData,
		13: zipData,
		14: gzData,
		15: gzData,
		16: zipData,
		21: gzData,
		22: gzData,
		23: zipData,
		24: gzData,
		25: gzData,
		26: zipData,
	}

	// generates checksum files automatically
	for i, release := range releases {
		rel := release.(*GitHubRelease)
		checksums := &bytes.Buffer{}
		for _, asset := range rel.assets {
			file, ok := files[asset.GetID()]
			if !ok {
				t.Errorf("file ID %d not found", asset.GetID())
			}
			hash := sha256.Sum256(file)
			checksums.WriteString(fmt.Sprintf("%x  %s\n", hash, asset.GetName()))
		}
		id := int64(i*10 + 101)
		rel.assets = append(rel.assets, &GitHubAsset{
			id:   id,
			name: "checksums.txt",
		})
		files[id] = checksums.Bytes()
		t.Logf("file id %d contains checksums:\n%s\n", id, string(files[id]))
	}

	return NewMockSource(releases, files)
}
