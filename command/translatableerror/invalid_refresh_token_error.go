package translatableerror

type InvalidRefreshTokenError struct {
}

func (InvalidRefreshTokenError) Error() string {
	return "The token expired, was revoked, or the token ID is incorrect. Please log back in to re-authenticate."
}

func (e InvalidRefreshTokenError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
