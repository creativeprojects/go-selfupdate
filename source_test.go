package selfupdate

import (
	"bytes"
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

// ListReleases returns a list of releases. Owner and repo parameters are not used.
func (s *MockSource) ListReleases(owner, repo string) ([]SourceRelease, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
		return nil, err
	}
	return s.releases, nil
}

// DownloadReleaseAsset returns a file from its ID. Owner and repo parameters are not used.
func (s *MockSource) DownloadReleaseAsset(owner, repo string, releaseID, id int64) (io.ReadCloser, error) {
	err := checkOwnerRepoParameters(owner, repo)
	if err != nil {
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
