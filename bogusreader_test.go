package selfupdate

import (
	"errors"
	"io"
)

type bogusReader struct{}

func (r *bogusReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("cannot read")
}

// Verify interface
var _ io.Reader = &bogusReader{}
