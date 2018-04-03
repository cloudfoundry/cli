package uaa

type AccessTokenSession struct {
	initToken StaleAccessToken
	lastToken AccessToken
}

func NewAccessTokenSession(accessToken StaleAccessToken) *AccessTokenSession {
	return &AccessTokenSession{initToken: accessToken}
}

// TokenFunc retrieves new access token on first time use
// instead of using existing access token optimizing for token
// being valid for a longer period of time. Subsequent calls
// will reuse access token until it's time for it to be refreshed.
func (s *AccessTokenSession) TokenFunc(retried bool) (string, error) {
	if s.lastToken == nil || retried {
		token, err := s.initToken.Refresh()
		if err != nil {
			return "", err
		}

		s.lastToken = token
		s.initToken = token
	}

	return s.lastToken.Type() + " " + s.lastToken.Value(), nil
}
