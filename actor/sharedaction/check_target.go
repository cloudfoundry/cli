package sharedaction

// NotLoggedInError represents the scenario when the user is not logged in.
type NotLoggedInError struct {
	BinaryName string
}

func (_ NotLoggedInError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// NoTargetedOrganizationError represents the scenario when an org is not targeted.
type NoTargetedOrganizationError struct {
	BinaryName string
}

func (_ NoTargetedOrganizationError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// NoTargetedSpaceError represents the scenario when a space is not targeted.
type NoTargetedSpaceError struct {
	BinaryName string
}

func (_ NoTargetedSpaceError) Error() string {
	// The error message will be replaced by a translated message, returning the
	// empty string does not add to the translation files.
	return ""
}

// CheckTarget confirms that the user is logged in. Optionally it will also
// check if an organization and space are targeted.
func (_ Actor) CheckTarget(config Config, targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if config.AccessToken() == "" && config.RefreshToken() == "" {
		return NotLoggedInError{
			BinaryName: config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if !config.HasTargetedOrganization() {
			return NoTargetedOrganizationError{
				BinaryName: config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if !config.HasTargetedSpace() {
				return NoTargetedSpaceError{
					BinaryName: config.BinaryName(),
				}
			}
		}
	}

	return nil
}
