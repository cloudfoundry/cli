package types

import "strconv"

// NullInt is a wrapper around integer values that can be null or an integer.
// Use IsSet to check if the value is provided, instead of checking against 0.
type NullInt struct {
	Value int
	IsSet bool
}

// ParseFlagValue is used to parse a user provided flag argument.
func (n *NullInt) ParseFlagValue(val string) error {
	if val == "" {
		return nil
	}

	intVal, err := strconv.Atoi(val)
	if err != nil {
		return err
	}

	n.Value = intVal
	n.IsSet = true

	return nil
}
