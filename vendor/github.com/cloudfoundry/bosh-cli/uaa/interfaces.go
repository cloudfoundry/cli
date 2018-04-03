package uaa

//go:generate counterfeiter . UAA

type UAA interface {
	NewStaleAccessToken(refreshValue string) StaleAccessToken

	Prompts() ([]Prompt, error)

	ClientCredentialsGrant() (Token, error)
	OwnerPasswordCredentialsGrant([]PromptAnswer) (AccessToken, error)
}

//go:generate counterfeiter . Token

// Token is a plain token with a value.
type Token interface {
	Type() string
	Value() string
}

//go:generate counterfeiter . AccessToken

// AccessToken is a token that can be refreshed.
type AccessToken interface {
	Token
	RefreshToken() Token
	Refresh() (AccessToken, error)
}

// StaleAccessToken represents a token that should only be refreshed.
// Its value cannot be retrieved since it's stale hence should not be used.
type StaleAccessToken interface {
	RefreshToken() Token
	Refresh() (AccessToken, error)
}
