package pgp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/creativeprojects/go-selfupdate"
)

// memorySource is an in-memory selfupdate.Source
type memorySource struct {
	releases []selfupdate.SourceRelease
	files    map[int64][]byte
}

func (s *memorySource) ListReleases(_ context.Context, _ selfupdate.Repository) ([]selfupdate.SourceRelease, error) {
	return s.releases, nil
}

func (s *memorySource) DownloadReleaseAsset(_ context.Context, _ *selfupdate.Release, assetID int64) (io.ReadCloser, error) {
	content, ok := s.files[assetID]
	if !ok {
		return nil, selfupdate.ErrAssetNotFound
	}
	return io.NopCloser(bytes.NewReader(content)), nil
}

func TestUpdateWithSignedChecksums(t *testing.T) {
	const (
		archiveID   = 1
		checksumsID = 2
		signatureID = 3
		assetName   = "new_version_1.0.0_linux_amd64.tar.gz"
	)
	keyRing, entity := getTestKeyRing(t)

	archive, err := os.ReadFile("../testdata/new_version.tar.gz")
	require.NoError(t, err)

	checksums := fmt.Sprintf("%x  %s\n", sha256.Sum256(archive), assetName)
	signature := &bytes.Buffer{}
	require.NoError(t, openpgp.ArmoredDetachSign(signature, entity, strings.NewReader(checksums), nil))

	source := &memorySource{
		releases: []selfupdate.SourceRelease{
			&selfupdate.HttpRelease{
				ID:          1,
				Name:        "v1.0.0",
				TagName:     "v1.0.0",
				PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Assets: []*selfupdate.HttpAsset{
					{ID: archiveID, Name: assetName, Size: len(archive)},
					{ID: checksumsID, Name: "checksums.txt", Size: len(checksums)},
					{ID: signatureID, Name: "checksums.txt.asc", Size: signature.Len()},
				},
			},
		},
		files: map[int64][]byte{
			archiveID:   archive,
			checksumsID: []byte(checksums),
			signatureID: signature.Bytes(),
		},
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: NewChecksumValidator("checksums.txt", keyRing),
		OS:        "linux",
		Arch:      "amd64",
	})
	require.NoError(t, err)

	release, found, err := updater.DetectLatest(context.Background(), selfupdate.NewRepositorySlug("owner", "new_version"))
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, release.ValidationChain, 2)
	assert.Equal(t, "checksums.txt", release.ValidationChain[0].ValidationAssetName)
	assert.Equal(t, "checksums.txt.asc", release.ValidationChain[1].ValidationAssetName)

	// the executable name must match the file inside the archive
	target := filepath.Join(t.TempDir(), "new_version")
	require.NoError(t, os.WriteFile(target, nil, 0o600))

	t.Run("ValidSignature", func(t *testing.T) {
		assert.NoError(t, updater.UpdateTo(context.Background(), release, target))
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		source.files[signatureID] = archive
		err := updater.UpdateTo(context.Background(), release, target)
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("TamperedChecksums", func(t *testing.T) {
		source.files[signatureID] = signature.Bytes()
		source.files[checksumsID] = []byte(checksums + "deadbeef  tampered_file\n")
		err := updater.UpdateTo(context.Background(), release, target)
		assert.ErrorContains(t, err, "openpgp: invalid signature")
	})
}
