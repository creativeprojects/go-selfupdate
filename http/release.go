// Copyright (c) 2024 Mr. Gecko's Media (James Coleman). http://mrgeckosmedia.com/
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package http

import (
	"time"

	"github.com/creativeprojects/go-selfupdate/types"
)

type Asset struct {
	ID   int64  `yaml:"id"`
	Name string `yaml:"name"`
	Size int    `yaml:"size"`
	URL  string `yaml:"url"`
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

var _ types.SourceAsset = &Asset{}

type Release struct {
	ID           int64     `yaml:"id"`
	Name         string    `yaml:"name"`
	TagName      string    `yaml:"tag_name"`
	URL          string    `yaml:"url"`
	Draft        bool      `yaml:"draft"`
	Prerelease   bool      `yaml:"prerelease"`
	PublishedAt  time.Time `yaml:"published_at"`
	ReleaseNotes string    `yaml:"release_notes"`
	Assets       []*Asset  `yaml:"assets"`
}

func (r *Release) GetID() int64 {
	return r.ID
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

func (r *Release) GetAssets() []types.SourceAsset {
	assets := make([]types.SourceAsset, len(r.Assets))
	for i, asset := range r.Assets {
		assets[i] = asset
	}
	return assets
}

var _ types.SourceRelease = &Release{}
