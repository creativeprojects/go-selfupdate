package selfupdate

import (
	"context"
	"io"
	"time"
)

// Source interface to load the releases from
type Source interface {
	// ListReleases returns all releases for the given repository.
	ListReleases(ctx context.Context, repository Repository) ([]SourceRelease, error)
	// GetLatestRelease returns the latest release for the given repository:
	// ErrReleaseNotFound when the repository has no release, ErrRepositoryNotFound when it does not exist.
	GetLatestRelease(ctx context.Context, repository Repository) (SourceRelease, error)
	// DownloadReleaseAsset downloads the asset with the given ID for the specified release.
	DownloadReleaseAsset(ctx context.Context, rel *Release, assetID int64) (io.ReadCloser, error)
}

type SourceRelease interface {
	GetID() int64
	GetTagName() string
	GetDraft() bool
	GetPrerelease() bool
	GetPublishedAt() time.Time
	GetReleaseNotes() string
	GetName() string
	GetURL() string
	GetAssets() []SourceAsset
}

type SourceAsset interface {
	GetID() int64
	GetName() string
	GetSize() int
	GetBrowserDownloadURL() string
}
