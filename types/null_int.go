package types

import "strconv"

type NullInt struct {
	Value int
	IsSet bool
}

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
