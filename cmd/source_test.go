package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitDomainSlug(t *testing.T) {
	fixtures := []struct {
		repo    string
		domain  string
		slug    string
		isValid bool
	}{
		{"owner/name", "", "owner/name", true},
		{"owner/name/", "", "", false},
		{"/owner/name", "", "", false},
		{"github.com/owner/name", "http://github.com", "owner/name", true},
		{"http://github.com/owner/name", "http://github.com", "owner/name", true},
		{"http://github.com", "", "", false},
		{"http://github.com/", "", "", false},
		{"https://github.com/", "", "", false},
		{"github.com/", "", "", false},
		{"github.com", "", "", false},
	}

	for _, fixture := range fixtures {
		t.Run(fixture.repo, func(t *testing.T) {
			domain, slug, err := SplitDomainSlug(fixture.repo)
			assert.Equal(t, fixture.domain, domain)
			assert.Equal(t, fixture.slug, slug)
			if fixture.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				t.Log(err)
			}
		})
	}
}
