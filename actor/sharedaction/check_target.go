package sharedaction

import "code.cloudfoundry.org/cli/actor/actionerror"

// CheckTarget confirms that the user is logged in. Optionally it will also
// check if an organization and space are targeted.
func (actor Actor) CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error {
	if !actor.IsLoggedIn() {
		return actionerror.NotLoggedInError{
			BinaryName: actor.Config.BinaryName(),
		}
	}

	if targetedOrganizationRequired {
		if !actor.IsOrgTargeted() {
			return actionerror.NoOrganizationTargetedError{
				BinaryName: actor.Config.BinaryName(),
			}
		}

		if targetedSpaceRequired {
			if !actor.IsSpaceTargeted() {
				return actionerror.NoSpaceTargetedError{
					BinaryName: actor.Config.BinaryName(),
				}
			}
		}
	}

	return nil
}

func (actor Actor) RequireCurrentUser() (string, error) {
	if !actor.IsLoggedIn() {
		return "", actionerror.NotLoggedInError{
			BinaryName: actor.Config.BinaryName(),
		}
	}

	return actor.Config.CurrentUserName()
}

func (actor Actor) RequireTargetedOrg() (string, error) {
	if !actor.IsOrgTargeted() {
		return "", actionerror.NoOrganizationTargetedError{
			BinaryName: actor.Config.BinaryName(),
		}
	}

	return actor.Config.TargetedOrganizationName(), nil
}
