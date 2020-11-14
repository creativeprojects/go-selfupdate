package selfupdate

import "errors"

// Error
var (
	ErrIncorrectParameterOwner  = errors.New("incorrect parameter \"owner\"")
	ErrIncorrectParameterRepo   = errors.New("incorrect parameter \"repo\"")
	ErrAssetNotFound            = errors.New("asset not found")
	ErrIncorrectChecksumFile    = errors.New("incorrect checksum file format")
	ErrChecksumValidationFailed = errors.New("sha256 validation failed")
	ErrHashNotFound             = errors.New("hash not found in checksum file")
	ErrECDSAValidationFailed    = errors.New("ECDSA signature verification failed")
	ErrInvalidECDSASignature    = errors.New("invalid ECDSA signature")
)
