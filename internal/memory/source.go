// Package memory provides an in-memory selfupdate.Source for unit tests.
//
// Releases, assets and file contents are set directly on Source;
// its ReadError and NotFound fields simulate failures of a remote source.
package memory

import (
	"bytes"
	"context"
	"io"

	"github.com/creativeprojects/go-selfupdate/v2"
)

// Source is a selfupdate.Source in memory used for unit tests
type Source struct {
	// Repository restricts the source to a single repository: any other repository is reported as not found.
	// All repositories are accepted when nil.
	Repository    selfupdate.Repository
	LatestRelease selfupdate.SourceRelease
	Releases      []selfupdate.SourceRelease
	Files         map[int64][]byte
	ReadError     bool // set to true to simulate a read error in DownloadReleaseAsset
	NotFound      bool // set to true for a repository not found
}

// ListReleases returns a list of releases.
func (s *Source) ListReleases(ctx context.Context, repository selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	if err := s.check(ctx, repository); err != nil {
		return nil, err
	}
	return s.Releases, nil
}

// GetLatestRelease returns LatestRelease, or ErrReleaseNotFound when it is not set.
func (s *Source) GetLatestRelease(ctx context.Context, repository selfupdate.Repository) (selfupdate.SourceRelease, error) {
	if err := s.check(ctx, repository); err != nil {
		return nil, err
	}
	if s.LatestRelease == nil {
		return nil, selfupdate.ErrReleaseNotFound
	}
	return s.LatestRelease, nil
}

// DownloadReleaseAsset returns a file from its ID.
func (s *Source) DownloadReleaseAsset(ctx context.Context, rel *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	if rel == nil {
		return nil, selfupdate.ErrInvalidRelease
	}
	if err := s.check(ctx, rel.Repository); err != nil {
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

// check fails like a remote source would: on a cancelled context, an invalid slug or an unknown repository.
func (s *Source) check(ctx context.Context, repository selfupdate.Repository) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	owner, repo, err := repository.GetSlug()
	if err != nil {
		return err
	}
	if s.NotFound {
		return selfupdate.ErrRepositoryNotFound
	}
	if s.Repository != nil {
		expectedOwner, expectedRepo, err := s.Repository.GetSlug()
		if err != nil {
			return err
		}
		if owner != expectedOwner || repo != expectedRepo {
			return selfupdate.ErrRepositoryNotFound
		}
	}
	return nil
}

// Verify interface
var _ selfupdate.Source = &Source{}
