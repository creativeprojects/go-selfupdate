package selfupdate

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressionNotRequired(t *testing.T) {
	buf := []byte{'a', 'b', 'c'}
	want := bytes.NewReader(buf)
	r, err := DecompressCommand(want, "https://github.com/foo/bar/releases/download/v1.2.3/foo", "foo", runtime.GOOS, runtime.GOOS)
	require.NoError(t, err)

	have, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Equal(t, buf, have)
}

func getArchiveFileExt(file string) string {
	if strings.HasSuffix(file, ".tar.gz") {
		return ".tar.gz"
	}
	if strings.HasSuffix(file, ".tar.xz") {
		return ".tar.xz"
	}
	return filepath.Ext(file)
}

func TestDecompress(t *testing.T) {
	for _, testCase := range []string{
		"testdata/foo.zip",
		"testdata/single-file.zip",
		"testdata/single-file.gz",
		"testdata/single-file.gzip",
		"testdata/foo.tar.gz",
		"testdata/foo.tgz",
		"testdata/foo.tar.xz",
		"testdata/single-file.xz",
		"testdata/single-file.bz2",
	} {
		t.Run(testCase, func(t *testing.T) {
			f, err := os.Open(testCase)
			require.NoError(t, err)

			ext := getArchiveFileExt(testCase)
			url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
			r, err := DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
			require.NoError(t, err)

			content, err := io.ReadAll(r)
			require.NoError(t, err)

			assert.Equal(t, "this is test\n", string(content), "Decompressing zip failed into unexpected content")
		})
	}
}

func TestDecompressInvalidArchive(t *testing.T) {
	for _, testCase := range []struct {
		name string
		msg  string
	}{
		{"testdata/invalid.zip", "failed to decompress zip file"},
		{"testdata/invalid.gz", "failed to decompress gzip file"},
		{"testdata/invalid-tar.tar.gz", "failed to decompress tar file"},
		{"testdata/invalid-gzip.tar.gz", "failed to decompress tar.gz file"},
		{"testdata/invalid.xz", "failed to decompress xzip file"},
		{"testdata/invalid-tar.tar.xz", "failed to decompress tar file"},
		{"testdata/invalid-xz.tar.xz", "failed to decompress tar.xz file"},
	} {
		f, err := os.Open(testCase.name)
		require.NoError(t, err)

		ext := getArchiveFileExt(testCase.name)
		url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
		_, err = DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
		assert.ErrorIs(t, err, ErrCannotDecompressFile)
		if !strings.Contains(err.Error(), testCase.msg) {
			t.Fatal("Unexpected error:", err)
		}
	}
}

func TestTargetNotFound(t *testing.T) {
	for _, testCase := range []struct {
		name string
		msg  string
	}{
		{"testdata/empty.zip", "not found"},
		{"testdata/bar-not-found.zip", "not found"},
		{"testdata/bar-not-found.gzip", "not found"},
		{"testdata/empty.tar.gz", "not found"},
		{"testdata/bar-not-found.tar.gz", "not found"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			f, err := os.Open(testCase.name)
			require.NoError(t, err)

			ext := getArchiveFileExt(testCase.name)
			url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
			_, err = DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
			assert.ErrorIs(t, err, ErrExecutableNotFoundInArchive)
		})
	}
}

func TestMatchExecutableName(t *testing.T) {
	testData := []struct {
		cmd    string
		os     string
		arch   string
		target string
		found  bool
	}{
		// valid
		{"gostuff", "linux", "amd64", "gostuff", true},
		{"gostuff", "linux", "amd64", "gostuff_0.16.0", true},
		{"gostuff", "linux", "amd64", "gostuff-0.16.0", true},
		{"gostuff", "linux", "amd64", "gostuff_v0.16.0", true},
		{"gostuff", "linux", "amd64", "gostuff-v0.16.0", true},
		{"gostuff", "linux", "amd64", "gostuff_linux_amd64", true},
		{"gostuff", "linux", "amd64", "gostuff-linux-amd64", true},
		{"gostuff", "linux", "amd64", "gostuff_0.16.0_linux_amd64", true},
		{"gostuff", "linux", "amd64", "gostuff-0.16.0-linux-amd64", true},
		{"gostuff", "linux", "amd64", "gostuff_v0.16.0_linux_amd64", true},
		{"gostuff", "linux", "amd64", "gostuff-v0.16.0-linux-amd64", true},
		// invalid
		{"gostuff", "linux", "amd64", "gostuff_darwin_amd64", false},
		{"gostuff", "linux", "amd64", "gostuff0.16.0", false},
		{"gostuff", "linux", "amd64", "gostuffv0.16.0", false},
		{"gostuff", "linux", "amd64", "gostuff_0.16.0_amd64", false},
		{"gostuff", "linux", "amd64", "gostuff_v0.16.0_amd64", false},
		{"gostuff", "linux", "amd64", "gostuff_0.16.0_linux", false},
		{"gostuff", "linux", "amd64", "gostuff_v0.16.0_linux", false},
		// windows valid
		{"gostuff", "windows", "amd64", "gostuff.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_0.16.0.exe", true},
		{"gostuff", "windows", "amd64", "gostuff-0.16.0.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_v0.16.0.exe", true},
		{"gostuff", "windows", "amd64", "gostuff-v0.16.0.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_windows_amd64.exe", true},
		{"gostuff", "windows", "amd64", "gostuff-windows-amd64.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_0.16.0_windows_amd64.exe", true},
		{"gostuff", "windows", "amd64", "gostuff-0.16.0-windows-amd64.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_v0.16.0_windows_amd64.exe", true},
		{"gostuff", "windows", "amd64", "gostuff-v0.16.0-windows-amd64.exe", true},
		// windows invalid
		{"gostuff", "windows", "amd64", "gostuff_darwin_amd64.exe", false},
		{"gostuff", "windows", "amd64", "gostuff0.16.0.exe", false},
		{"gostuff", "windows", "amd64", "gostuff_0.16.0_amd64.exe", false},
		{"gostuff", "windows", "amd64", "gostuff_0.16.0_windows.exe", false},
		{"gostuff", "windows", "amd64", "gostuffv0.16.0.exe", false},
		{"gostuff", "windows", "amd64", "gostuff_v0.16.0_amd64.exe", false},
		{"gostuff", "windows", "amd64", "gostuff_v0.16.0_windows.exe", false},
	}

	for _, testItem := range testData {
		t.Run(testItem.target, func(t *testing.T) {
			assert.Equal(t, testItem.found, matchExecutableName(testItem.cmd, testItem.os, testItem.arch, testItem.target))
		})
		// also try with .exe already in cmd for windows
		if testItem.os == "windows" {
			t.Run(testItem.target, func(t *testing.T) {
				assert.Equal(t, testItem.found, matchExecutableName(testItem.cmd+".exe", testItem.os, testItem.arch, testItem.target))
			})
		}
	}
}

func TestErrorFromReader(t *testing.T) {
	extensions := []string{
		"zip",
		"tar.gz",
		"tgz",
		"gzip",
		"gz",
		"tar.xz",
		"xz",
		"bz2",
	}

	for _, extension := range extensions {
		t.Run(extension, func(t *testing.T) {
			reader, err := DecompressCommand(&bogusReader{}, "foo."+extension, "foo."+extension, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				t.Log(err)
				assert.ErrorIs(t, err, ErrCannotDecompressFile)
			} else {
				// bz2 does not return an error straight away: it only fails when you start reading from the output reader
				_, err = io.ReadAll(reader)
				t.Log(err)
				assert.Error(t, err)
			}
		})
	}
}
