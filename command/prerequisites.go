package command

// CheckTarget confirms that the user is logged in. Optionally it will also
// check if an organization and space are targeted.
func CheckTarget(config Config, targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if config.AccessToken() == "" && config.RefreshToken() == "" {
		return NotLoggedInError{
			BinaryName: config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if config.TargetedOrganization().GUID == "" {
			return NoTargetedOrgError{
				BinaryName: config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if config.TargetedSpace().GUID == "" {
				return NoTargetedSpaceError{
					BinaryName: config.BinaryName(),
				}
			}
		}
	}

	return nil
}
