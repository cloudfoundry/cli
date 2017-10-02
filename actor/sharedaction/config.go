package sharedaction

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	AccessToken() string
	BinaryName() string
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	RefreshToken() string
	Verbose() (bool, []string)
}
