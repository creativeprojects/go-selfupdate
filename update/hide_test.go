package update

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHideFile(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), t.Name())
	err := hideFile(tempFile)
	assert.NoError(t, err)
}
