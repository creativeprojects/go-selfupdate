package selfupdate

import "github.com/sinspired/go-selfupdate/internal"

func ExecutablePath() (string, error) {
	return internal.GetExecutablePath()
}
