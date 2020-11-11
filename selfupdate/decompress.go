package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

var (
	fileTypes = []struct {
		ext        string
		decompress func(src io.Reader, url, cmd, os, arch string) (io.Reader, error)
	}{
		{".zip", unzip},
		{".tar.gz", untar},
		{".tgz", untar},
		{".gzip", gunzip},
		{".gz", gunzip},
		{".tar.xz", untarxz},
		{".xz", unxz},
		{".bz2", unbz2},
	}
)

// DecompressCommand decompresses the given source. Archive and compression format is
// automatically detected from 'url' parameter, which represents the URL of asset.
// This returns a reader for the decompressed command given by 'cmd'. '.zip',
// '.tar.gz', '.tar.xz', '.tgz', '.gz', '.bz2' and '.xz' are supported.
func DecompressCommand(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	for _, fileType := range fileTypes {
		if strings.HasSuffix(url, fileType.ext) {
			return fileType.decompress(src, url, cmd, os, arch)
		}
	}
	log.Printf("File %q is not compressed", url)
	return src, nil
}

func unzip(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing zip file %s", url)

	// Zip format requires its file size for Decompressing.
	// So we need to read the HTTP response into a buffer at first.
	buf, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to create buffer for zip file: %s", err)
	}

	r := bytes.NewReader(buf)
	z, err := zip.NewReader(r, r.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to decompress zip file: %s", err)
	}

	for _, file := range z.File {
		_, name := filepath.Split(file.Name)
		if !file.FileInfo().IsDir() && matchExecutableName(cmd, os, arch, name) {
			log.Printf("Executable file %q was found in zip archive", file.Name)
			return file.Open()
		}
	}

	return nil, fmt.Errorf("file %q for the command is not found in %s", cmd, url)
}

func untar(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing tar.gz file %s", url)

	gz, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress .tar.gz file: %s", err)
	}

	return unarchiveTar(gz, url, cmd, os, arch)
}

func gunzip(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing gzip file %s", url)

	r, err := gzip.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress gzip file downloaded from %s: %s", url, err)
	}

	name := r.Header.Name
	if !matchExecutableName(cmd, os, arch, name) {
		return nil, fmt.Errorf("file name '%s' does not match to command '%s' found in %s", name, cmd, url)
	}

	log.Printf("Executable file %q was found in gzip file", name)
	return r, nil
}

func untarxz(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing tar.xz file %s", url)

	xzip, err := xz.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress .tar.xz file: %s", err)
	}

	return unarchiveTar(xzip, url, cmd, os, arch)
}

func unxz(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing xzip file %s", url)

	xzip, err := xz.NewReader(src)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress xzip file downloaded from %s: %s", url, err)
	}

	log.Printf("Decompressed file from xzip is assumed to be an executable: %s", cmd)
	return xzip, nil
}

func unbz2(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	log.Printf("Decompressing bzip2 file %s", url)

	bz2 := bzip2.NewReader(src)

	log.Printf("Decompressed file from bzip2 is assumed to be an executable: %s", cmd)
	return bz2, nil
}

func matchExecutableName(cmd, os, arch, target string) bool {
	if cmd == target {
		return true
	}

	// When the contained executable name is full name (e.g. foo_darwin_amd64),
	// it is also regarded as a target executable file.
	for _, delimiter := range []rune{'_', '-'} {
		c := fmt.Sprintf("%s%c%s%c%s", cmd, delimiter, os, delimiter, arch)
		if os == "windows" {
			c += ".exe"
		}
		if c == target {
			return true
		}
	}

	return false
}

func unarchiveTar(src io.Reader, url, cmd, os, arch string) (io.Reader, error) {
	t := tar.NewReader(src)
	for {
		h, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to unarchive .tar file: %s", err)
		}
		_, name := filepath.Split(h.Name)
		if matchExecutableName(cmd, os, arch, name) {
			log.Printf("Executable file %q was found in tar archive", h.Name)
			return t, nil
		}
	}

	return nil, fmt.Errorf("file %q for the command is not found in %s", cmd, url)
}
