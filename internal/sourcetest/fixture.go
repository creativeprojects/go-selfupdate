package sourcetest

import (
	"fmt"
	"strings"
	"time"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
)

// Repository is the only repository served by the data returned by Fixture.
var Repository = selfupdate.NewRepositorySlug("owner", "repo")

// Fixture returns the reference data a source under test is expected to serve:
// four releases, including a pre-release and a draft, with three assets each.
// Asset URLs are placeholders: the fake servers serve the content from their own URL.
func Fixture() *memory.Source {
	files := make(map[int64][]byte)
	releases := []selfupdate.SourceRelease{
		newFixtureRelease(files, 1, "v1.0.0", "first release", time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC), false, false),
		newFixtureRelease(files, 2, "v1.1.0-rc1", "release candidate", time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC), false, true),
		newFixtureRelease(files, 3, "v1.1.0", "stable release", time.Date(2021, 2, 1, 12, 0, 0, 0, time.UTC), false, false),
		newFixtureRelease(files, 4, "v2.0.0", "work in progress", time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC), true, false),
	}
	return &memory.Source{
		Repository:    Repository,
		LatestRelease: releases[2],
		Releases:      releases,
		Files:         files,
	}
}

// newFixtureRelease creates a release with a linux archive, a windows archive and a checksum file,
// and stores their content in files.
func newFixtureRelease(files map[int64][]byte, id int64, tag, notes string, publishedAt time.Time, draft, prerelease bool) *memory.Release {
	version := strings.TrimPrefix(tag, "v")
	names := []string{
		fmt.Sprintf("app_%s_linux_amd64.tar.gz", version),
		fmt.Sprintf("app_%s_windows_amd64.zip", version),
		"checksums.txt",
	}
	assets := make([]selfupdate.SourceAsset, len(names))
	for i, name := range names {
		assetID := id*10 + int64(i) + 1
		content := []byte(fmt.Sprintf("%s: content of %s\n", tag, name))
		files[assetID] = content
		assets[i] = &memory.Asset{
			ID:   assetID,
			Name: name,
			Size: len(content),
			URL:  fmt.Sprintf("https://example.com/owner/repo/releases/download/%s/%s", tag, name),
		}
	}
	return &memory.Release{
		ReleaseID:    id,
		Name:         "Release " + tag,
		TagName:      tag,
		URL:          "https://example.com/owner/repo/releases/tag/" + tag,
		Draft:        draft,
		Prerelease:   prerelease,
		PublishedAt:  publishedAt,
		ReleaseNotes: notes,
		Assets:       assets,
	}
}
