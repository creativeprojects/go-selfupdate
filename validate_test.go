package selfupdate

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorAssetNames(t *testing.T) {
	filename := "asset"
	for _, test := range []struct {
		validator      Validator
		validationName string
	}{
		{
			validator:      &SHAValidator{},
			validationName: filename + ".sha256",
		},
		{
			validator:      &ECDSAValidator{},
			validationName: filename + ".sig",
		},
		{
			validator:      &ChecksumValidator{"funny_sha256"},
			validationName: "funny_sha256",
		},
	} {
		want := test.validationName
		got := test.validator.GetValidationAssetName(filename)
		if want != got {
			t.Errorf("Wanted %q but got %q", want, got)
		}
	}
}

// ======= SHAValidator ====================================================

func TestSHAValidatorEmptyFile(t *testing.T) {
	validator := &SHAValidator{}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)
	err = validator.Validate("foo.zip", data, nil)
	assert.EqualError(t, err, ErrIncorrectChecksumFile.Error())
}

func TestSHAValidatorInvalidFile(t *testing.T) {
	validator := &SHAValidator{}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)
	err = validator.Validate("foo.zip", data, []byte("blahblahblah\n"))
	assert.EqualError(t, err, ErrIncorrectChecksumFile.Error())
}

func TestSHAValidator(t *testing.T) {
	validator := &SHAValidator{}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	err = validator.Validate("foo.zip", data, hashData)
	assert.NoError(t, err)
}

func TestSHAValidatorFail(t *testing.T) {
	validator := &SHAValidator{}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	hashData[0] = '0'
	err = validator.Validate("foo.zip", data, hashData)
	assert.EqualError(t, err, ErrChecksumValidationFailed.Error())
}

// ======= ECDSAValidator ====================================================

func TestECDSAValidatorNoPublicKey(t *testing.T) {
	validator := &ECDSAValidator{
		PublicKey: nil,
	}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	signatureData, err := os.ReadFile("testdata/foo.zip.sig")
	require.NoError(t, err)

	err = validator.Validate("foo.zip", data, signatureData)
	assert.EqualError(t, err, ErrECDSAValidationFailed.Error())
}

func TestECDSAValidatorEmptySignature(t *testing.T) {
	validator := &ECDSAValidator{
		PublicKey: getTestPublicKey(t),
	}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	err = validator.Validate("foo.zip", data, nil)
	assert.EqualError(t, err, ErrInvalidECDSASignature.Error())
}

func TestECDSAValidator(t *testing.T) {
	validator := &ECDSAValidator{
		PublicKey: getTestPublicKey(t),
	}
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	signatureData, err := os.ReadFile("testdata/foo.zip.sig")
	require.NoError(t, err)

	err = validator.Validate("foo.zip", data, signatureData)
	assert.NoError(t, err)
}

func TestECDSAValidatorFail(t *testing.T) {
	validator := &ECDSAValidator{
		PublicKey: getTestPublicKey(t),
	}
	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	signatureData, err := os.ReadFile("testdata/foo.zip.sig")
	require.NoError(t, err)

	err = validator.Validate("foo.tar.xz", data, signatureData)
	assert.EqualError(t, err, ErrECDSAValidationFailed.Error())
}

func getTestPublicKey(t *testing.T) *ecdsa.PublicKey {
	pemData, err := os.ReadFile("testdata/Test.crt")
	require.NoError(t, err)

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatalf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)

	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Errorf("PublicKey is not ECDSA")
	}
	return pubKey
}

// ======= ChecksumValidator ====================================================

func TestChecksumValidatorEmptyFile(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, nil)
	assert.EqualError(t, err, ErrHashNotFound.Error())
}

func TestChecksumValidatorInvalidChecksumFile(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, []byte("blahblahblah"))
	assert.EqualError(t, err, ErrIncorrectChecksumFile.Error())
}

func TestChecksumValidatorWithUniqueLine(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, hashData)
	require.NoError(t, err)
}

func TestChecksumValidatorWillFailWithWrongHash(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, hashData)
	assert.EqualError(t, err, ErrChecksumValidationFailed.Error())
}

func TestChecksumNotFound(t *testing.T) {
	data, err := os.ReadFile("testdata/bar-not-found.zip")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("bar-not-found.zip", data, hashData)
	assert.EqualError(t, err, ErrHashNotFound.Error())
}

func TestChecksumValidatorSuccess(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	validator := &ChecksumValidator{"SHA256SUM"}
	err = validator.Validate("foo.tar.xz", data, hashData)
	assert.NoError(t, err)
}
