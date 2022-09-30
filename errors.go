package selfupdate

import "errors"

// Possible errors returned
var (
	ErrInvalidSlug                 = errors.New("invalid slug format, expected 'owner/name'")
	ErrIncorrectParameterOwner     = errors.New("incorrect parameter \"owner\"")
	ErrIncorrectParameterRepo      = errors.New("incorrect parameter \"repo\"")
	ErrInvalidID                   = errors.New("invalid repository ID, expected 'owner/name' but found number")
	ErrAssetNotFound               = errors.New("asset not found")
	ErrValidationAssetNotFound     = errors.New("validation file not found")
	ErrIncorrectChecksumFile       = errors.New("incorrect checksum file format")
	ErrChecksumValidationFailed    = errors.New("sha256 validation failed")
	ErrHashNotFound                = errors.New("hash not found in checksum file")
	ErrECDSAValidationFailed       = errors.New("ECDSA signature verification failed")
	ErrInvalidECDSASignature       = errors.New("invalid ECDSA signature")
	ErrCannotDecompressFile        = errors.New("failed to decompress")
	ErrExecutableNotFoundInArchive = errors.New("executable not found")
)
