package selfupdate

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
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

// ======= PatternValidator ================================================

func TestPatternValidator(t *testing.T) {
	data, err := os.ReadFile("testdata/foo.zip")
	require.NoError(t, err)

	hashData, err := os.ReadFile("testdata/foo.zip.sha256")
	require.NoError(t, err)

	t.Run("Mapping", func(t *testing.T) {
		validator := new(PatternValidator).Add("foo.*", new(SHAValidator))
		{
			v, _ := validator.findValidator("foo.ext")
			assert.IsType(t, &SHAValidator{}, v)
		}

		assert.True(t, validator.MustContinueValidation("foo.zip"))
		assert.NoError(t, validator.Validate("foo.zip", data, hashData))
		assert.Equal(t, "foo.zip.sha256", validator.GetValidationAssetName("foo.zip"))

		assert.Error(t, validator.Validate("foo.zip", data, data))
		assert.Error(t, validator.Validate("unmapped", data, hashData))
	})

	t.Run("MappingInvalidPanics", func(t *testing.T) {
		assert.PanicsWithError(t, "failed adding \"\\\\\": syntax error in pattern", func() {
			new(PatternValidator).Add("\\", new(SHAValidator))
		})
	})

	t.Run("Skip", func(t *testing.T) {
		validator := new(PatternValidator).SkipValidation("*.skipped")

		assert.False(t, validator.MustContinueValidation("foo.skipped"))
		assert.NoError(t, validator.Validate("foo.skipped", nil, nil))
		assert.Equal(t, "foo.skipped", validator.GetValidationAssetName("foo.skipped"))
	})

	t.Run("Unmapped", func(t *testing.T) {
		validator := new(PatternValidator)

		assert.False(t, validator.MustContinueValidation("foo.zip"))
		assert.ErrorIs(t, ErrValidatorNotFound, validator.Validate("foo.zip", data, hashData))
		assert.Equal(t, "foo.zip", validator.GetValidationAssetName("foo.zip"))
	})

	t.Run("SupportsNesting", func(t *testing.T) {
		nested := new(PatternValidator).Add("**/*.zip", new(SHAValidator))
		validator := new(PatternValidator).Add("path/**", nested)
		{
			v, _ := validator.findValidator("path/foo")
			assert.Equal(t, nested, v)
		}

		assert.True(t, validator.MustContinueValidation("path/foo.zip"))
		assert.False(t, validator.MustContinueValidation("path/other"))
		assert.NoError(t, validator.Validate("path/foo.zip", data, hashData))
		assert.Error(t, validator.Validate("foo.zip", data, hashData))
	})
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
	assert.ErrorIs(t, err, ErrChecksumValidationFailed)
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

func TestECDSAValidatorWithKeyFromPem(t *testing.T) {
	pemData, err := os.ReadFile("testdata/Test.crt")
	require.NoError(t, err)

	validator := new(ECDSAValidator).WithPublicKey(pemData)
	assert.True(t, getTestPublicKey(t).Equal(validator.PublicKey))

	assert.PanicsWithError(t, "failed to decode PEM block", func() {
		new(ECDSAValidator).WithPublicKey([]byte{})
	})

	assert.PanicsWithError(t, "failed to parse certificate in PEM block: x509: malformed certificate", func() {
		new(ECDSAValidator).WithPublicKey([]byte(`
-----BEGIN CERTIFICATE-----

-----END CERTIFICATE-----
`))
	})
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
	assert.ErrorIs(t, err, ErrChecksumValidationFailed)
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

// ======= Utilities =========================================================

func TestHexStringEquals(t *testing.T) {
	tests := []struct {
		equal bool
		size  int
		a, b  string
		err   error
	}{
		{true, 0, "", "", nil},
		{true, 1, "b1", "b1", nil},
		{true, 1, "b1", "B1", nil},
		{true, 2, "b1AA", "B1aa", nil},
		{false, 1, "", "", nil},
		{false, 0, "b1", "b1", nil},
		{false, 0, "b", "b", nil},
		{false, 1, "b", "b", nil},
		{false, 1, "b2", "b1", nil},
		{false, 2, "b1", "b1", nil},
		{false, 2, "b1AA", "aab1", nil},
		{false, 3, "b1", "b1", nil},
		{false, 2, "aaXX", "aaXX", hex.InvalidByteError('X')},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			equal, err := hexStringEquals(test.size, test.a, test.b)
			assert.Equal(t, test.equal, equal)
			assert.ErrorIs(t, err, test.err)
		})
	}
}

func TestNewChecksumWithECDSAValidator(t *testing.T) {
	pemData, err := os.ReadFile("testdata/Test.crt")
	require.NoError(t, err)

	validator := NewChecksumWithECDSAValidator("checksums", pemData)
	assert.Implements(t, (*RecursiveValidator)(nil), validator)
	assert.Equal(t, "checksums", validator.GetValidationAssetName("anything"))
	assert.Equal(t, "checksums.sig", validator.GetValidationAssetName("checksums"))
}
