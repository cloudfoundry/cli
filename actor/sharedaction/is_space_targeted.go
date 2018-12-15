package sharedaction

// IsSpaceTargeted determines whether a space is being targeted by the CLI
func (actor Actor) IsSpaceTargeted() bool {
	return actor.Config.HasTargetedSpace()
}
