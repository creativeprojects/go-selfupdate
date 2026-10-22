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

// Package http provides a selfupdate.Source loading releases from a plain HTTP server.
//
// Each repository lives at {BaseURL}/{owner}/{repo}/ and holds a manifest.yaml
// (see Manifest) listing its releases and assets next to the asset files.
// Asset URLs in the manifest are either absolute or relative to the repository directory.
//
// Such a repository can be built with goreleaser-http-repo-builder:
// https://github.com/GRMrGecko/goreleaser-http-repo-builder
package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"

	"github.com/creativeprojects/go-selfupdate/v2"
	yaml "go.yaml.in/yaml/v3"
)

type Manifest struct {
	LastReleaseID int64      `yaml:"last_release_id"`
	LastAssetID   int64      `yaml:"last_asset_id"`
	Releases      []*Release `yaml:"releases"`
}

// Config is an object to pass to NewSource
type Config struct {
	// BaseURL is a base URL of your update server. This parameter has NO default value.
	BaseURL string
	// HTTP Transport Config
	Transport *http.Transport
	// Additional headers
	Headers http.Header
	// Logger receives the messages describing what the source is doing. Default is to discard them.
	Logger selfupdate.Logger
}

// Source is used to load release information from an http repository
type Source struct {
	baseURL   string
	transport *http.Transport
	headers   http.Header
	logger    selfupdate.Logger
}

// NewSource creates a new http.Source from a config object.
func NewSource(config Config) (*Source, error) {
	// Validate Base URL.
	if config.BaseURL == "" {
		return nil, fmt.Errorf("http base url must be set")
	}
	_, perr := url.ParseRequestURI(config.BaseURL)
	if perr != nil {
		return nil, perr
	}

	// Setup standard transport if not set.
	if config.Transport == nil {
		config.Transport = &http.Transport{}
	}
	logger := config.Logger
	if logger == nil {
		logger = selfupdate.DiscardLogger
	}

	// Return new source.
	return &Source{
		baseURL:   config.BaseURL,
		transport: config.Transport,
		headers:   config.Headers,
		logger:    logger,
	}, nil
}

// Returns a full URI for a relative path URI.
func (s *Source) uriRelative(uri, owner, repo string) string {
	// If URI is blank, its blank.
	if uri != "" {
		// If we're able to parse the URI, a full URI is already defined.
		_, perr := url.ParseRequestURI(uri)
		if perr != nil {
			// Join the paths if possible to make a full URI.
			newURL, jerr := url.JoinPath(s.baseURL, owner, repo, uri)
			if jerr == nil {
				uri = newURL
			}
		}
	}
	return uri
}

// resolveURLs makes the release and asset URLs absolute when the manifest holds them relative to the repository.
func (s *Source) resolveURLs(release *Release, owner, repo string) {
	release.URL = s.uriRelative(release.URL, owner, repo)
	for _, asset := range release.Assets {
		asset.URL = s.uriRelative(asset.URL, owner, repo)
	}
}

func (s *Source) getManifest(ctx context.Context, owner, repo string) (*Manifest, error) {
	// Make repository URI.
	uri, err := url.JoinPath(s.baseURL, owner, repo, "manifest.yaml")
	if err != nil {
		return nil, err
	}

	// Setup HTTP client.
	client := &http.Client{Transport: s.transport}

	// Make repository request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Add headers to request.
	req.Header = s.headers

	// Perform the request.
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		if res.StatusCode == http.StatusNotFound {
			return nil, selfupdate.ErrRepositoryNotFound
		}
		return nil, fmt.Errorf("HTTP request failed with status code %d", res.StatusCode)
	}

	// Decode the response.
	manifest := new(Manifest)
	defer res.Body.Close()
	decoder := yaml.NewDecoder(res.Body)
	err = decoder.Decode(manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

// ListReleases returns all available releases
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}

	manifest, err := s.getManifest(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	// Make a release array.
	releases := make([]selfupdate.SourceRelease, len(manifest.Releases))
	for i, release := range manifest.Releases {
		s.resolveURLs(release, owner, repo)
		releases[i] = release
	}

	return releases, nil
}

func (s *Source) GetLatestRelease(ctx context.Context, repository selfupdate.Repository) (selfupdate.SourceRelease, error) {
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return nil, err
	}

	manifest, err := s.getManifest(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(manifest.Releases, func(release *Release) bool {
		return release.ID == manifest.LastReleaseID
	})
	if index == -1 {
		return nil, selfupdate.ErrReleaseNotFound
	}
	release := manifest.Releases[index]
	s.resolveURLs(release, owner, repo)
	return release, nil
}

// DownloadReleaseAsset downloads an asset from a release.
// It returns an io.ReadCloser: it is your responsibility to Close it.
func (s *Source) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if rel == nil {
		return nil, selfupdate.ErrInvalidRelease
	}

	// Determine download url based on asset id.
	var downloadURL string
	if rel.AssetID == assetID {
		downloadURL = rel.AssetURL
	} else if rel.ValidationAssetID == assetID {
		downloadURL = rel.ValidationAssetURL
	}
	if downloadURL == "" {
		return nil, fmt.Errorf("asset ID %d: %w", assetID, selfupdate.ErrAssetNotFound)
	}

	s.logger.Printf("downloading %q", downloadURL)
	// Setup HTTP client.
	client := &http.Client{Transport: s.transport}

	// Make request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, err
	}

	// Add headers to request.
	req.Header = s.headers

	// Perform the request.
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("HTTP request failed with status code %d", response.StatusCode)
	}

	return response.Body, nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
