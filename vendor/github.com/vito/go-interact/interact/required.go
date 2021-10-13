package interact

// RequiredDestination wraps the real destination and indicates to Resolve
// that a value must be explicitly provided, and that there is no default. This
// is to distinguish from defaulting to the zero-value.
type RequiredDestination struct {
	Destination interface{}
}

// Required is a convenience function for constructing a RequiredDestination.
func Required(dst interface{}) RequiredDestination {
	return RequiredDestination{dst}
}
