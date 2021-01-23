package selfupdate

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// Validator represents an interface which enables additional validation of releases.
type Validator interface {
	// Validate validates release bytes against an additional asset bytes.
	// See SHAValidator or ECDSAValidator for more information.
	Validate(filename string, release, asset []byte) error
	// GetValidationAssetName returns the additional asset name containing the validation checksum.
	// The asset containing the checksum can be based on the release asset name
	// Please note if the validation file cannot be found, the DetectLatest and DetectVersion methods
	// will fail with a wrapped ErrValidationAssetNotFound error
	GetValidationAssetName(releaseFilename string) string
}

//=====================================================================================================================

// SHAValidator specifies a SHA256 validator for additional file validation
// before updating.
type SHAValidator struct {
}

// Validate checks the SHA256 sum of the release against the contents of an
// additional asset file.
func (v *SHAValidator) Validate(filename string, release, asset []byte) error {
	// we'd better check the size of the file otherwise it's going to panic
	if len(asset) < sha256.BlockSize {
		return ErrIncorrectChecksumFile
	}
	hash := fmt.Sprintf("%s", asset[:sha256.BlockSize])
	calculatedHash := fmt.Sprintf("%x", sha256.Sum256(release))
	if calculatedHash != hash {
		return ErrChecksumValidationFailed
	}
	return nil
}

// GetValidationAssetName returns the asset name for SHA256 validation.
func (v *SHAValidator) GetValidationAssetName(releaseFilename string) string {
	return releaseFilename + ".sha256"
}

//=====================================================================================================================

// ChecksumValidator is a SHA256 checksum validator where all the validation hash are in a single file (one per line)
type ChecksumValidator struct {
	// UniqueFilename is the name of the global file containing all the checksums
	// Usually "checksums.txt", "SHA256SUMS", etc.
	UniqueFilename string
}

// Validate the SHA256 sum of the release against the contents of an
// additional asset file containing all the checksums (one file per line).
func (v *ChecksumValidator) Validate(filename string, release, asset []byte) error {
	hash, err := findChecksum(filename, asset)
	if err != nil {
		return err
	}
	calculatedHash := fmt.Sprintf("%x", sha256.Sum256(release))
	if calculatedHash != hash {
		return ErrChecksumValidationFailed
	}
	return nil
}

func findChecksum(filename string, content []byte) (string, error) {
	// check if the file has windows line ending (probably better than just testing the platform)
	crlf := []byte("\r\n")
	lf := []byte("\n")
	eol := lf
	if bytes.Contains(content, crlf) {
		log.Print("Checksum file is using windows line ending")
		eol = crlf
	}
	lines := bytes.Split(content, eol)
	log.Printf("Checksum validator: %d checksums available, searching for %q", len(lines), filename)
	for _, line := range lines {
		// skip empty line
		if len(line) == 0 {
			continue
		}
		parts := bytes.Split(line, []byte("  "))
		if len(parts) != 2 {
			return "", ErrIncorrectChecksumFile
		}
		if string(parts[1]) == filename {
			return string(parts[0]), nil
		}
	}
	return "", ErrHashNotFound
}

// GetValidationAssetName returns the unique asset name for SHA256 validation.
func (v *ChecksumValidator) GetValidationAssetName(releaseFilename string) string {
	return v.UniqueFilename
}

//=====================================================================================================================

// ECDSAValidator specifies a ECDSA validator for additional file validation
// before updating.
type ECDSAValidator struct {
	PublicKey *ecdsa.PublicKey
}

// Validate checks the ECDSA signature the release against the signature
// contained in an additional asset file.
func (v *ECDSAValidator) Validate(filename string, input, signature []byte) error {
	h := sha256.New()
	h.Write(input)

	var rs struct {
		R *big.Int
		S *big.Int
	}
	if _, err := asn1.Unmarshal(signature, &rs); err != nil {
		return ErrInvalidECDSASignature
	}

	if v.PublicKey == nil || !ecdsa.Verify(v.PublicKey, h.Sum([]byte{}), rs.R, rs.S) {
		return ErrECDSAValidationFailed
	}

	return nil
}

// GetValidationAssetName returns the asset name for ECDSA validation.
func (v *ECDSAValidator) GetValidationAssetName(releaseFilename string) string {
	return releaseFilename + ".sig"
}

//=====================================================================================================================

// Verify interface
var (
	_ Validator = &SHAValidator{}
	_ Validator = &ChecksumValidator{}
	_ Validator = &ECDSAValidator{}
)
