package command

//go:generate counterfeiter . SharedActor

type SharedActor interface {
	CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error
	RequireCurrentUser() (string, error)
	RequireTargetedOrg() (string, error)
	IsLoggedIn() bool
	CheckOrgSpaceTargeted() bool
}
