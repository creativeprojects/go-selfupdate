// Package pgp provides a PGP signature validator for go-types.
//
// It is kept in a separate package so that applications not using PGP
// signatures do not pull in the OpenPGP library.
package pgp

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/creativeprojects/go-selfupdate/types"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// Possible errors returned
var (
	ErrInvalidSignature = errors.New("invalid PGP signature")
	ErrKeyRingNotSet    = errors.New("PGP key ring not set")
)

// Validator specifies a PGP validator for additional file validation
// before updating.
type Validator struct {
	// KeyRing is usually filled by openpgp.ReadArmoredKeyRing(bytes.NewReader(key)) with key being the PGP pub key.
	KeyRing openpgp.EntityList
	// Binary toggles whether to validate detached *.sig (binary) or *.asc (ascii) signature files
	Binary bool
}

// WithArmoredKeyRing is a convenience method to set KeyRing
func (g *Validator) WithArmoredKeyRing(key []byte) *Validator {
	if ring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(key)); err == nil {
		g.KeyRing = ring
	} else {
		panic(fmt.Errorf("failed setting armored public key ring: %w", err))
	}
	return g
}

// Validate checks the PGP signature of the release against the signature
// contained in an additional asset file.
func (g *Validator) Validate(filename string, release, signature []byte) (err error) {
	if g.KeyRing == nil {
		return ErrKeyRingNotSet
	}

	data, sig := bytes.NewReader(release), bytes.NewReader(signature)
	// A nil *packet.Config selects the library defaults.
	if g.Binary {
		_, err = openpgp.CheckDetachedSignature(g.KeyRing, data, sig, nil)
	} else {
		_, err = openpgp.CheckArmoredDetachedSignature(g.KeyRing, data, sig, nil)
	}

	if errors.Is(err, io.EOF) {
		err = ErrInvalidSignature
	}

	return err
}

// GetValidationAssetName returns the asset name for PGP validation.
func (g *Validator) GetValidationAssetName(releaseFilename string) string {
	if g.Binary {
		return releaseFilename + ".sig"
	}
	return releaseFilename + ".asc"
}

// NewChecksumValidator returns a validator that checks assets with a checksums file
// (e.g. SHA256SUMS) and the checksums file with an armored PGP signature (e.g. SHA256SUMS.asc).
func NewChecksumValidator(checksumsFilename string, armoredKeyRing []byte) types.Validator {
	return new(types.PatternValidator).
		Add(checksumsFilename, new(Validator).WithArmoredKeyRing(armoredKeyRing)).
		Add("*", &types.ChecksumValidator{UniqueFilename: checksumsFilename}).
		SkipValidation("*.asc")
}

// Verify interface
var _ types.Validator = &Validator{}
