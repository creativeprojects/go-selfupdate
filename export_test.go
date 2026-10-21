package selfupdate

import (
	"context"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// Wrappers giving the external test package access to unexported methods.

func (up *Updater) Validate(ctx context.Context, rel *Release, data []byte) error {
	return up.validate(ctx, rel, data)
}

func (up *Updater) FindReleaseAndAsset(rels []SourceRelease, targetVersion string) (SourceRelease, SourceAsset, *semver.Version, bool) {
	return up.findReleaseAndAsset(rels, targetVersion)
}

func (up *Updater) FindAssetFromRelease(rel SourceRelease, suffixes []string, targetVersion string) (SourceAsset, *semver.Version, bool) {
	return up.findAssetFromRelease(rel, suffixes, targetVersion)
}

func (up *Updater) Filters() []*regexp.Regexp {
	return up.filters
}
