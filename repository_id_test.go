package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryID(t *testing.T) {
	id := NewRepositoryID(11)

	repo, err := id.Get()
	assert.NoError(t, err)
	assert.Equal(t, 11, repo)

	_, _, err = id.GetSlug()
	assert.Error(t, err)
}
