package memory

import (
	"time"

	"github.com/creativeprojects/go-selfupdate/v2"
)

type Release struct {
	ReleaseID    int64
	Name         string
	TagName      string
	URL          string
	Draft        bool
	Prerelease   bool
	PublishedAt  time.Time
	ReleaseNotes string
	Assets       []selfupdate.SourceAsset
}

func (r *Release) GetID() int64 {
	return r.ReleaseID
}

func (r *Release) GetTagName() string {
	return r.TagName
}

func (r *Release) GetDraft() bool {
	return r.Draft
}

func (r *Release) GetPrerelease() bool {
	return r.Prerelease
}

func (r *Release) GetPublishedAt() time.Time {
	return r.PublishedAt
}

func (r *Release) GetReleaseNotes() string {
	return r.ReleaseNotes
}

func (r *Release) GetName() string {
	return r.Name
}

func (r *Release) GetURL() string {
	return r.URL
}

func (r *Release) GetAssets() []selfupdate.SourceAsset {
	return r.Assets
}

type Asset struct {
	ID   int64
	Name string
	Size int
	URL  string
}

func (a *Asset) GetID() int64 {
	return a.ID
}

func (a *Asset) GetName() string {
	return a.Name
}

func (a *Asset) GetSize() int {
	return a.Size
}

func (a *Asset) GetBrowserDownloadURL() string {
	return a.URL
}

// Verify interface
var (
	_ selfupdate.SourceRelease = &Release{}
	_ selfupdate.SourceAsset   = &Asset{}
)
