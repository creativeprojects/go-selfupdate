package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCommandWithWrongVersion(t *testing.T) {
	_, err := UpdateCommand(context.Background(), "path", "wrong version", ParseSlug("test/test"))
	assert.Error(t, err)
	assert.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func TestUpdateCommand(t *testing.T) {
	current := "0.10.0"
	new := "1.0.0"
	source := mockSourceRepository(t)
	updater, err := NewUpdater(Config{Source: source})
	require.NoError(t, err)

	filename := setupCurrentVersion(t)

	rel, err := updater.UpdateCommand(context.Background(), filename, current, ParseSlug("creativeprojects/new_version"))
	require.NoError(t, err)
	assert.Equal(t, new, rel.Version())

	assertNewVersion(t, filename)
}

func TestUpdateViaSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping because creating symlink on windows requires admin privilege")
	}

	current := "0.10.0"
	new := "1.0.0"
	source := mockSourceRepository(t)
	updater, err := NewUpdater(Config{Source: source})
	require.NoError(t, err)

	exePath := setupCurrentVersion(t)
	symPath := exePath + "-sym"

	err = os.Symlink(exePath, symPath)
	require.NoError(t, err)

	rel, err := updater.UpdateCommand(context.Background(), symPath, current, ParseSlug("creativeprojects/new_version"))
	require.NoError(t, err)
	assert.Equal(t, new, rel.Version())

	// check actual file (not symlink)
	assertNewVersion(t, exePath)

	s, err := os.Lstat(symPath)
	require.NoError(t, err)
	if s.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink.", symPath)
	}
	// check symlink
	assertNewVersion(t, symPath)
}

func TestUpdateBrokenSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping because creating symlink on windows requires admin privilege")
	}

	updater, err := NewUpdater(Config{Source: mockSourceRepository(t)})
	require.NoError(t, err)

	// unknown-xxx -> unknown-yyy -> {not existing}
	xxx := "unknown-xxx"
	yyy := "unknown-yyy"

	err = os.Symlink("not-existing", yyy)
	require.NoError(t, err)
	defer os.Remove(yyy)

	err = os.Symlink(yyy, xxx)
	require.NoError(t, err)
	defer os.Remove(xxx)

	for _, filename := range []string{yyy, xxx} {
		_, err := updater.UpdateCommand(context.Background(), filename, "0.10.0", ParseSlug("owner/repo"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve symlink")
	}
}

func TestNotExistingCommandPath(t *testing.T) {
	_, err := UpdateCommand(context.Background(), "not-existing-command-path", "1.2.2", ParseSlug("owner/repo"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file may not exist")
}

func TestNoReleaseFoundForUpdate(t *testing.T) {
	finalVersion := "1.0.0"
	fake := filepath.FromSlash("./testdata/fake-executable")
	updater, err := NewUpdater(Config{Source: &MockSource{}})
	require.NoError(t, err)

	rel, err := updater.UpdateCommand(context.Background(), fake, finalVersion, ParseSlug("owner/repo"))
	assert.NoError(t, err)
	assert.Equal(t, finalVersion, rel.Version())
	assert.Empty(t, rel.URL)
	assert.Empty(t, rel.AssetURL)
	assert.Empty(t, rel.ReleaseNotes)
}

func TestCurrentIsTheLatest(t *testing.T) {
	filename := setupCurrentVersion(t)

	updater, err := NewUpdater(Config{Source: mockSourceRepository(t)})
	require.NoError(t, err)

	latest := "1.0.0"
	rel, err := updater.UpdateCommand(context.Background(), filename, latest, ParseSlug("creativeprojects/new_version"))
	assert.NoError(t, err)
	assert.Equal(t, latest, rel.Version())
	assert.NotEmpty(t, rel.URL)
	assert.NotEmpty(t, rel.AssetURL)
	assert.NotEmpty(t, rel.ReleaseNotes)
}

func TestBrokenBinaryUpdate(t *testing.T) {
	fake := filepath.FromSlash("./testdata/fake-executable")

	source := NewMockSource([]SourceRelease{
		&GitHubRelease{
			name:        "v2.0.0",
			tagName:     "v2.0.0",
			url:         "v2.0.0",
			publishedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			assets: []SourceAsset{
				&GitHubAsset{
					id:   1,
					name: "invalid_v2.0.0_linux_amd64.tar.gz",
					url:  "invalid_v2.0.0_linux_amd64.tar.gz",
					size: len("invalid content"),
				},
				&GitHubAsset{
					id:   2,
					name: "invalid_v2.0.0_darwin_amd64.tar.gz",
					url:  "invalid_v2.0.0_darwin_amd64.tar.gz",
					size: len("invalid content"),
				},
				&GitHubAsset{
					id:   3,
					name: "invalid_v2.0.0_windows_amd64.zip",
					url:  "invalid_v2.0.0_windows_amd64.zip",
					size: len("invalid content"),
				},
			},
		},
	}, map[int64][]byte{
		1: []byte("invalid content"),
		2: []byte("invalid content"),
		3: []byte("invalid content"),
	})

	updater, err := NewUpdater(Config{Source: source})
	require.NoError(t, err)

	_, err = updater.UpdateCommand(context.Background(), fake, "1.2.2", ParseSlug("rhysd-test/test-incorrect-release"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decompress")
}

func TestInvalidSlugForUpdate(t *testing.T) {
	fake := filepath.FromSlash("./testdata/fake-executable")
	_, err := UpdateCommand(context.Background(), fake, "1.0.0", ParseSlug("rhysd/"))
	assert.Error(t, err)
}

func TestInvalidAssetURL(t *testing.T) {
	err := UpdateTo("https://github.com/creativeprojects/non-existing-repo/releases/download/v1.2.3/foo.zip", "foo.zip", "foo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download a release file")
}

func TestBrokenAsset(t *testing.T) {
	asset := "https://github.com/rhysd-test/test-incorrect-release/releases/download/invalid/broken-zip.zip"
	err := UpdateTo(asset, "broken-zip.zip", "foo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decompress zip file")
}

func TestBrokenGitHubEnterpriseURL(t *testing.T) {
	source, _ := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: "https://example.com"})
	up, err := NewUpdater(Config{Source: source})
	assert.NoError(t, err)

	err = up.UpdateTo(
		context.Background(),
		&Release{AssetURL: "https://example.com",
			repository: NewRepositorySlug("test", "test")},
		"foo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call GitHub Releases API for getting the asset")
}

// ======================== Test validate with Mock ============================================

func TestNoValidationFile(t *testing.T) {
	source := &MockSource{}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
	}
	updater := &Updater{
		source: source,
	}
	err := updater.validate(release, []byte("some data"))
	assert.EqualError(t, err, ErrAssetNotFound.Error())
}

func TestValidationWrongHash(t *testing.T) {
	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &MockSource{
		files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.zip",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.validate(release, data)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrChecksumValidationFailed), "Not the error we expected")
}

func TestValidationReadError(t *testing.T) {
	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &MockSource{
		readError: true,
		files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.validate(release, data)
	require.Error(t, err)
	assert.True(t, errors.Is(err, errTestRead))
}

func TestValidationSuccess(t *testing.T) {
	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &MockSource{
		files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.validate(release, data)
	require.NoError(t, err)
}

// ======================== Test UpdateTo with Mock ==========================================

func TestUpdateToInvalidOwner(t *testing.T) {
	source := &MockSource{}
	updater := &Updater{source: source}
	release := &Release{
		repository: NewRepositorySlug("", "test"),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), release, "")
	assert.EqualError(t, err, ErrIncorrectParameterOwner.Error())
}

func TestUpdateToInvalidRepo(t *testing.T) {
	source := &MockSource{}
	updater := &Updater{source: source}
	release := &Release{
		repository: NewRepositorySlug("test", ""),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), release, "")
	assert.EqualError(t, err, ErrIncorrectParameterRepo.Error())
}

func TestUpdateToReadError(t *testing.T) {
	source := &MockSource{
		readError: true,
		files: map[int64][]byte{
			123: []byte("some data"),
		},
	}
	updater := &Updater{source: source}
	release := &Release{
		repository: NewRepositorySlug("test", "test"),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), release, "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, errTestRead))
}

func TestUpdateToWithWrongHash(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &MockSource{
		files: map[int64][]byte{
			111: data,
			123: hashData,
		},
	}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.zip",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	err = updater.UpdateTo(context.Background(), release, "")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrChecksumValidationFailed))
}

func TestUpdateToSuccess(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &MockSource{
		files: map[int64][]byte{
			111: data,
			123: hashData,
		},
	}
	release := &Release{
		repository:        NewRepositorySlug("test", "test"),
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	tempfile, err := createEmptyFile(t, "foo")
	require.NoError(t, err)
	defer os.Remove(tempfile)

	err = updater.UpdateTo(context.Background(), release, tempfile)
	require.NoError(t, err)
}

// createEmptyFile creates an empty file with a unique name in the system temporary folder
func createEmptyFile(t *testing.T, basename string) (string, error) {
	t.Helper()
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s", basename))
	t.Logf("use temporary file %q", tempfile)
	file, err := os.OpenFile(tempfile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	file.Close()
	return tempfile, nil
}

func setupCurrentVersion(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "new_version")
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	err := os.WriteFile(filename, []byte("old version"), 0o777)
	require.NoError(t, err)

	return filename
}

func assertNewVersion(t *testing.T, filename string) {
	bytes, err := os.ReadFile(filename)
	require.NoError(t, err)

	assert.Equal(t, []byte("new version!\n"), bytes)
}
