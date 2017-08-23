package types

// FilteredString is a wrapper around string values that can be null/default or an
// actual value.  Use IsSet to check if the value is provided, instead of
// checking against the empty string.
type FilteredString struct {
	IsSet bool
	Value string
}

// ParseValue is used to parse a user provided flag argument.
func (n *FilteredString) ParseValue(val string) {
	if val == "" {
		return
	}

	n.IsSet = true

	switch val {
	case "null", "default":
	default:
		n.Value = val
	}
}
