package selfupdate

import (
	"context"
	"io"
	"time"
)

// Source interface to load the releases from (GitHubSource for example)
type Source interface {
	ListReleases(ctx context.Context, repository Repository) ([]SourceRelease, error)
	LatestRelease(ctx context.Context, repository Repository) (SourceRelease, error)
	DownloadReleaseAsset(ctx context.Context, repository Repository, releaseID, id int64) (io.ReadCloser, error)
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
