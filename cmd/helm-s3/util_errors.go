package main

import (
	"errors"

	"github.com/spf13/cobra"
)

type errorType int

const (
	errorTypeBadUsage errorType = iota + 1
	errorTypeSilent
)

func (t errorType) Is(err error) bool {
	if err == nil {
		return false
	}

	var target interface{ errorType() errorType }
	return errors.As(err, &target) && target.errorType() == t
}

type customError struct {
	errType errorType
	err     error
}

func (c customError) Error() string {
	return c.err.Error()
}

func (c customError) errorType() errorType {
	return c.errType
}

func newBadUsageError(err error) error {
	return customError{
		errType: errorTypeBadUsage,
		err:     err,
	}
}

func newSilentError() error {
	return customError{
		errType: errorTypeSilent,
		err:     errors.New(""),
	}
}

func wrapPositionalArgsBadUsage(f cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		err := f(cmd, args)
		if err != nil {
			return newBadUsageError(err)
		}
		return nil
	}
}
