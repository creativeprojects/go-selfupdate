package memory

import (
	"context"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/creativeprojects/go-selfupdate/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorReaderFailsAfterLimit(t *testing.T) {
	testCases := []struct {
		name      string
		content   string
		failAfter int
		expected  string
	}{
		{"content longer than limit", "hello world", 4, "hell"},
		{"content shorter than limit", "hi", 4, "hi"},
		{"empty content", "", 4, ""},
		{"zero limit", "hello world", 0, ""},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := newErrorReader(strings.NewReader(tc.content), tc.failAfter)
			buffer := make([]byte, 64)

			read, err := reader.Read(buffer)
			assert.ErrorIs(t, err, ErrRead)
			assert.Equal(t, tc.expected, string(buffer[:read]))
		})
	}
}

func TestErrorReaderSmallBufferReadsWithoutError(t *testing.T) {
	const content = "hello world"
	const failAfter = 4

	for _, size := range []int{1, 3, failAfter} {
		t.Run("buffer of "+strconv.Itoa(size), func(t *testing.T) {
			reader := newErrorReader(strings.NewReader(content), failAfter)
			buffer := make([]byte, size)

			read, err := reader.Read(buffer)
			require.NoError(t, err)
			assert.Equal(t, content[:size], string(buffer[:read]))
		})
	}
}

func TestErrorReaderReadAll(t *testing.T) {
	reader := newErrorReader(strings.NewReader("hello world"), 4)

	content, err := io.ReadAll(reader)
	assert.ErrorIs(t, err, ErrRead)
	assert.Equal(t, "hell", string(content))
}

func TestSourceReadError(t *testing.T) {
	source := &Source{
		ReadError: true,
		Files:     map[int64][]byte{1: []byte("hello world")},
	}
	release := &selfupdate.Release{
		Repository: selfupdate.NewRepositorySlug("owner", "repo"),
		AssetID:    1,
	}

	reader, err := source.DownloadReleaseAsset(context.Background(), release, 1)
	require.NoError(t, err)
	defer reader.Close()

	content, err := io.ReadAll(reader)
	assert.ErrorIs(t, err, ErrRead)
	assert.Equal(t, "hell", string(content))
}
