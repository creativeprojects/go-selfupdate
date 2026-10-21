package github

import (
	"time"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/google/go-github/v86/github"
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
	assets       []selfupdate.SourceAsset
}

func NewRelease(from *github.RepositoryRelease) *Release {
	release := &Release{
		releaseID:    from.GetID(),
		name:         from.GetName(),
		tagName:      from.GetTagName(),
		url:          from.GetHTMLURL(),
		publishedAt:  from.GetPublishedAt().Time,
		releaseNotes: from.GetBody(),
		draft:        from.GetDraft(),
		prerelease:   from.GetPrerelease(),
		assets:       make([]selfupdate.SourceAsset, len(from.Assets)),
	}
	for i, fromAsset := range from.Assets {
		release.assets[i] = NewGitHubAsset(fromAsset)
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

func (r *Release) GetAssets() []selfupdate.SourceAsset {
	return r.assets
}

type Asset struct {
	id   int64
	name string
	size int
	url  string
}

func NewGitHubAsset(from *github.ReleaseAsset) *Asset {
	return &Asset{
		id:   from.GetID(),
		name: from.GetName(),
		size: from.GetSize(),
		url:  from.GetBrowserDownloadURL(),
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
	_ selfupdate.SourceRelease = &Release{}
	_ selfupdate.SourceAsset   = &Asset{}
)
