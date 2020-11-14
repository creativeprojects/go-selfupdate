package selfupdate

import "errors"

// Error
var (
	ErrIncorrectParameterOwner = errors.New("incorrect parameter \"owner\"")
	ErrIncorrectParameterRepo  = errors.New("incorrect parameter \"repo\"")
	ErrAssetNotFound           = errors.New("asset not found")
)
