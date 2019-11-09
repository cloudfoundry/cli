package v7action

func (actor Actor) RefreshAccessToken() (string, error) {

	refreshToken := actor.Config.RefreshToken()

	tokens, err := actor.UAAClient.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	return tokens.AccessToken, nil
}
