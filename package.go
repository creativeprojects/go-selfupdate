package selfupdate

import "github.com/Masterminds/semver/v3"

// DetectLatest detects the latest release of the slug (owner/repo).
// This function is a shortcut version of updater.DetectLatest.
func DetectLatest(slug string) (*Release, bool, error) {
	return DefaultUpdater().DetectLatest(slug)
}

// DetectVersion detects the given release of the slug (owner/repo) from its version.
func DetectVersion(slug string, version string) (*Release, bool, error) {
	return DefaultUpdater().DetectVersion(slug, version)
}

// UpdateTo downloads an executable from assetURL and replaces the current binary with the downloaded one.
// This function is low-level API to update the binary. Because it does not use GitHub API and downloads asset directly from the URL via HTTP,
// this function is not available to update a release for private repositories.
// cmdPath is a file path to command executable.
func UpdateTo(assetURL, cmdPath string) error {
	up := DefaultUpdater()
	src, err := up.downloadDirectlyFromURL(assetURL)
	if err != nil {
		return err
	}
	defer src.Close()
	return up.decompressAndUpdate(src, assetURL, cmdPath)
}

// UpdateCommand updates a given command binary to the latest version.
// This function is a shortcut version of updater.UpdateCommand.
func UpdateCommand(cmdPath string, current *semver.Version, slug string) (*Release, error) {
	return DefaultUpdater().UpdateCommand(cmdPath, current, slug)
}

// UpdateSelf updates the running executable itself to the latest version.
// This function is a shortcut version of updater.UpdateSelf.
func UpdateSelf(current *semver.Version, slug string) (*Release, error) {
	return DefaultUpdater().UpdateSelf(current, slug)
}
