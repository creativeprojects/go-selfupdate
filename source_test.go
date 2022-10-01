package selfupdate

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
)

// MockSource is a Source in memory used for unit tests
type MockSource struct {
	releases  []SourceRelease
	files     map[int64][]byte
	readError bool
}

// NewMockSource instantiates a new MockSource
func NewMockSource(releases []SourceRelease, files map[int64][]byte) *MockSource {
	return &MockSource{
		releases: releases,
		files:    files,
	}
}

// ListReleases returns a list of releases. repository parameter is not used.
func (s *MockSource) ListReleases(ctx context.Context, repository Repository) ([]SourceRelease, error) {
	if _, _, err := repository.GetSlug(); err != nil {
		return nil, err
	}
	return s.releases, nil
}

// LatestRelease only returns the most recent release
func (s *MockSource) LatestRelease(ctx context.Context, repository Repository) (SourceRelease, error) {
	if _, _, err := repository.GetSlug(); err != nil {
		return nil, err
	}
	return s.releases[len(s.releases)-1], nil
}

// DownloadReleaseAsset returns a file from its ID. repository parameter is not used.
func (s *MockSource) DownloadReleaseAsset(ctx context.Context, repository Repository, releaseID, id int64) (io.ReadCloser, error) {
	if _, _, err := repository.GetSlug(); err != nil {
		return nil, err
	}
	content, ok := s.files[id]
	if !ok {
		return nil, ErrAssetNotFound
	}
	var buffer io.Reader = bytes.NewBuffer(content)
	if s.readError {
		// will return a read error after reading 4 caracters
		buffer = newErrorReader(buffer, 4)
	}
	return ioutil.NopCloser(buffer), nil
}

// Verify interface
var _ Source = &MockSource{}
