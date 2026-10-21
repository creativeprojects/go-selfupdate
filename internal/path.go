// Package internal holds helpers shared by the selfupdate packages:
// resolving the executable path and deciding whether an API token
// can be sent along with a download request.
package internal

import (
	"os"
)

// GetExecutablePath returns the path of the executable file with all symlinks resolved.
func GetExecutablePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	exe, err = ResolvePath(exe)
	if err != nil {
		return "", err
	}

	return exe, nil
}
