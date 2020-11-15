package selfupdate

import (
	"errors"
	"io"
)

var (
	errTestRead = errors.New("read error")
)

// errorReader is a reader that will throw an error after reading n characters
type errorReader struct {
	r         io.Reader
	failAfter int
}

// newErrorReader creates a new reader that will thrown an error after reading n characters
func newErrorReader(r io.Reader, failAfterBytes int) *errorReader {
	return &errorReader{
		r:         r,
		failAfter: failAfterBytes,
	}
}

// Read will throw an error after reading n characters
func (r *errorReader) Read(p []byte) (int, error) {
	if len(p) <= r.failAfter {
		return r.Read(p)
	}
	read, _ := r.r.Read(p[0 : r.failAfter-1])
	return read, errTestRead
}

// Verify interface
var _ io.Reader = &errorReader{}
