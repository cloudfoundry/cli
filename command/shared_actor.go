package command

//go:generate counterfeiter . SharedActor

type SharedActor interface {
	CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error
}
