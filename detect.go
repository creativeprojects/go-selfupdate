package selfupdate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var reVersion = regexp.MustCompile(`\d+\.\d+\.\d+`)

// DetectLatest tries to get the latest version of the repository on GitHub.
// 'slug' means 'owner/name' formatted string.
// It fetches releases information from GitHub API and find out the latest release with matching the tag names and asset names.
// Drafts and pre-releases are ignored.
// Assets would be suffixed by the OS name and the arch name such as 'foo_linux_amd64' where 'foo' is a command name.
// '-' can also be used as a separator. File can be compressed with zip, gzip, zxip, bzip2, tar&gzip or tar&zxip.
// So the asset can have a file extension for the corresponding compression format such as '.zip'.
// On Windows, '.exe' also can be contained such as 'foo_windows_amd64.exe.zip'.
func (up *Updater) DetectLatest(slug string) (release *Release, found bool, err error) {
	return up.DetectVersion(slug, "")
}

// DetectVersion tries to get the given version of the repository on Github. `slug` means `owner/name` formatted string.
// And version indicates the required version.
func (up *Updater) DetectVersion(slug string, version string) (release *Release, found bool, err error) {
	repo := strings.Split(slug, "/")
	if len(repo) != 2 || repo[0] == "" || repo[1] == "" {
		return nil, false, ErrInvalidSlug
	}

	rels, err := up.source.ListReleases(repo[0], repo[1])
	if err != nil {
		return nil, false, err
	}

	rel, asset, ver, found := up.findReleaseAndAsset(rels, version)
	if !found {
		return nil, false, nil
	}

	url := asset.GetBrowserDownloadURL()
	log.Printf("Successfully fetched the latest release. tag: %s, name: %s, URL: %s, Asset: %s", rel.GetTagName(), rel.GetName(), rel.GetURL(), url)

	release = &Release{
		version:           ver,
		repoOwner:         repo[0],
		repoName:          repo[1],
		AssetURL:          url,
		AssetByteSize:     asset.GetSize(),
		AssetID:           asset.GetID(),
		AssetName:         asset.GetName(),
		ValidationAssetID: -1,
		URL:               rel.GetURL(),
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
			return nil, false, fmt.Errorf("failed finding validation file %q", validationName)
		}
		release.ValidationAssetID = validationAsset.GetID()
	}

	return release, true, nil
}

func findAssetFromRelease(rel SourceRelease, suffixes []string, targetVersion string, filters []*regexp.Regexp) (SourceAsset, *semver.Version, bool) {
	if rel == nil {
		log.Print("Empty release instance!")
		return nil, nil, false
	}
	if targetVersion != "" && targetVersion != rel.GetTagName() {
		log.Printf("Skip %s not matching to specified version %s", rel.GetTagName(), targetVersion)
		return nil, nil, false
	}

	if targetVersion == "" && rel.GetDraft() {
		log.Printf("Skip draft version %s", rel.GetTagName())
		return nil, nil, false
	}
	if targetVersion == "" && rel.GetPrerelease() {
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
		// log.Printf("Strip prefix '%s' from '%s'", verText[:indices[0]], verText)
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
		if len(filters) > 0 {
			// if some filters are defined, match them: if any one matches, the asset is selected
			matched := false
			for _, filter := range filters {
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

		for _, s := range suffixes {
			if strings.HasSuffix(name, s) { // require version, arch etc
				// default: assume single artifact
				return asset, ver, true
			}
		}
	}

	log.Printf("No suitable asset was found in release %s", rel.GetTagName())
	return nil, nil, false
}

func findValidationAsset(rel SourceRelease, validationName string) (SourceAsset, bool) {
	for _, asset := range rel.GetAssets() {
		if asset.GetName() == validationName {
			return asset, true
		}
	}
	return nil, false
}

func (up *Updater) findReleaseAndAsset(rels []SourceRelease, targetVersion string) (SourceRelease, SourceAsset, *semver.Version, bool) {
	// we put the detected arch at the end of the list: that's fine for ARM so far,
	// as the additional arch are more accurate than the generic one
	for _, arch := range append(generateAdditionalArch(up.arch, up.arm), up.arch) {
		release, asset, version, found := findReleaseAndAssetForArch(up.os, arch, rels, targetVersion, up.filters)
		if found {
			return release, asset, version, found
		}
	}

	return nil, nil, nil, false
}

func findReleaseAndAssetForArch(
	os string,
	arch string,
	rels []SourceRelease,
	targetVersion string,
	filters []*regexp.Regexp,
) (SourceRelease, SourceAsset, *semver.Version, bool) {
	// Generate candidates
	suffixes := make([]string, 0, 2*7*2)
	for _, sep := range []rune{'_', '-'} {
		for _, ext := range []string{".zip", ".tar.gz", ".tgz", ".gzip", ".gz", ".tar.xz", ".xz", ".bz2", ""} {
			suffix := fmt.Sprintf("%s%c%s%s", os, sep, arch, ext)
			suffixes = append(suffixes, suffix)
			if os == "windows" {
				suffix = fmt.Sprintf("%s%c%s.exe%s", os, sep, arch, ext)
				suffixes = append(suffixes, suffix)
			}
		}
	}

	var ver *semver.Version
	var asset SourceAsset
	var release SourceRelease

	// Find the latest version from the list of releases.
	// Returned list from GitHub API is in the order of the date when created.
	for _, rel := range rels {
		if a, v, ok := findAssetFromRelease(rel, suffixes, targetVersion, filters); ok {
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
		log.Printf("Could not find any release for os %q and arch %q", os, arch)
		return nil, nil, nil, false
	}

	return release, asset, ver, true
}
