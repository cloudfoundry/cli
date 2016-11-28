package configaction

//go:generate counterfeiter . Config

// Config a way of getting basic CF configuration
type Config interface {
	Target() string
	SkipSSLValidation() bool
	SetTargetInformation(api string, apiVersion string, auth string, loggregator string, doppler string, uaa string, routing string, skipSSLValidation bool)
	SetTokenInformation(accessToken string, refreshToken string, sshOAuthClient string)
}
