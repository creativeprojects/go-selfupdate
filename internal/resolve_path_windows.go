//go:build windows

package internal

import (
	"golang.org/x/sys/windows"
	"os"
	"strings"
	"syscall"
)

// ResolvePath returns the path of a given filename with all symlinks resolved.
func ResolvePath(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	handle := windows.Handle(f.Fd())
	buf := make([]uint16, syscall.MAX_PATH)
	_, err = windows.GetFinalPathNameByHandle(handle, &buf[0], uint32(len(buf)), 0)
	if err != nil {
		return "", err
	}
	final := syscall.UTF16ToString(buf)

	// Strip possible "\\?\" prefix
	final = strings.TrimPrefix(final, `\\?\`)

	return final, nil
}
