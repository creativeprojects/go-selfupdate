package memory_test

import (
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/creativeprojects/go-selfupdate/v2/internal/memory"
	"github.com/creativeprojects/go-selfupdate/v2/internal/sourcetest"
)

// TestMemorySource runs the shared source test suite: the memory source is the reference implementation.
func TestMemorySource(t *testing.T) {
	sourcetest.Run(t, func(_ *testing.T, data *memory.Source) selfupdate.Source {
		return data
	}, sourcetest.Options{})
}
