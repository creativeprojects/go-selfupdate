package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidSlug(t *testing.T) {
	for _, slug := range []string{
		"foo",
		"/",
		"foo/",
		"/bar",
		"foo/bar/piyo",
	} {
		t.Run(slug, func(t *testing.T) {
			repo := ParseSlug(slug)

			_, _, err := repo.GetSlug()
			assert.Error(t, err)

			_, err = repo.Get()
			assert.Error(t, err)
		})
	}
}

func TestParseSlug(t *testing.T) {
	slug := ParseSlug("foo/bar")

	owner, repo, err := slug.GetSlug()
	assert.NoError(t, err)
	assert.Equal(t, "foo", owner)
	assert.Equal(t, "bar", repo)

	name, err := slug.Get()
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", name)
}

func TestNewRepositorySlug(t *testing.T) {
	slug := NewRepositorySlug("foo", "bar")

	owner, repo, err := slug.GetSlug()
	assert.NoError(t, err)
	assert.Equal(t, "foo", owner)
	assert.Equal(t, "bar", repo)

	name, err := slug.Get()
	assert.NoError(t, err)
	assert.Equal(t, "foo/bar", name)
}
