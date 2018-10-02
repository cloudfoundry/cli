package sharedaction

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	AccessToken() string
	BinaryName() string
	CurrentUserName() (string, error)
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	RefreshToken() string
	TargetedOrganizationName() string
	Verbose() (bool, []string)
}
