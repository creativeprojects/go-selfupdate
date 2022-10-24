package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanUseTokenForDomain(t *testing.T) {
	fixtures := []struct {
		origin, other string
		valid         bool
	}{
		{"http://gitlab.com", "http://gitlab.com", true},
		{"http://gitlab.com/owner/repo", "http://gitlab.com/file", true},
		{"http://gitlab.com/owner/repo", "http://download.gitlab.com/file", true},
		{"http://api.gitlab.com", "http://gitlab.com", false},
		{"http://api.gitlab.com/owner/repo", "http://gitlab.com/file", false},
		{"", "http://gitlab.com/file", false},
	}

	for _, fixture := range fixtures {
		t.Run("", func(t *testing.T) {
			ok, err := canUseTokenForDomain(fixture.origin, fixture.other)
			assert.NoError(t, err)
			assert.Equal(t, fixture.valid, ok)
		})
	}
}
