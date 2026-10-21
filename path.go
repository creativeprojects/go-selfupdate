package selfupdate

import "github.com/creativeprojects/go-selfupdate/v2/internal"

func ExecutablePath() (string, error) {
	return internal.GetExecutablePath()
}
