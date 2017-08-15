package v2action

func (actor Actor) RefreshAccessToken(refreshToken string) (string, error) {
	tokens, err := actor.UAAClient.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	actor.Config.SetAccessToken(tokens.AuthorizationToken())
	actor.Config.SetRefreshToken(tokens.RefreshToken)

	return tokens.AuthorizationToken(), nil
}
