package selfupdate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/inconshreveable/go-update"
)

// UpdateTo downloads an executable from GitHub Releases API and replace current binary with the downloaded one.
// It downloads a release asset via GitHub Releases API so this function is available for update releases on private repository.
// If a redirect occurs, it fallbacks into directly downloading from the redirect URL.
func (up *Updater) UpdateTo(rel *Release, cmdPath string) error {
	var client http.Client
	src, redirectURL, err := up.api.Repositories.DownloadReleaseAsset(up.apiCtx, rel.RepoOwner, rel.RepoName, rel.AssetID, &client)
	if err != nil {
		return fmt.Errorf("failed to call GitHub Releases API for getting an asset(ID: %d) for repository '%s/%s': %s", rel.AssetID, rel.RepoOwner, rel.RepoName, err)
	}
	if redirectURL != "" {
		log.Printf("Redirect URL was returned while trying to download a release asset from GitHub API. Falling back to downloading from asset URL directly: %s", redirectURL)
		src, err = up.downloadDirectlyFromURL(redirectURL)
		if err != nil {
			return err
		}
	}
	defer src.Close()

	data, err := ioutil.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed reading asset body: %v", err)
	}

	if up.validator == nil {
		return up.decompressAndUpdate(bytes.NewReader(data), rel.AssetURL, cmdPath)
	}

	validationSrc, validationRedirectURL, err := up.api.Repositories.DownloadReleaseAsset(up.apiCtx, rel.RepoOwner, rel.RepoName, rel.ValidationAssetID, &client)
	if err != nil {
		return fmt.Errorf("failed to call GitHub Releases API for getting an validation asset(ID: %d) for repository '%s/%s': %s", rel.ValidationAssetID, rel.RepoOwner, rel.RepoName, err)
	}
	if validationRedirectURL != "" {
		log.Print("Redirect URL was returned while trying to download a release validation asset from GitHub API. Falling back to downloading from asset URL directly:", redirectURL)
		validationSrc, err = up.downloadDirectlyFromURL(validationRedirectURL)
		if err != nil {
			return err
		}
	}

	defer validationSrc.Close()

	validationData, err := ioutil.ReadAll(validationSrc)
	if err != nil {
		return fmt.Errorf("failed reading validation asset body: %v", err)
	}

	if err := up.validator.Validate(data, validationData); err != nil {
		return fmt.Errorf("failed validating asset content: %v", err)
	}

	return up.decompressAndUpdate(bytes.NewReader(data), rel.AssetURL, cmdPath)
}

// UpdateCommand updates a given command binary to the latest version.
// 'slug' represents 'owner/name' repository on GitHub and 'current' means the current version.
func (up *Updater) UpdateCommand(cmdPath string, current *semver.Version, slug string) (*Release, error) {
	if up.os == "windows" && !strings.HasSuffix(cmdPath, ".exe") {
		// Ensure to add '.exe' to given path on Windows
		cmdPath = cmdPath + ".exe"
	}

	stat, err := os.Lstat(cmdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat '%s'. file may not exist: %s", cmdPath, err)
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		p, err := filepath.EvalSymlinks(cmdPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlink '%s' for executable: %s", cmdPath, err)
		}
		cmdPath = p
	}

	rel, ok, err := up.DetectLatest(slug)
	if err != nil {
		return nil, err
	}
	if !ok {
		log.Print("No release detected. Current version is considered up-to-date")
		return &Release{version: current}, nil
	}
	if current.Equal(rel.version) {
		log.Print("Current version", current, "is the latest. Update is not needed")
		return rel, nil
	}
	log.Printf("Will update %s to the latest version %s", cmdPath, rel.Version())
	if err := up.UpdateTo(rel, cmdPath); err != nil {
		return nil, err
	}
	return rel, nil
}

// UpdateSelf updates the running executable itself to the latest version.
// 'slug' represents 'owner/name' repository on GitHub and 'current' means the current version.
func (up *Updater) UpdateSelf(current *semver.Version, slug string) (*Release, error) {
	cmdPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return up.UpdateCommand(cmdPath, current, slug)
}

func (up *Updater) decompressAndUpdate(src io.Reader, assetURL, cmdPath string) error {
	_, cmd := filepath.Split(cmdPath)
	asset, err := DecompressCommand(src, assetURL, cmd, up.os, up.arch)
	if err != nil {
		return err
	}

	log.Printf("Will update %s to the latest downloaded from %s", cmdPath, assetURL)
	return update.Apply(asset, update.Options{
		TargetPath: cmdPath,
	})
}

func (up *Updater) downloadDirectlyFromURL(assetURL string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to %s: %s", assetURL, err)
	}

	req.Header.Add("Accept", "application/octet-stream")
	req = req.WithContext(up.apiCtx)

	// OAuth HTTP client is not available to download blob from URL when the URL is a redirect URL
	// returned from GitHub Releases API (response status 400).
	// Use default HTTP client instead.
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download a release file from %s: %s", assetURL, err)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download a release file from %s: Not successful status %d", assetURL, res.StatusCode)
	}

	return res.Body, nil
}
