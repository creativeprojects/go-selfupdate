package selfupdate

import (
	"io"
	"time"
)

type SourceRelease interface {
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

// Source interface to load the releases from (GitHubSource for example)
type Source interface {
	ListReleases(owner, repo string) ([]SourceRelease, error)
	DownloadReleaseAsset(owner, repo string, id int64) (io.ReadCloser, error)
}
