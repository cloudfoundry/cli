package ccv2

//go:generate counterfeiter . AuthenticationStore

type AuthenticationStore interface {
	AccessToken() string
	RefreshToken() string
	// SetAccessToken(token string)
	// SetRefreshToken(token string)

	// ClientName() string
	// ClientSecret() string
	// SkipSSLValidation() bool
}
