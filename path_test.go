package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutablePath(t *testing.T) {
	t.Parallel()

	exe, err := ExecutablePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, exe)
}
