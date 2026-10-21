package selfupdate_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/memory"
	"github.com/stretchr/testify/require"
)

// mockMemorySource creates a new *memory.Source pre-populated with different versions and assets
func mockMemorySource(t *testing.T) *memory.Source {

	gzData, err := os.ReadFile("testdata/new_version.tar.gz")
	require.NoError(t, err)

	zipData, err := os.ReadFile("testdata/new_version.zip")
	require.NoError(t, err)

	releases := []selfupdate.SourceRelease{
		&memory.Release{
			Name:         "v0.1.0",
			TagName:      "v0.1.0",
			URL:          "v0.1.0",
			Prerelease:   true,
			PublishedAt:  time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "first stable",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   1,
					Name: "resticprofile_0.1.0_linux_amd64.tar.gz",
					URL:  "resticprofile_0.1.0_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   2,
					Name: "resticprofile_0.1.0_darwin_amd64.tar.gz",
					URL:  "resticprofile_0.1.0_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   3,
					Name: "resticprofile_0.1.0_windows_amd64.zip",
					URL:  "resticprofile_0.1.0_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   31,
					Name: "resticprofile_0.1.0_darwin_arm64.tar.gz",
					URL:  "resticprofile_0.1.0_darwin_arm64.tar.gz",
					Size: len(gzData),
				},
			},
		},
		&memory.Release{
			Name:         "v0.14.0",
			TagName:      "v0.14.0",
			URL:          "v0.14.0",
			Prerelease:   false,
			PublishedAt:  time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "latest stable",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   4,
					Name: "resticprofile_0.14.0_linux_amd64.tar.gz",
					URL:  "resticprofile_0.14.0_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   5,
					Name: "resticprofile_0.14.0_darwin_amd64.tar.gz",
					URL:  "resticprofile_0.14.0_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   6,
					Name: "resticprofile_0.14.0_windows_amd64.zip",
					URL:  "resticprofile_0.14.0_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   32,
					Name: "resticprofile_0.14.0_darwin_arm64.tar.gz",
					URL:  "resticprofile_0.14.0_darwin_arm64.tar.gz",
					Size: len(gzData),
				},
			},
		},
		&memory.Release{
			Name:         "v1.0.0-rc",
			TagName:      "v1.0.0-rc",
			URL:          "v1.0.0-rc",
			Prerelease:   false,
			PublishedAt:  time.Date(2011, 1, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "release candidate",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   11,
					Name: "resticprofile_1.0.0-rc_linux_amd64.tar.gz",
					URL:  "resticprofile_1.0.0-rc_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   12,
					Name: "resticprofile_1.0.0-rc_darwin_amd64.tar.gz",
					URL:  "resticprofile_1.0.0-rc_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   13,
					Name: "resticprofile_1.0.0-rc_windows_amd64.zip",
					URL:  "resticprofile_1.0.0-rc_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   33,
					Name: "resticprofile_1.0.0-rc_darwin_arm64.tar.gz",
					URL:  "resticprofile_1.0.0-rc_darwin_arm64.tar.gz",
					Size: len(gzData),
				},
			},
		},
		&memory.Release{
			Name:         "v1.0.0",
			TagName:      "v1.0.0",
			URL:          "v1.0.0",
			Prerelease:   false,
			PublishedAt:  time.Date(2011, 2, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "final v1",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   14,
					Name: "resticprofile_1.0.0_linux_amd64.tar.gz",
					URL:  "resticprofile_1.0.0_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   15,
					Name: "resticprofile_1.0.0_darwin_amd64.tar.gz",
					URL:  "resticprofile_1.0.0_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   16,
					Name: "resticprofile_1.0.0_windows_amd64.zip",
					URL:  "resticprofile_1.0.0_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   34,
					Name: "resticprofile_1.0.0_darwin_arm64.tar.gz",
					URL:  "resticprofile_1.0.0_darwin_arm64.tar.gz",
					Size: len(gzData),
				},
			},
		},
		&memory.Release{
			Name:         "v2.0.0-beta",
			TagName:      "v2.0.0-beta",
			URL:          "v2.0.0-beta",
			Prerelease:   true,
			PublishedAt:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "beta",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   21,
					Name: "resticprofile_2.0.0-beta_linux_amd64.tar.gz",
					URL:  "resticprofile_2.0.0-beta_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   22,
					Name: "resticprofile_2.0.0-beta_darwin_amd64.tar.gz",
					URL:  "resticprofile_2.0.0-beta_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   23,
					Name: "resticprofile_2.0.0-beta_windows_amd64.zip",
					URL:  "resticprofile_2.0.0-beta_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   35,
					Name: "resticprofile_2.0.0-beta_darwin_arm64.tar.gz",
					URL:  "resticprofile_2.0.0-beta_darwin_arm64.tar.gz",
					Size: len(gzData),
				},
			},
		},
		&memory.Release{
			Name:         "v2.0.0",
			TagName:      "v2.0.0",
			URL:          "v2.0.0",
			Draft:        true,
			PublishedAt:  time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			ReleaseNotes: "almost there",
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   24,
					Name: "resticprofile_2.0.0_linux_amd64.tar.gz",
					URL:  "resticprofile_2.0.0_linux_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   25,
					Name: "resticprofile_2.0.0_darwin_amd64.tar.gz",
					URL:  "resticprofile_2.0.0_darwin_amd64.tar.gz",
					Size: len(gzData),
				},
				&memory.Asset{
					ID:   26,
					Name: "resticprofile_2.0.0_windows_amd64.zip",
					URL:  "resticprofile_2.0.0_windows_amd64.zip",
					Size: len(zipData),
				},
				&memory.Asset{
					ID:   36,
					Name: "resticprofile_2.0.0_darwin_arm64.tar.gz",
					URL:  "resticprofile_2.0.0_darwin_arm64.tar.gz",
					Size: len(gzData),
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
		31: gzData,
		32: gzData,
		33: gzData,
		34: gzData,
		35: gzData,
		36: gzData,
	}

	// generates checksum files automatically
	for i, release := range releases {
		rel := release.(*memory.Release)
		checksums := &bytes.Buffer{}
		for _, asset := range rel.Assets {
			file, ok := files[asset.GetID()]
			if !ok {
				t.Errorf("file ID %d not found", asset.GetID())
			}
			hash := sha256.Sum256(file)
			fmt.Fprintf(checksums, "%x  %s\n", hash, asset.GetName())
		}
		id := int64(i*10 + 101)
		rel.Assets = append(rel.Assets, &memory.Asset{
			ID:   id,
			Name: "checksums.txt",
		})
		files[id] = checksums.Bytes()
		// t.Logf("file id %d contains checksums:\n%s\n", id, string(files[id]))
	}

	return &memory.Source{Releases: releases, Files: files}
}

// mockSignedSourceRepository creates a variant of mockSourceRepository where "checksums.txt" is "signed":
// "checksums.txt.asc" holds the SHA256 of "checksums.txt" and is checked by signatureValidator.
func mockSignedSourceRepository(t *testing.T) *memory.Source {
	source := mockMemorySource(t)

	for i, release := range source.Releases {
		rel := release.(*memory.Release)

		id := int64(i*10 + 101)
		signatureID := id + 1
		shaSums := source.Files[id]

		rel.Assets = append(rel.Assets, &memory.Asset{
			ID:   signatureID,
			Name: "checksums.txt.asc",
		})
		source.Files[signatureID] = []byte(fmt.Sprintf("%x", sha256.Sum256(shaSums)))
	}

	return source
}

// signatureValidator stands in for a signature validator in multistep
// validation tests: it checks a SHA256 hash stored in a ".asc" file.
type signatureValidator struct {
	selfupdate.SHAValidator
}

func (v *signatureValidator) GetValidationAssetName(releaseFilename string) string {
	return releaseFilename + ".asc"
}

// newChecksumWithSignatureValidator mirrors pgp.NewChecksumValidator using signatureValidator
func newChecksumWithSignatureValidator(checksumsFilename string) selfupdate.Validator {
	return new(selfupdate.PatternValidator).
		Add(checksumsFilename, new(signatureValidator)).
		Add("*", &selfupdate.ChecksumValidator{UniqueFilename: checksumsFilename}).
		SkipValidation("*.asc")
}

// newMemorySource creates a *memory.Source from a list of releases and their files
func newMemorySource(releases []selfupdate.SourceRelease, files map[int64][]byte) *memory.Source {
	return &memory.Source{
		Releases: releases,
		Files:    files,
	}
}
