package selfupdate

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSHAValidator(t *testing.T) {
	validator := &SHAValidator{}
	data, err := ioutil.ReadFile("testdata/foo.zip")
	if err != nil {
		t.Fatal(err)
	}
	hashData, err := ioutil.ReadFile("testdata/foo.zip.sha256")
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.Validate("foo.zip", data, hashData); err != nil {
		t.Fatal(err)
	}
}

func TestSHAValidatorFail(t *testing.T) {
	validator := &SHAValidator{}
	data, err := ioutil.ReadFile("testdata/foo.zip")
	if err != nil {
		t.Fatal(err)
	}
	hashData, err := ioutil.ReadFile("testdata/foo.zip.sha256")
	if err != nil {
		t.Fatal(err)
	}
	hashData[0] = '0'
	if err := validator.Validate("foo.zip", data, hashData); err == nil {
		t.Fatal(err)
	}
}

func TestECDSAValidator(t *testing.T) {
	pemData, err := ioutil.ReadFile("testdata/Test.crt")
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatalf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate")
	}

	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Errorf("PublicKey is not ECDSA")
	}

	validator := &ECDSAValidator{
		PublicKey: pubKey,
	}
	data, err := ioutil.ReadFile("testdata/foo.zip")
	if err != nil {
		t.Fatal(err)
	}
	signatureData, err := ioutil.ReadFile("testdata/foo.zip.sig")
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.Validate("foo.zip", data, signatureData); err != nil {
		t.Fatal(err)
	}
}

func TestECDSAValidatorFail(t *testing.T) {
	pemData, err := ioutil.ReadFile("testdata/Test.crt")
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatalf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("failed to parse certificate")
	}

	pubKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		t.Errorf("PublicKey is not ECDSA")
	}

	validator := &ECDSAValidator{
		PublicKey: pubKey,
	}
	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	if err != nil {
		t.Fatal(err)
	}
	signatureData, err := ioutil.ReadFile("testdata/foo.zip.sig")
	if err != nil {
		t.Fatal(err)
	}
	if err := validator.Validate("foo.tar.xz", data, signatureData); err == nil {
		t.Fatal(err)
	}
}

func TestValidatorSuffix(t *testing.T) {
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

func TestChecksumValidatorEmptyFile(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash for file \"foo.zip\" not found in checksum file")
}

func TestChecksumValidatorInvalidChecksumFile(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, []byte("blahblahblah"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect checksum file format")
}

func TestChecksumValidatorWithUniqueLine(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, hashData)
	require.NoError(t, err)
}

func TestChecksumValidatorWillFailWithWrongHash(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("foo.zip", data, hashData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sha256 validation failed")
}

func TestChecksumNotFound(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/bar-not-found.zip")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	validator := &ChecksumValidator{}
	err = validator.Validate("bar-not-found.zip", data, hashData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "hash for file \"bar-not-found.zip\" not found in checksum file")
}

func TestChecksumValidatorSuccess(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/foo.tar.xz")
	require.NoError(t, err)

	hashData, err := ioutil.ReadFile("testdata/SHA256SUM")
	require.NoError(t, err)

	validator := &ChecksumValidator{"SHA256SUM"}
	err = validator.Validate("foo.tar.xz", data, hashData)
	require.NoError(t, err)
}
