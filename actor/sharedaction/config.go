package sharedaction

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Config

// Config a way of getting basic CF configuration
type Config interface {
	AccessToken() string
	BinaryName() string
	CurrentUserName() (string, error)
	HasTargetedOrganization() bool
	HasTargetedSpace() bool
	IsCFOnK8s() bool
	RefreshToken() string
	TargetedOrganizationName() string
	Verbose() (bool, []string)
}
