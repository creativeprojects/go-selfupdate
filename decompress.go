package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ulikunitz/xz"
)

var (
	fileTypes = []struct {
		ext        string
		decompress func(up *Updater, src io.Reader, releaseBinary string) (io.Reader, error)
	}{
		{".zip", (*Updater).unzip},
		{".tar.gz", (*Updater).untargz},
		{".tgz", (*Updater).untargz},
		{".gzip", (*Updater).gunzip},
		{".gz", (*Updater).gunzip},
		{".tar.xz", (*Updater).untarxz},
		{".xz", (*Updater).unxz},
		{".bz2", (*Updater).unbz2},
	}
	// pattern copied from bottom of the page: https://semver.org/
	semverPattern = `(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?`
)

// fromStream decompresses the binary from a stream. Archive and compression format is
// automatically detected from the 'url' parameter, which represents the URL of asset,
// or simply a filename (with an extension).
// This returns a reader for the decompressed command given by 'url'. '.zip',
// '.tar.gz', '.tar.xz', '.tgz', '.gz', '.bz2' and '.xz' are supported.
//
// These wrapped errors can be returned:
//   - ErrCannotDecompressFile
//   - ErrExecutableNotFoundInArchive
func (up *Updater) fromStream(reader io.Reader, url, releaseBinary string) (io.Reader, error) {
	for _, fileType := range fileTypes {
		if strings.HasSuffix(url, fileType.ext) {
			return fileType.decompress(up, reader, releaseBinary)
		}
	}
	up.logger.Print("File is not compressed")
	return reader, nil
}

func (up *Updater) unzip(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing zip file")

	// Zip format requires its file size for Decompressing.
	// So we need to read the HTTP response into a buffer at first.
	buf, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("%w zip file: %v", ErrCannotDecompressFile, err)
	}

	r := bytes.NewReader(buf)
	z, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, fmt.Errorf("%w zip file: %s", ErrCannotDecompressFile, err)
	}

	for _, file := range z.File {
		_, target := filepath.Split(file.Name)
		if !file.FileInfo().IsDir() && matchExecutableName(releaseBinary, up.os, up.arch, target) {
			up.logger.Printf("Executable file %q was found in zip archive", file.Name)
			return file.Open()
		}
	}

	return nil, fmt.Errorf("%w in zip file: %q", ErrExecutableNotFoundInArchive, releaseBinary)
}

func (up *Updater) untargz(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing tar.gz file")

	gz, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("%w tar.gz file: %s", ErrCannotDecompressFile, err)
	}

	return up.unarchiveTar(gz, releaseBinary)
}

func (up *Updater) gunzip(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing gzip file")

	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("%w gzip file: %s", ErrCannotDecompressFile, err)
	}

	target := r.Name
	if !matchExecutableName(releaseBinary, up.os, up.arch, target) {
		return nil, fmt.Errorf("%w: expected %q but found %q", ErrExecutableNotFoundInArchive, releaseBinary, target)
	}

	up.logger.Printf("Executable file %q was found in gzip file", target)
	return r, nil
}

func (up *Updater) untarxz(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing tar.xz file")

	xzip, err := xz.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("%w tar.xz file: %s", ErrCannotDecompressFile, err)
	}

	return up.unarchiveTar(xzip, releaseBinary)
}

func (up *Updater) unxz(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing xzip file")

	xzip, err := xz.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("%w xzip file: %s", ErrCannotDecompressFile, err)
	}

	up.logger.Printf("Decompressed file from xzip is assumed to be an executable: %s", releaseBinary)
	return xzip, nil
}

func (up *Updater) unbz2(src io.Reader, releaseBinary string) (io.Reader, error) {
	up.logger.Print("Decompressing bzip2 file")

	bz2 := bzip2.NewReader(src)

	up.logger.Printf("Decompressed file from bzip2 is assumed to be an executable: %s", releaseBinary)
	return bz2, nil
}

func matchExecutableName(releaseBinary, os, arch, target string) bool {
	releaseBinary = strings.TrimSuffix(releaseBinary, ".exe")
	pattern := regexp.MustCompile(
		fmt.Sprintf(
			`^%s([_-]v?%s)?([_-]%s[_-]%s)?(\.exe)?$`,
			regexp.QuoteMeta(releaseBinary),
			semverPattern,
			regexp.QuoteMeta(os),
			regexp.QuoteMeta(arch),
		),
	)
	return pattern.MatchString(target)
}

func (up *Updater) unarchiveTar(src io.Reader, releaseBinary string) (io.Reader, error) {
	t := tar.NewReader(src)
	for {
		h, err := t.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%w tar file: %s", ErrCannotDecompressFile, err)
		}
		_, target := filepath.Split(h.Name)
		if matchExecutableName(releaseBinary, up.os, up.arch, target) {
			up.logger.Printf("Executable file %q was found in tar archive", h.Name)
			return t, nil
		}
	}
	return nil, fmt.Errorf("%w in tar: %q", ErrExecutableNotFoundInArchive, releaseBinary)
}
