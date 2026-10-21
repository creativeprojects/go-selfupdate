package gitlab

import (
	"time"

	"github.com/creativeprojects/go-selfupdate/v2"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Release struct {
	releaseID   int64
	name        string
	tagName     string
	url         string
	publishedAt time.Time
	description string
	assets      []selfupdate.SourceAsset
}

func NewRelease(from *gitlab.Release) *Release {
	release := &Release{
		releaseID:   0,
		name:        from.Name,
		tagName:     from.TagName,
		url:         from.Commit.WebURL,
		publishedAt: *from.ReleasedAt,
		description: from.Description,
		assets:      make([]selfupdate.SourceAsset, len(from.Assets.Links)),
	}
	for i, fromLink := range from.Assets.Links {
		release.assets[i] = NewAsset(fromLink)
	}
	return release
}

func (r *Release) GetID() int64 {
	return 0
}

func (r *Release) GetTagName() string {
	return r.tagName
}

func (r *Release) GetDraft() bool {
	return false
}

func (r *Release) GetPrerelease() bool {
	return false
}

func (r *Release) GetPublishedAt() time.Time {
	return r.publishedAt
}

func (r *Release) GetReleaseNotes() string {
	return r.description
}

func (r *Release) GetName() string {
	return r.name
}

func (r *Release) GetURL() string {
	return r.url
}

func (r *Release) GetAssets() []selfupdate.SourceAsset {
	return r.assets
}

type Asset struct {
	id   int64
	name string
	url  string
}

func NewAsset(from *gitlab.ReleaseLink) *Asset {
	return &Asset{
		id:   from.ID,
		name: from.Name,
		url:  from.URL,
	}
}

func (a *Asset) GetID() int64 {
	return a.id
}

func (a *Asset) GetName() string {
	return a.name
}

func (a *Asset) GetSize() int {
	return 0
}

func (a *Asset) GetBrowserDownloadURL() string {
	return a.url
}

// Verify interface
var (
	_ selfupdate.SourceRelease = &Release{}
	_ selfupdate.SourceAsset   = &Asset{}
)
