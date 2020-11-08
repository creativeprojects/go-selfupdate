package selfupdate

import (
	"time"

	"github.com/Masterminds/semver/v3"
)

// Release represents a release asset for current OS and arch.
type Release struct {
	// Version is the version of the release
	Version string
	// AssetURL is a URL to the uploaded file for the release
	AssetURL string
	// AssetSize represents the size of asset in bytes
	AssetByteSize int
	// AssetID is the ID of the asset on GitHub
	AssetID int64
	// ValidationAssetID is the ID of additional validaton asset on GitHub
	ValidationAssetID int64
	// URL is a URL to release page for browsing
	URL string
	// ReleaseNotes is a release notes of the release
	ReleaseNotes string
	// Name represents a name of the release
	Name string
	// PublishedAt is the time when the release was published
	PublishedAt *time.Time
	// RepoOwner is the owner of the repository of the release
	RepoOwner string
	// RepoName is the name of the repository of the release
	RepoName string
	// version is the parsed *semver.Version
	version *semver.Version
}

// Give access to some of the method of the internal semver
// so we can change the version without breaking compatibility

// Equal tests if two versions are equal to each other.
func (r Release) Equal(other string) bool {
	return r.version.Equal(semver.MustParse(other))
}

// LessThan tests if one version is less than another one.
func (r Release) LessThan(other string) bool {
	return r.version.LessThan(semver.MustParse(other))
}

// GreaterThan tests if one version is greater than another one.
func (r Release) GreaterThan(other string) bool {
	return r.version.GreaterThan(semver.MustParse(other))
}
