package types

import "errors"

var (
	ErrInvalidName            = errors.New("must be lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character")
	ErrInvalidURL             = errors.New("must be a valid URL")
	ErrInvalidPathNotFound    = errors.New("must be a valid path")
	ErrInvalidPathIsDir       = errors.New("must be a file")
	ErrInvalidPathIsFile      = errors.New("must be a directory")
	ErrValidtorStopValidation = errors.New("stop validation")
)
