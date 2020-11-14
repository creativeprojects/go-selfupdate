package selfupdate

import (
	"bytes"
	"io/ioutil"
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

	have, err := ioutil.ReadAll(r)
	require.NoError(t, err)

	for i, b := range have {
		if buf[i] != b {
			t.Error(i, "th elem is not the same as wanted. want", buf[i], "but got", b)
		}
	}
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
	for _, n := range []string{
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
		t.Run(n, func(t *testing.T) {
			f, err := os.Open(n)
			require.NoError(t, err)

			ext := getArchiveFileExt(n)
			url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
			r, err := DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
			require.NoError(t, err)

			bytes, err := ioutil.ReadAll(r)
			require.NoError(t, err)

			s := string(bytes)
			if s != "this is test\n" {
				t.Fatal("Decompressing zip failed into unexpected content", s)
			}
		})
	}
}

func TestDecompressInvalidArchive(t *testing.T) {
	for _, a := range []struct {
		name string
		msg  string
	}{
		{"testdata/invalid.zip", "not a valid zip file"},
		{"testdata/invalid.gz", "failed to decompress gzip file"},
		{"testdata/invalid-tar.tar.gz", "failed to unarchive .tar file"},
		{"testdata/invalid-gzip.tar.gz", "failed to decompress .tar.gz file"},
		{"testdata/invalid.xz", "failed to decompress xzip file"},
		{"testdata/invalid-tar.tar.xz", "failed to unarchive .tar file"},
		{"testdata/invalid-xz.tar.xz", "failed to decompress .tar.xz file"},
	} {
		f, err := os.Open(a.name)
		require.NoError(t, err)

		ext := getArchiveFileExt(a.name)
		url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
		_, err = DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
		if err == nil {
			t.Fatal("Error should be raised")
		}
		if !strings.Contains(err.Error(), a.msg) {
			t.Fatal("Unexpected error:", err)
		}
	}
}

func TestTargetNotFound(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  string
	}{
		{"testdata/empty.zip", "is not found"},
		{"testdata/bar-not-found.zip", "is not found"},
		{"testdata/bar-not-found.gzip", "does not match to command"},
		{"testdata/empty.tar.gz", "is not found"},
		{"testdata/bar-not-found.tar.gz", "is not found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.Open(tc.name)
			require.NoError(t, err)

			ext := getArchiveFileExt(tc.name)
			url := "https://github.com/foo/bar/releases/download/v1.2.3/bar" + ext
			_, err = DecompressCommand(f, url, "bar", runtime.GOOS, runtime.GOOS)
			if err == nil {
				t.Fatal("Error should be raised for")
			}
			if !strings.Contains(err.Error(), tc.msg) {
				t.Fatal("Unexpected error:", err)
			}
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
		{"gostuff", "linux", "amd64", "gostuff", true},
		{"gostuff", "linux", "amd64", "gostuff_linux_amd64", true},
		{"gostuff", "linux", "amd64", "gostuff_darwin_amd64", false},
		{"gostuff", "windows", "amd64", "gostuff.exe", true},
		{"gostuff", "windows", "amd64", "gostuff_windows_amd64.exe", true},
	}

	for _, testItem := range testData {
		t.Run(testItem.target, func(t *testing.T) {
			assert.Equal(t, testItem.found, matchExecutableName(testItem.cmd, testItem.os, testItem.arch, testItem.target))
		})
	}
}
