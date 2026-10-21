package selfupdate_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/github"
	"github.com/creativeprojects/go-selfupdate/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCommandWithWrongVersion(t *testing.T) {
	_, err := newMockUpdater(t, selfupdate.Config{}).UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"file", "path", "wrong version", selfupdate.ParseSlug("test/test")})
	assert.Error(t, err)
	assert.ErrorIs(t, err, semver.ErrInvalidSemVer)
}

func TestUpdateCommand(t *testing.T) {
	currentVersion := "0.14.0"
	newVersion := "1.0.0"
	source := mockMemorySource(t)
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source})
	require.NoError(t, err)

	filename := setupCurrentVersion(t)

	rel, err := updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"new_version", filename, currentVersion, selfupdate.ParseSlug("creativeprojects/new_version")})
	require.NoError(t, err)
	assert.Equal(t, newVersion, rel.Version())

	assertNewVersion(t, filename)
}

func TestUpdateViaSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping because creating symlink on windows requires admin privilege")
	}

	currentVersion := "0.14.0"
	newVersion := "1.0.0"
	source := mockMemorySource(t)
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source})
	require.NoError(t, err)

	exePath := setupCurrentVersion(t)
	symPath := exePath + "-sym"

	err = os.Symlink(exePath, symPath)
	require.NoError(t, err)

	rel, err := updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"new_version", symPath, currentVersion, selfupdate.ParseSlug("creativeprojects/new_version")})
	require.NoError(t, err)
	assert.Equal(t, newVersion, rel.Version())

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

	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: mockMemorySource(t)})
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
		_, err := updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{filename, filename, "0.14.0", selfupdate.ParseSlug("owner/repo")})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve symlink")
	}
}

func TestNotExistingCommandPath(t *testing.T) {
	_, err := newMockUpdater(t, selfupdate.Config{}).UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"not-existing-exe-name", "not-existing-command-path", "1.2.2", selfupdate.ParseSlug("owner/repo")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file may not exist")
}

func TestUpdateToInvalidRelease(t *testing.T) {
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: &memory.Source{}})
	require.NoError(t, err)

	err = updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{nil, "", ""})
	assert.ErrorIs(t, err, selfupdate.ErrInvalidRelease)
}

func TestNoReleaseFoundForUpdate(t *testing.T) {
	finalVersion := "1.0.0"
	fake := filepath.FromSlash("./testdata/fake-executable")
	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: &memory.Source{}})
	require.NoError(t, err)

	rel, err := updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"fake-executable", fake, finalVersion, selfupdate.ParseSlug("owner/repo")})
	assert.NoError(t, err)
	assert.Equal(t, finalVersion, rel.Version())
	assert.Empty(t, rel.URL)
	assert.Empty(t, rel.AssetURL)
	assert.Empty(t, rel.ReleaseNotes)
}

func TestCurrentIsTheLatest(t *testing.T) {
	filename := setupCurrentVersion(t)

	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: mockMemorySource(t)})
	require.NoError(t, err)

	latest := "1.0.0"
	rel, err := updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"new_version", filename, latest, selfupdate.ParseSlug("creativeprojects/new_version")})
	assert.NoError(t, err)
	assert.Equal(t, latest, rel.Version())
	assert.NotEmpty(t, rel.URL)
	assert.NotEmpty(t, rel.AssetURL)
	assert.NotEmpty(t, rel.ReleaseNotes)
}

func TestBrokenBinaryUpdate(t *testing.T) {
	fake := filepath.FromSlash("./testdata/fake-executable")

	source := newMemorySource([]selfupdate.SourceRelease{
		&memory.Release{
			Name:        "v2.0.0",
			TagName:     "v2.0.0",
			URL:         "v2.0.0",
			PublishedAt: time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC),
			Assets: []selfupdate.SourceAsset{
				&memory.Asset{
					ID:   1,
					Name: "invalid_v2.0.0_linux_amd64.tar.gz",
					URL:  "invalid_v2.0.0_linux_amd64.tar.gz",
					Size: len("invalid content"),
				},
				&memory.Asset{
					ID:   2,
					Name: "invalid_v2.0.0_darwin_amd64.tar.gz",
					URL:  "invalid_v2.0.0_darwin_amd64.tar.gz",
					Size: len("invalid content"),
				},
				&memory.Asset{
					ID:   3,
					Name: "invalid_v2.0.0_darwin_arm64.tar.gz",
					URL:  "invalid_v2.0.0_darwin_arm64.tar.gz",
					Size: len("invalid content"),
				},
				&memory.Asset{
					ID:   4,
					Name: "invalid_v2.0.0_windows_amd64.zip",
					URL:  "invalid_v2.0.0_windows_amd64.zip",
					Size: len("invalid content"),
				},
			},
		},
	}, map[int64][]byte{
		1: []byte("invalid content"),
		2: []byte("invalid content"),
		3: []byte("invalid content"),
		4: []byte("invalid content"),
	})

	updater, err := selfupdate.NewUpdater(selfupdate.Config{Source: source})
	require.NoError(t, err)

	_, err = updater.UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"fake-executable", fake, "1.2.2", selfupdate.ParseSlug("rhysd-test/test-incorrect-release")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decompress")
}

func TestInvalidSlugForUpdate(t *testing.T) {
	fake := filepath.FromSlash("./testdata/fake-executable")
	_, err := newMockUpdater(t, selfupdate.Config{}).UpdateCommand(context.Background(), selfupdate.UpdateCommandOpt{"fake-executable", fake, "1.0.0", selfupdate.ParseSlug("rhysd/")})
	assert.Error(t, err)
}

func TestBrokenGitHubEnterpriseURL(t *testing.T) {
	source, _ := github.NewSource(github.Config{APIToken: "my_token", EnterpriseBaseURL: "https://example.com"})
	up, err := selfupdate.NewUpdater(selfupdate.Config{Source: source})
	assert.NoError(t, err)

	err = up.UpdateTo(
		context.Background(),
		selfupdate.UpdateToOpt{&selfupdate.Release{AssetURL: "https://example.com",
			Repository: selfupdate.NewRepositorySlug("test", "test")},
			"foo",
			"foo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call GitHub Releases API for getting the asset")
}

// ======================== Test validate with Mock ============================================

func TestNoValidationFile(t *testing.T) {
	source := &memory.Source{}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source})
	err := updater.Validate(context.Background(), release, []byte("some data"))
	assert.EqualError(t, err, fmt.Sprintf("failed reading validation data \"\": %s", selfupdate.ErrAssetNotFound.Error()))
	assert.ErrorIs(t, err, selfupdate.ErrAssetNotFound)
}

func TestValidationWrongHash(t *testing.T) {
	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &memory.Source{
		Files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.zip",
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source, Validator: &selfupdate.ChecksumValidator{}})

	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.Validate(context.Background(), release, data)
	require.Error(t, err)
	assert.True(t, errors.Is(err, selfupdate.ErrChecksumValidationFailed), "Not the error we expected")
}

func TestValidationReadError(t *testing.T) {
	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &memory.Source{
		ReadError: true,
		Files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source, Validator: &selfupdate.ChecksumValidator{}})

	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.Validate(context.Background(), release, data)
	require.Error(t, err)
	assert.True(t, errors.Is(err, memory.ErrRead))
}

func TestValidationSuccess(t *testing.T) {
	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &memory.Source{
		Files: map[int64][]byte{
			123: hashData,
		},
	}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source, Validator: &selfupdate.ChecksumValidator{}})

	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	err = updater.Validate(context.Background(), release, data)
	require.NoError(t, err)
}

// ======================== Test UpdateTo with Mock ==========================================

func TestUpdateToInvalidOwner(t *testing.T) {
	source := &memory.Source{}
	updater := newMockUpdater(t, selfupdate.Config{Source: source})
	release := &selfupdate.Release{
		Repository: selfupdate.NewRepositorySlug("", "test"),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "", ""})
	assert.EqualError(t, err, fmt.Sprintf("failed to read asset \"\": %s", selfupdate.ErrIncorrectParameterOwner.Error()))
	assert.ErrorIs(t, err, selfupdate.ErrIncorrectParameterOwner)
}

func TestUpdateToInvalidRepo(t *testing.T) {
	source := &memory.Source{}
	updater := newMockUpdater(t, selfupdate.Config{Source: source})
	release := &selfupdate.Release{
		Repository: selfupdate.NewRepositorySlug("test", ""),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "", ""})
	assert.EqualError(t, err, fmt.Sprintf("failed to read asset \"\": %s", selfupdate.ErrIncorrectParameterRepo.Error()))
	assert.ErrorIs(t, err, selfupdate.ErrIncorrectParameterRepo)
}

func TestUpdateToReadError(t *testing.T) {
	source := &memory.Source{
		ReadError: true,
		Files: map[int64][]byte{
			123: []byte("some data"),
		},
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source})
	release := &selfupdate.Release{
		Repository: selfupdate.NewRepositorySlug("test", "test"),
		AssetID:    123,
	}
	err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "", ""})
	require.Error(t, err)
	assert.True(t, errors.Is(err, memory.ErrRead))
}

func TestUpdateToWithWrongHash(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &memory.Source{
		Files: map[int64][]byte{
			111: data,
			123: hashData,
		},
	}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.zip",
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source, Validator: &selfupdate.ChecksumValidator{}})

	err = updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "", ""})
	require.Error(t, err)
	assert.True(t, errors.Is(err, selfupdate.ErrChecksumValidationFailed))
}

func TestUpdateToSuccess(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	source := &memory.Source{
		Files: map[int64][]byte{
			111: data,
			123: hashData,
		},
	}
	release := &selfupdate.Release{
		Repository:        selfupdate.NewRepositorySlug("test", "test"),
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := newMockUpdater(t, selfupdate.Config{Source: source, Validator: &selfupdate.ChecksumValidator{}})

	tempfile := createEmptyFile(t, "foo")

	err = updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "foo", tempfile})
	require.NoError(t, err)
}

func TestUpdateToWithMultistepValidationChain(t *testing.T) {
	testVersion := "v0.14.0"
	source := mockSignedSourceRepository(t)
	updater, _ := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: newChecksumWithSignatureValidator("checksums.txt"),
	})

	tempFile := createEmptyFile(t, "new_version")

	getRelease := func(t *testing.T) *selfupdate.Release {
		release, found, err := updater.DetectVersion(context.Background(), selfupdate.DetectVersionOpt{Repository: testGithubRepository, Version: testVersion})
		require.NotNil(t, release)
		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, 2, len(release.ValidationChain))
		return release
	}

	t.Run("Succeeds", func(t *testing.T) {
		release := getRelease(t)
		err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "new_version", tempFile})
		assert.NoError(t, err)
	})

	t.Run("ValidationFailInStep1", func(t *testing.T) {
		release := getRelease(t)
		release.ValidationChain[0].ValidationAssetID = 1

		err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "new_version", tempFile})
		assert.EqualError(t, err, fmt.Sprintf("failed validating asset content %q: incorrect checksum file format", release.AssetName))
	})

	t.Run("ValidationFailInStep2", func(t *testing.T) {
		release := getRelease(t)
		release.ValidationChain[1].ValidationAssetID = 1

		err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{release, "new_version", tempFile})
		assert.ErrorContains(t, err, "failed validating asset content \"checksums.txt\"")
		assert.ErrorIs(t, err, selfupdate.ErrChecksumValidationFailed)
	})
}

// createEmptyFile creates an empty file with a unique name in the system temporary folder
func createEmptyFile(t *testing.T, filename string) string {
	t.Helper()
	tempfile := filepath.Join(t.TempDir(), filename)
	t.Logf("use temporary file %q", tempfile)
	file, err := os.OpenFile(tempfile, os.O_WRONLY|os.O_CREATE, 0777)
	if err == nil {
		err = file.Close()
	}
	require.NoError(t, err)
	return tempfile
}

func setupCurrentVersion(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "new_version")
	if runtime.GOOS == "windows" {
		filename += ".exe"
	}

	err := os.WriteFile(filename, []byte("old version"), 0o600)
	require.NoError(t, err)

	return filename
}

func assertNewVersion(t *testing.T, filename string) {
	bytes, err := os.ReadFile(filename)
	require.NoError(t, err)

	assert.Equal(t, []byte("new version!\n"), bytes)
}
