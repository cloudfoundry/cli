package constant

// GrantType is the type of authentication being used to obtain the token.
type GrantType string

const (
	// GrantTypeClientCredentials is used for a preconfigured client ID/secret
	// authentication.
	GrantTypeClientCredentials GrantType = "client_credentials"
	// GrantTypePassword is used for user's username/password authentication.
	GrantTypePassword     GrantType = "password"
	GrantTypeRefreshToken GrantType = "refresh_token"
)
