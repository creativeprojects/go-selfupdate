package main

import (
	"context"
	"fmt"
	"log"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/github"
)

// keep this function here, this is the example from the README
//
//nolint:unused
func update(version string) error {
	source, err := github.NewSource(github.Config{})
	if err != nil {
		return fmt.Errorf("error occurred while creating GitHub source: %w", err)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source})
	if err != nil {
		return fmt.Errorf("error occurred while creating updater: %w", err)
	}

	latest, err := updater.DetectLatest(
		context.Background(),
		selfupdate.DetectLatestOpt{Repository: selfupdate.ParseSlug("creativeprojects/resticprofile")},
	)
	if err != nil {
		return err
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
	if err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{Rel: latest, CmdPath: cmdPath}); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
	return nil
}
