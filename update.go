package selfupdate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate/update"
)

// UpdateTo downloads an executable from the source provider and replace current binary with the downloaded one.
// It downloads a release asset via the source provider so this function is available for update releases on private repository.
func (up *Updater) UpdateTo(ctx context.Context, rel *Release, cmdPath string) error {
	if rel == nil {
		return ErrInvalidRelease
	}
	src, err := up.source.DownloadReleaseAsset(ctx, rel, rel.AssetID)
	if err != nil {
		return err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("failed to read asset: %w", err)
	}

	if up.validator != nil {
		err = up.validate(ctx, rel, data)
		if err != nil {
			return err
		}
	}

	return up.decompressAndUpdate(bytes.NewReader(data), rel.AssetName, rel.AssetURL, cmdPath)
}

// UpdateCommand updates a given command binary to the latest version.
// 'current' is used to check the latest version against the current version.
func (up *Updater) UpdateCommand(ctx context.Context, cmdPath string, current string, repository Repository) (*Release, error) {
	version, err := semver.NewVersion(current)
	if err != nil {
		return nil, fmt.Errorf("incorrect version %q: %w", current, err)
	}

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

	rel, ok, err := up.DetectLatest(ctx, repository)
	if err != nil {
		return nil, err
	}
	if !ok {
		log.Print("No release detected. Current version is considered up-to-date")
		return &Release{version: version}, nil
	}
	if version.Equal(rel.version) {
		log.Printf("Current version %s is the latest. Update is not needed", version.String())
		return rel, nil
	}
	log.Printf("Will update %s to the latest version %s", cmdPath, rel.Version())
	if err := up.UpdateTo(ctx, rel, cmdPath); err != nil {
		return nil, err
	}
	return rel, nil
}

// UpdateSelf updates the running executable itself to the latest version.
// 'current' is used to check the latest version against the current version.
func (up *Updater) UpdateSelf(ctx context.Context, current string, repository Repository) (*Release, error) {
	cmdPath, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return up.UpdateCommand(ctx, cmdPath, current, repository)
}

func (up *Updater) decompressAndUpdate(src io.Reader, assetName, assetURL, cmdPath string) error {
	_, cmd := filepath.Split(cmdPath)
	asset, err := DecompressCommand(src, assetName, cmd, up.os, up.arch)
	if err != nil {
		return err
	}

	log.Printf("Will update %s to the latest downloaded from %s", cmdPath, assetURL)
	return update.Apply(asset, update.Options{
		TargetPath: cmdPath,
	})
}

// validate loads the validation file and passes it to the validator.
// The validation is successful if no error was returned
func (up *Updater) validate(ctx context.Context, rel *Release, data []byte) error {
	if rel == nil {
		return ErrInvalidRelease
	}
	validationSrc, err := up.source.DownloadReleaseAsset(ctx, rel, rel.ValidationAssetID)
	if err != nil {
		return err
	}
	defer validationSrc.Close()

	validationData, err := io.ReadAll(validationSrc)
	if err != nil {
		return fmt.Errorf("failed reading validation data: %w", err)
	}

	if err := up.validator.Validate(rel.AssetName, data, validationData); err != nil {
		return fmt.Errorf("failed validating asset content: %w", err)
	}
	return nil
}
