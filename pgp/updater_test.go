package pgp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/memory"
	"github.com/creativeprojects/go-selfupdate/types"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	source := &memory.Source{
		Releases: []types.SourceRelease{
			&memory.Release{
				ReleaseID:   1,
				Name:        "v1.0.0",
				TagName:     "v1.0.0",
				PublishedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Assets: []types.SourceAsset{
					&memory.Asset{ID: archiveID, Name: assetName, Size: len(archive)},
					&memory.Asset{ID: checksumsID, Name: "checksums.txt", Size: len(checksums)},
					&memory.Asset{ID: signatureID, Name: "checksums.txt.asc", Size: signature.Len()},
				},
			},
		},
		Files: map[int64][]byte{
			archiveID:   archive,
			checksumsID: []byte(checksums),
			signatureID: signature.Bytes(),
		},
	}

	updater, err := selfupdate.NewUpdater(types.Config{
		Source:    source,
		Validator: NewChecksumValidator("checksums.txt", keyRing),
		OS:        "linux",
		Arch:      "amd64",
	})
	require.NoError(t, err)

	release, found, err := updater.DetectLatest(context.Background(), selfupdate.DetectLatestOpt{Repository: types.NewRepositorySlug("owner", "new_version")})
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, release.ValidationChain, 2)
	assert.Equal(t, "checksums.txt", release.ValidationChain[0].ValidationAssetName)
	assert.Equal(t, "checksums.txt.asc", release.ValidationChain[1].ValidationAssetName)

	// the executable name must match the file inside the archive
	target := filepath.Join(t.TempDir(), "new_version")
	require.NoError(t, os.WriteFile(target, nil, 0o600))

	t.Run("ValidSignature", func(t *testing.T) {
		assert.NoError(t, updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{Rel: release, CmdPath: target}))
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		source.Files[signatureID] = archive
		err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{Rel: release, CmdPath: target})
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("TamperedChecksums", func(t *testing.T) {
		source.Files[signatureID] = signature.Bytes()
		source.Files[checksumsID] = []byte(checksums + "deadbeef  tampered_file\n")
		err := updater.UpdateTo(context.Background(), selfupdate.UpdateToOpt{Rel: release, CmdPath: target})
		assert.ErrorContains(t, err, "openpgp: invalid signature")
	})
}
