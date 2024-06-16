package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExecutablePath(t *testing.T) {
	t.Parallel()

	exe, err := GetExecutablePath()
	assert.NoError(t, err)
	assert.NotEmpty(t, exe)
}
