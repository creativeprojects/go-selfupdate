package update

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	oldFile = []byte{0xDE, 0xAD, 0xBE, 0xEF}
	newFile = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06}
)

func cleanup(path string) {
	os.Remove(path)
	os.Remove(fmt.Sprintf(".%s.new", path))
}

// we write with a separate name for each test so that we can run them in parallel
func writeOldFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, oldFile, 0o600); err != nil {
		t.Fatalf("Failed to write file for testing preparation: %v", err)
	}
}

func validateUpdate(t *testing.T, path string, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file post-update: %v", err)
	}

	if !bytes.Equal(buf, newFile) {
		t.Fatalf("File was not updated! Bytes read: %v, Bytes expected: %v", buf, newFile)
	}
}

func TestApplySimple(t *testing.T) {
	t.Parallel()

	fName := t.Name()
	defer cleanup(fName)
	writeOldFile(t, fName)

	err := Apply(bytes.NewReader(newFile), Options{
		TargetPath: fName,
	})
	validateUpdate(t, fName, err)
}

func TestApplyOldSavePath(t *testing.T) {
	t.Parallel()

	fName := t.Name()
	defer cleanup(fName)
	writeOldFile(t, fName)

	oldfName := "OldSavePath"

	err := Apply(bytes.NewReader(newFile), Options{
		TargetPath:  fName,
		OldSavePath: oldfName,
	})
	validateUpdate(t, fName, err)

	if _, err := os.Stat(oldfName); os.IsNotExist(err) {
		t.Fatalf("Failed to find the old file: %v", err)
	}

	cleanup(oldfName)
}

func TestApplyNewFile(t *testing.T) {
	t.Parallel()

	fName := t.Name()
	defer cleanup(fName)
	// no old file: the target doesn't exist yet

	err := Apply(bytes.NewReader(newFile), Options{
		TargetPath: fName,
	})
	validateUpdate(t, fName, err)

	assert.NoFileExists(t, fmt.Sprintf(".%s.old", fName))
	assert.NoFileExists(t, fmt.Sprintf(".%s.new", fName))
}

func TestApplyNewFileOldSavePath(t *testing.T) {
	t.Parallel()

	fName := t.Name()
	defer cleanup(fName)
	// no old file: the target doesn't exist yet

	oldfName := "NewFileOldSavePath"
	defer cleanup(oldfName)

	err := Apply(bytes.NewReader(newFile), Options{
		TargetPath:  fName,
		OldSavePath: oldfName,
	})
	validateUpdate(t, fName, err)

	// there was nothing to save
	assert.NoFileExists(t, oldfName)
}

func TestWriteError(t *testing.T) {
	// fix this test patching the global openFile variable
	// t.Parallel()

	fName := t.Name()
	defer cleanup(fName)
	writeOldFile(t, fName)

	openFile = func(name string, flags int, perm os.FileMode) (*os.File, error) {
		f, err := os.OpenFile(name, flags, perm)

		// simulate Write() error by closing the file prematurely
		f.Close()

		return f, err
	}
	defer func() {
		openFile = os.OpenFile
	}()

	err := Apply(bytes.NewReader(newFile), Options{TargetPath: fName})
	if err == nil {
		t.Fatalf("Allowed an update to an empty file")
	}
}
