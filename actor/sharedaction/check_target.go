package sharedaction

// NotLoggedInError represents the scenario when the user is not logged in.
type NotLoggedInError struct {
	BinaryName string
}

func (NotLoggedInError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// NoOrganizationTargetedError represents the scenario when an org is not targeted.
type NoOrganizationTargetedError struct {
	BinaryName string
}

func (NoOrganizationTargetedError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// NoSpaceTargetedError represents the scenario when a space is not targeted.
type NoSpaceTargetedError struct {
	BinaryName string
}

func (NoSpaceTargetedError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// CheckTarget confirms that the user is logged in. Optionally it will also
// check if an organization and space are targeted.
func (Actor) CheckTarget(config Config, targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if config.AccessToken() == "" && config.RefreshToken() == "" {
		return NotLoggedInError{
			BinaryName: config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if !config.HasTargetedOrganization() {
			return NoOrganizationTargetedError{
				BinaryName: config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if !config.HasTargetedSpace() {
				return NoSpaceTargetedError{
					BinaryName: config.BinaryName(),
				}
			}
		}
	}

	return nil
}
