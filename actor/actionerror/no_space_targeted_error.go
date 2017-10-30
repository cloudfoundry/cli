package actionerror

// NoSpaceTargetedError represents the scenario when a space is not targeted.
type NoSpaceTargetedError struct {
	BinaryName string
}

func (NoSpaceTargetedError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}
