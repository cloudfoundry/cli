package actionerror

// NoOrganizationTargetedError represents the scenario when an org is not targeted.
type NoOrganizationTargetedError struct {
	BinaryName string
}

func (NoOrganizationTargetedError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}
