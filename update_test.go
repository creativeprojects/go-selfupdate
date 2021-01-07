package selfupdate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestBinary() {
	if err := exec.Command("go", "build", "./testdata/github-release-test/").Run(); err != nil {
		panic(err)
	}
}

func teardownTestBinary() {
	bin := "github-release-test"
	if runtime.GOOS == "windows" {
		bin = "github-release-test.exe"
	}
	if err := os.Remove(bin); err != nil {
		panic(err)
	}
}

func TestUpdateCommandWithWrongVersion(t *testing.T) {
	_, err := UpdateCommand("path", "wrong version", "test/test")
	assert.Error(t, err)
}

func TestUpdateCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skip tests in short mode.")
	}

	for _, slug := range []string{
		"rhysd-test/test-release-zip",
		"rhysd-test/test-release-tar",
		"rhysd-test/test-release-gzip",
		"rhysd-test/test-release-tar-xz",
		"rhysd-test/test-release-xz",
		"rhysd-test/test-release-contain-version",
	} {
		t.Run(slug, func(t *testing.T) {
			setupTestBinary()
			defer teardownTestBinary()
			prev := "1.2.2"
			rel, err := UpdateCommand("github-release-test", prev, slug)
			if err != nil {
				t.Fatal(err)
			}
			if !rel.Equal("1.2.3") {
				t.Error("Version is not latest", rel.Version())
			}
			bytes, err := exec.Command(filepath.FromSlash("./github-release-test")).Output()
			if err != nil {
				t.Fatal("Failed to run test binary after update:", err)
			}
			out := string(bytes)
			if out != "v1.2.3\n" {
				t.Error("Output from test binary after update is unexpected:", out)
			}
		})
	}
}

func TestUpdateViaSymlink(t *testing.T) {
	if testing.Short() {
		t.Skip("skip tests in short mode.")
	}
	if runtime.GOOS == "windows" {
		t.Skip("skipping because creating symlink on windows requires admin privilege")
	}

	setupTestBinary()
	defer teardownTestBinary()
	exePath := "github-release-test"
	symPath := "github-release-test-sym"
	if runtime.GOOS == "windows" {
		exePath = "github-release-test.exe"
		symPath = "github-release-test-sym.exe"
	}
	if err := os.Symlink(exePath, symPath); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(symPath)

	prev := "1.2.2"
	rel, err := UpdateCommand(symPath, prev, "rhysd-test/test-release-zip")
	if err != nil {
		t.Fatal(err)
	}
	if !rel.Equal("1.2.3") {
		t.Error("Version is not latest", rel.Version())
	}

	// Test not symbolic link, but actual physical executable
	bytes, err := exec.Command(filepath.FromSlash("./github-release-test")).Output()
	if err != nil {
		t.Fatal("Failed to run test binary after update:", err)
	}
	out := string(bytes)
	if out != "v1.2.3\n" {
		t.Error("Output from test binary after update is unexpected:", out)
	}

	s, err := os.Lstat(symPath)
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s is not a symlink.", symPath)
	}
	p, err := filepath.EvalSymlinks(symPath)
	if err != nil {
		t.Fatal(err)
	}
	if p != exePath {
		t.Fatal("Created symlink no longer points the executable:", p)
	}
}

func TestUpdateBrokenSymlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping because creating symlink on windows requires admin privilege")
	}

	// unknown-xxx -> unknown-yyy -> {not existing}
	xxx := "unknown-xxx"
	yyy := "unknown-yyy"
	if runtime.GOOS == "windows" {
		xxx = "unknown-xxx.exe"
		yyy = "unknown-yyy.exe"
	}
	if err := os.Symlink("not-existing", yyy); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(yyy)
	if err := os.Symlink(yyy, xxx); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(xxx)

	v := "1.2.2"
	for _, p := range []string{yyy, xxx} {
		_, err := UpdateCommand(p, v, "owner/repo")
		if err == nil {
			t.Fatal("Error should occur for unlinked symlink", p)
		}
		if !strings.Contains(err.Error(), "failed to resolve symlink") {
			t.Fatal("Unexpected error for broken symlink", p, err)
		}
	}
}

func TestNotExistingCommandPath(t *testing.T) {
	_, err := UpdateCommand("not-existing-command-path", "1.2.2", "owner/repo")
	if err == nil {
		t.Fatal("Not existing command path should cause an error")
	}
	if !strings.Contains(err.Error(), "file may not exist") {
		t.Fatal("Unexpected error for not existing command path", err)
	}
}

func TestNoReleaseFoundForUpdate(t *testing.T) {
	v := "1.0.0"
	fake := filepath.FromSlash("./testdata/fake-executable")
	rel, err := UpdateCommand(fake, v, "rhysd/misc")
	skipRateLimitExceeded(t, err)
	if err != nil {
		t.Fatal("No release should not make an error:", err)
	}
	if !rel.Equal("1.0.0") {
		t.Error("No release should return the current version as the latest:", rel.Version())
	}
	if rel.URL != "" {
		t.Error("Browse URL should be empty when no release found:", rel.URL)
	}
	if rel.AssetURL != "" {
		t.Error("Asset URL should be empty when no release found:", rel.AssetURL)
	}
	if rel.ReleaseNotes != "" {
		t.Error("Release notes should be empty when no release found:", rel.ReleaseNotes)
	}
}

func TestCurrentIsTheLatest(t *testing.T) {
	if testing.Short() {
		t.Skip("skip tests in short mode.")
	}
	setupTestBinary()
	defer teardownTestBinary()

	v := "1.2.3"
	rel, err := UpdateCommand("github-release-test", v, "rhysd-test/test-release-zip")
	if err != nil {
		t.Fatal(err)
	}
	if !rel.Equal("1.2.3") {
		t.Error("v1.2.3 should be the latest:", rel.Version())
	}
	if rel.URL == "" {
		t.Error("Browse URL should not be empty when release found:", rel.URL)
	}
	if rel.AssetURL == "" {
		t.Error("Asset URL should not be empty when release found:", rel.AssetURL)
	}
	if rel.ReleaseNotes == "" {
		t.Error("Release notes should not be empty when release found:", rel.ReleaseNotes)
	}
}

func TestBrokenBinaryUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skip tests in short mode.")
	}

	fake := filepath.FromSlash("./testdata/fake-executable")
	_, err := UpdateCommand(fake, "1.2.2", "rhysd-test/test-incorrect-release")
	if err == nil {
		t.Fatal("Error should occur for broken package")
	}
	if !strings.Contains(err.Error(), "failed to decompress .tar.gz file") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestInvalidSlugForUpdate(t *testing.T) {
	fake := filepath.FromSlash("./testdata/fake-executable")
	_, err := UpdateCommand(fake, "1.0.0", "rhysd/")
	assert.EqualError(t, err, ErrInvalidSlug.Error())
}

func TestInvalidAssetURL(t *testing.T) {
	err := UpdateTo("https://github.com/creativeprojects/non-existing-repo/releases/download/v1.2.3/foo.zip", "foo.zip", "foo")
	if err == nil {
		t.Fatal("Error should occur for URL not found")
	}
	if !strings.Contains(err.Error(), "failed to download a release file") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestBrokenAsset(t *testing.T) {
	asset := "https://github.com/rhysd-test/test-incorrect-release/releases/download/invalid/broken-zip.zip"
	err := UpdateTo(asset, "broken-zip.zip", "foo")
	if err == nil {
		t.Fatal("Error should occur for URL not found")
	}
	if !strings.Contains(err.Error(), "failed to decompress zip file") {
		t.Fatal("Unexpected error:", err)
	}
}

func TestBrokenGitHubEnterpriseURL(t *testing.T) {
	source, _ := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: "https://example.com"})
	up, err := NewUpdater(Config{Source: source})
	if err != nil {
		t.Fatal(err)
	}
	err = up.UpdateTo(&Release{AssetURL: "https://example.com", repoOwner: "test", repoName: "test"}, "foo")
	if err == nil {
		t.Fatal("Invalid GitHub Enterprise base URL should raise an error")
	}
	if !strings.Contains(err.Error(), "failed to call GitHub Releases API for getting the asset") {
		t.Error("Unexpected error occurred:", err)
	}
}

func TestUpdateFromGitHubPrivateRepo(t *testing.T) {
	token := os.Getenv("GITHUB_PRIVATE_TOKEN")
	if token == "" {
		t.Skip("because GITHUB_PRIVATE_TOKEN is not set")
	}

	setupTestBinary()
	defer teardownTestBinary()

	source, _ := NewGitHubSource(GitHubConfig{APIToken: token})
	up, err := NewUpdater(Config{Source: source})
	if err != nil {
		t.Fatal(err)
	}

	prev := "1.2.2"
	rel, err := up.UpdateCommand("github-release-test", prev, "rhysd/private-release-test")
	if err != nil {
		t.Fatal(err)
	}

	if !rel.Equal("1.2.3") {
		t.Error("Version is not latest", rel.Version())
	}

	bytes, err := exec.Command(filepath.FromSlash("./github-release-test")).Output()
	if err != nil {
		t.Fatal("Failed to run test binary after update:", err)
	}

	out := string(bytes)
	if out != "v1.2.3\n" {
		t.Error("Output from test binary after update is unexpected:", out)
	}
}

// ======================== Test validate with Mock ============================================

func TestNoValidationFile(t *testing.T) {
	source := &MockSource{}
	release := &Release{
		repoOwner:         "test",
		repoName:          "test",
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
		repoOwner:         "test",
		repoName:          "test",
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
		repoOwner:         "test",
		repoName:          "test",
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
		repoOwner:         "test",
		repoName:          "test",
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
		repoName: "test",
		AssetID:  123,
	}
	err := updater.UpdateTo(release, "")
	assert.EqualError(t, err, ErrIncorrectParameterOwner.Error())
}

func TestUpdateToInvalidRepo(t *testing.T) {
	source := &MockSource{}
	updater := &Updater{source: source}
	release := &Release{
		repoOwner: "test",
		AssetID:   123,
	}
	err := updater.UpdateTo(release, "")
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
		repoOwner: "test",
		repoName:  "test",
		AssetID:   123,
	}
	err := updater.UpdateTo(release, "")
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
		repoOwner:         "test",
		repoName:          "test",
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.zip",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	err = updater.UpdateTo(release, "")
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
		repoOwner:         "test",
		repoName:          "test",
		AssetID:           111,
		ValidationAssetID: 123,
		AssetName:         "foo.tar.xz",
	}
	updater := &Updater{
		source:    source,
		validator: &ChecksumValidator{},
	}

	tempfile, err := createEmptyFile(t, "TestUpdateToSuccess")
	require.NoError(t, err)
	defer os.Remove(tempfile)

	err = updater.UpdateTo(release, tempfile)
	require.NoError(t, err)
}

// createEmptyFile creates an empty file with a unique name in the system temporary folder
func createEmptyFile(t *testing.T, basename string) (string, error) {
	tempfile := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%d.tmp", basename, time.Now().UnixNano(), os.Getpid()))
	t.Logf("use temporary file %q", tempfile)
	file, err := os.OpenFile(tempfile, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return "", err
	}
	file.Close()
	return tempfile, nil
}
