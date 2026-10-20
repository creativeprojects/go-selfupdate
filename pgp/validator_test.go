package pgp

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/creativeprojects/go-selfupdate/types"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorAssetNames(t *testing.T) {
	filename := "asset"
	for _, test := range []struct {
		validator      types.Validator
		validationName string
	}{
		{
			validator:      &Validator{},
			validationName: filename + ".asc",
		},
		{
			validator:      &Validator{Binary: true},
			validationName: filename + ".sig",
		},
	} {
		assert.Equal(t, test.validationName, test.validator.GetValidationAssetName(filename))
	}
}

func TestValidator(t *testing.T) {
	data, err := os.ReadFile("../testdata/foo.zip")
	require.NoError(t, err)

	otherData, err := os.ReadFile("../testdata/foo.tar.xz")
	require.NoError(t, err)

	keyRing, entity := getTestKeyRing(t)
	require.NotNil(t, keyRing)
	require.NotNil(t, entity)

	var signatureData []byte
	{
		signature := &bytes.Buffer{}
		err = openpgp.ArmoredDetachSign(signature, entity, bytes.NewReader(data), nil)
		require.NoError(t, err)
		signatureData = signature.Bytes()
	}

	t.Run("NoPublicKey", func(t *testing.T) {
		validator := new(Validator)
		err = validator.Validate("foo.zip", data, signatureData)
		assert.ErrorIs(t, err, ErrKeyRingNotSet)
		err = validator.Validate("foo.zip", data, nil)
		assert.ErrorIs(t, err, ErrKeyRingNotSet)
		err = validator.Validate("foo.zip", data, []byte{})
		assert.ErrorIs(t, err, ErrKeyRingNotSet)
	})

	t.Run("EmptySignature", func(t *testing.T) {
		validator := new(Validator).WithArmoredKeyRing(keyRing)
		err = validator.Validate("foo.zip", data, nil)
		assert.ErrorIs(t, err, ErrInvalidSignature)
		err = validator.Validate("foo.zip", data, []byte{})
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		validator := new(Validator).WithArmoredKeyRing(keyRing)
		err = validator.Validate("foo.zip", data, []byte{0, 1, 2})
		assert.ErrorIs(t, err, ErrInvalidSignature)
		err = validator.Validate("foo.zip", data, data)
		assert.ErrorIs(t, err, ErrInvalidSignature)
	})

	t.Run("ValidSignature", func(t *testing.T) {
		validator := new(Validator).WithArmoredKeyRing(keyRing)
		err = validator.Validate("foo.zip", data, signatureData)
		assert.NoError(t, err)
	})

	t.Run("Fail", func(t *testing.T) {
		validator := new(Validator).WithArmoredKeyRing(keyRing)
		err = validator.Validate("foo.tar.xz", otherData, signatureData)
		// Assert the error category rather than the exact wording: the reason
		// reported after "invalid signature" is an implementation detail of the
		// openpgp library and is not part of this package's contract.
		assert.ErrorContains(t, err, "openpgp: invalid signature")
	})
}

func TestValidatorWithArmoredKeyRing(t *testing.T) {
	keyRing, entity := getTestKeyRing(t)
	validator := new(Validator).WithArmoredKeyRing(keyRing)
	assert.Equal(t, entity.PrimaryKey.KeyIdString(), validator.KeyRing[0].PrimaryKey.KeyIdString())

	assert.PanicsWithError(t, "failed setting armored public key ring: openpgp: invalid argument: no armored data found", func() {
		new(Validator).WithArmoredKeyRing([]byte{})
	})
}

func TestNewChecksumValidator(t *testing.T) {
	keyRing, _ := getTestKeyRing(t)

	validator := NewChecksumValidator("checksums", keyRing)
	assert.Implements(t, (*types.RecursiveValidator)(nil), validator)
	assert.Equal(t, "checksums", validator.GetValidationAssetName("anything"))
	assert.Equal(t, "checksums.asc", validator.GetValidationAssetName("checksums"))
}

func getTestKeyRing(t *testing.T) (keyRing []byte, entity *openpgp.Entity) {
	var err error
	var armoredWriter io.WriteCloser
	entity, err = openpgp.NewEntity("go-selfupdate", "", "info@go-types.local", nil)
	require.NoError(t, err)

	buffer := &bytes.Buffer{}
	if armoredWriter, err = armor.Encode(buffer, openpgp.PublicKeyType, nil); err == nil {
		if err = entity.Serialize(armoredWriter); err == nil {
			err = armoredWriter.Close()
		}
	}
	require.NoError(t, err)
	keyRing = buffer.Bytes()
	return
}
