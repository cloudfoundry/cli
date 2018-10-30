package command

//go:generate counterfeiter . SharedActor

type SharedActor interface {
	IsLoggedIn() bool
	IsSpaceTargeted() bool
	IsOrgTargeted() bool

	CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error
	RequireCurrentUser() (string, error)
	RequireTargetedOrg() (string, error)
}
