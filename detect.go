package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var reVersion = regexp.MustCompile(`\d+\.\d+\.\d+`)

// DetectLatest tries to get the latest version from the source provider.
// It fetches releases information from the source provider and find out the latest release with matching the tag names and asset names.
// Drafts and pre-releases are ignored.
// Assets would be suffixed by the OS name and the arch name such as 'foo_linux_amd64' where 'foo' is a command name.
// '-' can also be used as a separator. File can be compressed with zip, gzip, zxip, bzip2, tar&gzip or tar&zxip.
// So the asset can have a file extension for the corresponding compression format such as '.zip'.
// On Windows, '.exe' also can be contained such as 'foo_windows_amd64.exe.zip'.
func (up *Updater) DetectLatest(ctx context.Context, repository Repository) (release *Release, found bool, err error) {
	return up.DetectVersion(ctx, repository, "")
}

// DetectVersion tries to get the given version from the source provider.
// And version indicates the required version.
func (up *Updater) DetectVersion(ctx context.Context, repository Repository, version string) (release *Release, found bool, err error) {
	rel, err := up.source.LatestRelease(ctx, repository)
	if err != nil && !errors.Is(err, ErrNotSupported) {
		log.Printf("Cannot directly fetch latest release: %s", err)
	}
	if rel != nil {
		log.Printf("Latest available release is %s\n", rel.GetTagName())
	}

	rels, err := up.source.ListReleases(ctx, repository)
	if err != nil {
		return nil, false, err
	}

	rel, asset, ver, found := up.findReleaseAndAsset(rels, version)
	if !found {
		return nil, false, nil
	}

	log.Printf("Successfully fetched release %s, name: %s, URL: %s, asset: %s", rel.GetTagName(), rel.GetName(), rel.GetURL(), asset.GetBrowserDownloadURL())

	release = &Release{
		version:           ver,
		repository:        repository,
		AssetURL:          asset.GetBrowserDownloadURL(),
		AssetByteSize:     asset.GetSize(),
		AssetID:           asset.GetID(),
		AssetName:         asset.GetName(),
		ValidationAssetID: -1,
		URL:               rel.GetURL(),
		ReleaseID:         rel.GetID(),
		ReleaseNotes:      rel.GetReleaseNotes(),
		Name:              rel.GetName(),
		PublishedAt:       rel.GetPublishedAt(),
		OS:                up.os,
		Arch:              up.arch,
		Arm:               up.arm,
	}

	if up.validator != nil {
		validationName := up.validator.GetValidationAssetName(asset.GetName())
		validationAsset, ok := findValidationAsset(rel, validationName)
		if !ok {
			return nil, false, fmt.Errorf("%w: %q", ErrValidationAssetNotFound, validationName)
		}
		release.ValidationAssetID = validationAsset.GetID()
	}

	return release, true, nil
}

// findValidationAsset returns the source asset used for validation
func findValidationAsset(rel SourceRelease, validationName string) (SourceAsset, bool) {
	for _, asset := range rel.GetAssets() {
		if asset.GetName() == validationName {
			return asset, true
		}
	}
	return nil, false
}

// findReleaseAndAsset returns the release and asset matching the target version, or latest if target version is empty
func (up *Updater) findReleaseAndAsset(rels []SourceRelease, targetVersion string) (SourceRelease, SourceAsset, *semver.Version, bool) {
	// we put the detected arch at the end of the list: that's fine for ARM so far,
	// as the additional arch are more accurate than the generic one
	for _, arch := range append(generateAdditionalArch(up.arch, up.arm), up.arch) {
		release, asset, version, found := up.findReleaseAndAssetForArch(arch, rels, targetVersion)
		if found {
			return release, asset, version, found
		}
	}

	return nil, nil, nil, false
}

func (up *Updater) findReleaseAndAssetForArch(arch string, rels []SourceRelease, targetVersion string,
) (SourceRelease, SourceAsset, *semver.Version, bool) {
	var ver *semver.Version
	var asset SourceAsset
	var release SourceRelease

	log.Printf("Searching for a possible candidate for os %q and arch %q", up.os, arch)

	// Find the latest version from the list of releases.
	// Returned list from GitHub API is in the order of the date when created.
	for _, rel := range rels {
		if a, v, ok := up.findAssetFromRelease(rel, up.getSuffixes(arch), targetVersion); ok {
			// Note: any version with suffix is less than any version without suffix.
			// e.g. 0.0.1 > 0.0.1-beta
			if release == nil || v.GreaterThan(ver) {
				ver = v
				asset = a
				release = rel
			}
		}
	}

	if release == nil {
		log.Printf("Could not find any release for os %q and arch %q", up.os, arch)
		return nil, nil, nil, false
	}

	return release, asset, ver, true
}

func (up *Updater) findAssetFromRelease(rel SourceRelease, suffixes []string, targetVersion string) (SourceAsset, *semver.Version, bool) {
	if rel == nil {
		log.Print("No source release information")
		return nil, nil, false
	}
	if targetVersion != "" && targetVersion != rel.GetTagName() {
		log.Printf("Skip %s not matching to specified version %s", rel.GetTagName(), targetVersion)
		return nil, nil, false
	}

	if rel.GetDraft() && !up.draft && targetVersion == "" {
		log.Printf("Skip draft version %s", rel.GetTagName())
		return nil, nil, false
	}
	if rel.GetPrerelease() && !up.prerelease && targetVersion == "" {
		log.Printf("Skip pre-release version %s", rel.GetTagName())
		return nil, nil, false
	}

	verText := rel.GetTagName()
	indices := reVersion.FindStringIndex(verText)
	if indices == nil {
		log.Printf("Skip version not adopting semver: %s", verText)
		return nil, nil, false
	}
	if indices[0] > 0 {
		verText = verText[indices[0]:]
	}

	// If semver cannot parse the version text, it means that the text is not adopting
	// the semantic versioning. So it should be skipped.
	ver, err := semver.NewVersion(verText)
	if err != nil {
		log.Printf("Failed to parse a semantic version: %s", verText)
		return nil, nil, false
	}

	for _, asset := range rel.GetAssets() {
		name := asset.GetName()
		if len(up.filters) > 0 {
			// if some filters are defined, match them: if any one matches, the asset is selected
			matched := false
			for _, filter := range up.filters {
				if filter.MatchString(name) {
					log.Printf("Selected filtered asset: %s", name)
					matched = true
					break
				}
				log.Printf("Skipping asset %q not matching filter %v\n", name, filter)
			}
			if !matched {
				continue
			}
		}

		// case insensitive search
		name = strings.ToLower(name)

		for _, s := range suffixes {
			if strings.HasSuffix(name, s) { // require version, arch etc
				// assuming a unique artifact will be a match (or first one will do)
				return asset, ver, true
			}
		}
	}

	log.Printf("No suitable asset was found in release %s", rel.GetTagName())
	return nil, nil, false
}

// getSuffixes returns all candidates to check against the assets
func (up *Updater) getSuffixes(arch string) []string {
	suffixes := make([]string, 0)
	for _, sep := range []rune{'_', '-'} {
		for _, ext := range []string{".zip", ".tar.gz", ".tgz", ".gzip", ".gz", ".tar.xz", ".xz", ".bz2", ""} {
			suffix := fmt.Sprintf("%s%c%s%s", up.os, sep, arch, ext)
			suffixes = append(suffixes, suffix)
			if up.os == "windows" {
				suffix = fmt.Sprintf("%s%c%s.exe%s", up.os, sep, arch, ext)
				suffixes = append(suffixes, suffix)
			}
		}
	}
	return suffixes
}
