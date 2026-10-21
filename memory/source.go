package memory

import (
	"bytes"
	"context"
	"io"

	"github.com/creativeprojects/go-selfupdate"
)

// Source is a selfupdate.Source in memory used for unit tests
type Source struct {
	LatestRelease selfupdate.SourceRelease
	Releases      []selfupdate.SourceRelease
	Files         map[int64][]byte
	ReadError     bool // set to true to simulate a read error in DownloadReleaseAsset
}

// ListReleases returns a list of releases. repository parameter is not used.
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	if _, _, err := repository.GetSlug(); err != nil {
		return nil, err
	}
	return s.Releases, nil
}

func (s *Source) GetLatestRelease(ctx context.Context, repository selfupdate.Repository) (selfupdate.SourceRelease, error) {
	if _, _, err := repository.GetSlug(); err != nil {
		return nil, err
	}
	return s.LatestRelease, nil
}

// DownloadReleaseAsset returns a file from its ID. repository parameter is not used.
func (s *Source) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if rel == nil {
		return nil, selfupdate.ErrInvalidRelease
	}
	if _, _, err := rel.Repository.GetSlug(); err != nil {
		return nil, err
	}
	content, ok := s.Files[assetID]
	if !ok {
		return nil, selfupdate.ErrAssetNotFound
	}
	var buffer io.Reader = bytes.NewBuffer(content)
	if s.ReadError {
		// will return a read error after reading 4 characters
		buffer = newErrorReader(buffer, 4)
	}
	return io.NopCloser(buffer), nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
