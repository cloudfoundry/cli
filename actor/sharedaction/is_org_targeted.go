package sharedaction

// IsOrgTargeted determines whether an org is being targeted by the CLI
func (actor Actor) IsOrgTargeted() bool {
	return actor.Config.HasTargetedOrganization()
}
