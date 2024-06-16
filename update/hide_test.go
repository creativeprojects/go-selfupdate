package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHideFile(t *testing.T) {
	t.Parallel()

	tempFile := filepath.Join(t.TempDir(), t.Name())
	err := os.WriteFile(tempFile, []byte("test"), 0o644)
	assert.NoError(t, err)

	err = hideFile(tempFile)
	assert.NoError(t, err)
}
