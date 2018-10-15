package sharedaction

// CheckOrgSpaceTargeted confirms that the user is targeting an org and a space.
func (actor Actor) CheckOrgSpaceTargeted() bool {
	return actor.Config.HasTargetedOrganization() && actor.Config.HasTargetedSpace()
}
