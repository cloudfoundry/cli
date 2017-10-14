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
func (actor Actor) CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if actor.Config.AccessToken() == "" && actor.Config.RefreshToken() == "" {
		return NotLoggedInError{
			BinaryName: actor.Config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if !actor.Config.HasTargetedOrganization() {
			return NoOrganizationTargetedError{
				BinaryName: actor.Config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if !actor.Config.HasTargetedSpace() {
				return NoSpaceTargetedError{
					BinaryName: actor.Config.BinaryName(),
				}
			}
		}
	}

	return nil
}
