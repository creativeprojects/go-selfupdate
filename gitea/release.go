package gitea

import (
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/creativeprojects/go-selfupdate/types"
)

type Release struct {
	releaseID    int64
	name         string
	tagName      string
	url          string
	draft        bool
	prerelease   bool
	publishedAt  time.Time
	releaseNotes string
	assets       []types.SourceAsset
}

func NewRelease(from *gitea.Release) *Release {
	release := &Release{
		releaseID:    from.ID,
		name:         from.Title,
		tagName:      from.TagName,
		url:          "", //FIXME: we kind of have no url ?
		publishedAt:  from.PublishedAt,
		releaseNotes: from.Note,
		draft:        from.IsDraft,
		prerelease:   from.IsPrerelease,
		assets:       make([]types.SourceAsset, len(from.Attachments)),
	}

	for i, fromAsset := range from.Attachments {
		release.assets[i] = NewAsset(fromAsset)
	}

	return release
}

func (r *Release) GetID() int64 {
	return r.releaseID
}

func (r *Release) GetTagName() string {
	return r.tagName
}

func (r *Release) GetDraft() bool {
	return r.draft
}

func (r *Release) GetPrerelease() bool {
	return r.prerelease
}

func (r *Release) GetPublishedAt() time.Time {
	return r.publishedAt
}

func (r *Release) GetReleaseNotes() string {
	return r.releaseNotes
}

func (r *Release) GetName() string {
	return r.name
}

func (r *Release) GetURL() string {
	return r.url
}

func (r *Release) GetAssets() []types.SourceAsset {
	return r.assets
}

type Asset struct {
	id   int64
	name string
	size int
	url  string
}

func NewAsset(from *gitea.Attachment) *Asset {
	return &Asset{
		id:   from.ID,
		name: from.Name,
		size: int(from.Size),
		url:  from.DownloadURL,
	}
}

func (a *Asset) GetID() int64 {
	return a.id
}

func (a *Asset) GetName() string {
	return a.name
}

func (a *Asset) GetSize() int {
	return a.size
}

func (a *Asset) GetBrowserDownloadURL() string {
	return a.url
}

// Verify interface
var (
	_ types.SourceRelease = &Release{}
	_ types.SourceAsset   = &Asset{}
)
