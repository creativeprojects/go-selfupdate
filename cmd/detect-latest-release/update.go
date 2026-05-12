package main

import (
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
)

// keep this function here, this is the example from the README
//
//nolint:unused
func update(version string) error {
	latest, found, err := selfupdate.DetectLatest(
		context.Background(),
		selfupdate.DetectLatestOpt{Repository: selfupdate.ParseSlug("creativeprojects/resticprofile")},
	)
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		log.Printf("Current version (%s) is the latest", version)
		return nil
	}

	cmdPath, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	// In the new release, the binary file can have any name — this is set by the selfupdate.UpdateToOpt.RelExe.
	// If it's the same as the old one, then selfupdate.UpdateToOpt.RelExe should be empty.
	if err := selfupdate.UpdateTo(context.Background(), selfupdate.UpdateToOpt{Rel: latest, CmdPath: cmdPath}); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
	return nil
}
