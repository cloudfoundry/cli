package command

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SharedActor

type SharedActor interface {
	IsLoggedIn() bool
	IsSpaceTargeted() bool
	IsOrgTargeted() bool

	CheckTarget(targetedOrganizationRequired bool, targetedSpaceRequired bool) error
	RequireCurrentUser() (string, error)
	RequireTargetedOrg() (string, error)
}
