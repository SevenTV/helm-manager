package errors

import "errors"

var (
	ErrFileNotSupported      = errors.New("file not supported for this command")
	ErrNamespaceNotSupported = errors.New("namespace not supported for this command")
	ErrSubcommandRequired    = errors.New("you must provide a subcommand")
	ErrUnexpectedArgs        = errors.New("unexpected arguments")
	ErrMissingRequiredArg    = func(arg string) error {
		return errors.New("missing required argument: " + arg)
	}
)
