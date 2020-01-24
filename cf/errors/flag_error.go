package errors

import "fmt"

type FlagError struct {
	FlagName string
	InvalidValue string
}

func NewFlagError(flagName, invalidValue string) error {
	return &FlagError{
		FlagName: flagName,
		InvalidValue: invalidValue,
	}
}

func (err *FlagError) Error() string {
	return fmt.Sprintf("Invalid value for flag %s: %s", err.FlagName, err.InvalidValue)
}
