package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

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
// This function is low-level API to update the binary. Because it does not use a source provider and downloads asset directly from the URL via HTTP,
// this function is not available to update a release for private repositories.
// cmdPath is a file path to command executable.
func UpdateTo(assetURL, assetFileName, cmdPath string) error {
	up := DefaultUpdater()
	src, err := downloadReleaseAssetFromURL(context.Background(), assetURL)
	if err != nil {
		return err
	}
	defer src.Close()
	return up.decompressAndUpdate(src, assetURL, assetFileName, cmdPath)
}

// UpdateCommand updates a given command binary to the latest version.
// This function is a shortcut version of updater.UpdateCommand using a DefaultUpdater()
func UpdateCommand(cmdPath string, current string, slug string) (*Release, error) {
	return DefaultUpdater().UpdateCommand(cmdPath, current, slug)
}

// UpdateSelf updates the running executable itself to the latest version.
// This function is a shortcut version of updater.UpdateSelf using a DefaultUpdater()
func UpdateSelf(current string, slug string) (*Release, error) {
	return DefaultUpdater().UpdateSelf(current, slug)
}

func downloadReleaseAssetFromURL(ctx context.Context, url string) (rc io.ReadCloser, err error) {
	client := http.DefaultClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Accept", "*/*")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download a release file from %s: %w", url, err)
	}
	if resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to download a release file from %s: HTTP %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}
