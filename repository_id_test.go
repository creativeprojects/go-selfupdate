package selfupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryID(t *testing.T) {
	id := NewRepositoryID(11)
	assert.Equal(t, 11, id.Get())
	_, _, err := id.GetSlug()
	assert.Error(t, err)
}
