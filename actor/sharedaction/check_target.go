package sharedaction

import "code.cloudfoundry.org/cli/actor/actionerror"

// CheckTarget confirms that the user is logged in. Optionally it will also
// check if an organization and space are targeted.
func (actor Actor) CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if actor.Config.AccessToken() == "" && actor.Config.RefreshToken() == "" {
		return actionerror.NotLoggedInError{
			BinaryName: actor.Config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if !actor.Config.HasTargetedOrganization() {
			return actionerror.NoOrganizationTargetedError{
				BinaryName: actor.Config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if !actor.Config.HasTargetedSpace() {
				return actionerror.NoSpaceTargetedError{
					BinaryName: actor.Config.BinaryName(),
				}
			}
		}
	}

	return nil
}
