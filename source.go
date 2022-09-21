package selfupdate

import (
	"context"
	"io"
	"time"
)

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

// Source interface to load the releases from (GitHubSource for example)
type Source interface {
	ListReleases(ctx context.Context, owner, repo string) ([]SourceRelease, error)
	DownloadReleaseAsset(ctx context.Context, owner, repo string, releaseID, id int64) (io.ReadCloser, error)
}

// checkOwnerRepoParameters is a helper function to check both parameters are valid
func checkOwnerRepoParameters(owner, repo string) error {
	if owner == "" {
		return ErrIncorrectParameterOwner
	}
	if repo == "" {
		return ErrIncorrectParameterRepo
	}
	return nil
}
