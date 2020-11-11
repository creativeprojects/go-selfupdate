package selfupdate

import (
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
	Validate(release, asset []byte) error
	// Suffix describes the additional file ending which is used for finding the
	// additional asset.
	Suffix() string
}

// SHAValidator specifies a SHA256 validator for additional file validation
// before updating.
type SHAValidator struct {
}

// Validate checks the SHA256 sum of the release against the contents of an
// additional asset file.
func (v *SHAValidator) Validate(release, asset []byte) error {
	calculatedHash := fmt.Sprintf("%x", sha256.Sum256(release))
	hash := fmt.Sprintf("%s", asset[:sha256.BlockSize])
	if calculatedHash != hash {
		return fmt.Errorf("sha256 validation failed: expected=%q, got=%q", calculatedHash, hash)
	}
	return nil
}

// Suffix returns the suffix for SHA256 validation.
func (v *SHAValidator) Suffix() string {
	return ".sha256"
}

// ECDSAValidator specifies a ECDSA validator for additional file validation
// before updating.
type ECDSAValidator struct {
	PublicKey *ecdsa.PublicKey
}

// Validate checks the ECDSA signature the release against the signature
// contained in an additional asset file.
// additional asset file.
func (v *ECDSAValidator) Validate(input, signature []byte) error {
	h := sha256.New()
	h.Write(input)

	var rs struct {
		R *big.Int
		S *big.Int
	}
	if _, err := asn1.Unmarshal(signature, &rs); err != nil {
		return fmt.Errorf("failed to unmarshal ecdsa signature: %v", err)
	}

	if !ecdsa.Verify(v.PublicKey, h.Sum([]byte{}), rs.R, rs.S) {
		return fmt.Errorf("ecdsa: signature verification failed")
	}

	return nil
}

// Suffix returns the suffix for ECDSA validation.
func (v *ECDSAValidator) Suffix() string {
	return ".sig"
}

// Verify interface
var (
	_ Validator = &SHAValidator{}
	_ Validator = &ECDSAValidator{}
)
