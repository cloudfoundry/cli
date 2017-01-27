package v2action

//go:generate counterfeiter . Config

type Config interface {
	UnsetOrganizationInformation()
	UnsetSpaceInformation()
	SetTargetInformation(api string, apiVersion string, auth string, minCLIVersion string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
	SkipSSLValidation() bool
	Target() string
}
