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
	"github.com/creativeprojects/go-selfupdate/internal"
	"github.com/creativeprojects/go-selfupdate/update"
)

type UpdateToOpt struct {
	Rel *Release
	RelExe,
	CmdPath string
}

type UpdateCommandOpt struct {
	RelExe,
	CmdPath,
	Current string
	Repository Repository
}

type UpdateSelfOpt struct {
	RelExe,
	Current string
	Repository Repository
}

// UpdateTo downloads an executable from the source provider and replace current binary with the downloaded one.
// It downloads a release asset via the source provider so this function is available for update releases on private repository.
func (up *Updater) UpdateTo(ctx context.Context, opt UpdateToOpt) error {
	if opt.Rel == nil {
		return ErrInvalidRelease
	}
	if opt.RelExe == "" {
		_, opt.RelExe = filepath.Split(opt.CmdPath)
	}

	data, err := up.download(ctx, opt.Rel, opt.Rel.AssetID)
	if err != nil {
		return fmt.Errorf("failed to read asset %q: %w", opt.Rel.AssetName, err)
	}

	if up.validator != nil {
		err = up.validate(ctx, opt.Rel, data)
		if err != nil {
			return err
		}
	}

	return up.decompressAndUpdate(bytes.NewReader(data), opt.Rel.AssetURL, opt.RelExe, opt.CmdPath)
}

// UpdateCommand updates a given command binary to the latest version.
// 'opt.Current' is used to check the latest version against the current version.
func (up *Updater) UpdateCommand(ctx context.Context, opt UpdateCommandOpt) (*Release, error) {
	version, err := semver.NewVersion(opt.Current)
	if err != nil {
		return nil, fmt.Errorf("incorrect version %q: %w", opt.Current, err)
	}

	if up.os == "windows" && !strings.HasSuffix(opt.CmdPath, ".exe") {
		// Ensure to add '.exe' to given path on Windows
		opt.CmdPath = opt.CmdPath + ".exe"
	}

	stat, err := os.Lstat(opt.CmdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat '%s'. file may not exist: %s", opt.CmdPath, err)
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		p, err := internal.ResolvePath(opt.CmdPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlink '%s' for executable: %s", opt.CmdPath, err)
		}
		opt.CmdPath = p
	}

	rel, ok, err := up.DetectLatest(ctx, DetectLatestOpt{opt.Repository})
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
	log.Printf("Will update %s to the latest version %s", opt.CmdPath, rel.Version())
	if err := up.UpdateTo(ctx, UpdateToOpt{rel, opt.RelExe, opt.CmdPath}); err != nil {
		return nil, err
	}
	return rel, nil
}

// UpdateSelf updates the running executable itself to the latest version.
// 'current' is used to check the latest version against the current version.
func (up *Updater) UpdateSelf(ctx context.Context, opt UpdateSelfOpt) (*Release, error) {
	cmdPath, err := internal.GetExecutablePath()
	if err != nil {
		return nil, err
	}
	return up.UpdateCommand(ctx, UpdateCommandOpt{opt.RelExe, cmdPath, opt.Current, opt.Repository})
}

func (up *Updater) decompressAndUpdate(src io.Reader, assetURL, relExe, cmdPath string) error {
	asset, err := DecompressCommand(DecompressCommandOpt{src, assetURL, relExe, up.os, up.arch})
	if err != nil {
		return err
	}

	log.Printf("Will update %s to the latest downloaded from %s", cmdPath, assetURL)
	return update.Apply(asset, update.Options{
		TargetPath:  cmdPath,
		OldSavePath: up.oldSavePath,
	})
}

// validate loads the validation file and passes it to the validator.
// The validation is successful if no error was returned
func (up *Updater) validate(ctx context.Context, rel *Release, data []byte) error {
	if rel == nil {
		return ErrInvalidRelease
	}

	// compatibility with setting rel.ValidationAssetID directly
	if len(rel.ValidationChain) == 0 {
		rel.ValidationChain = append(rel.ValidationChain, struct {
			ValidationAssetID                       int64
			ValidationAssetName, ValidationAssetURL string
		}{
			ValidationAssetID:   rel.ValidationAssetID,
			ValidationAssetName: "",
			ValidationAssetURL:  rel.ValidationAssetURL,
		})
	}

	validationName := rel.AssetName

	for _, va := range rel.ValidationChain {
		validationData, err := up.download(ctx, rel, va.ValidationAssetID)
		if err != nil {
			return fmt.Errorf("failed reading validation data %q: %w", va.ValidationAssetName, err)
		}

		if err = up.validator.Validate(validationName, data, validationData); err != nil {
			return fmt.Errorf("failed validating asset content %q: %w", validationName, err)
		}

		// Select what next to validate
		validationName = va.ValidationAssetName
		data = validationData
	}
	return nil
}

func (up *Updater) download(ctx context.Context, rel *Release, assetID int64) (data []byte, err error) {
	var reader io.ReadCloser
	if reader, err = up.source.DownloadReleaseAsset(ctx, rel, assetID); err == nil {
		defer func() { _ = reader.Close() }()
		data, err = io.ReadAll(reader)
	}
	return
}
