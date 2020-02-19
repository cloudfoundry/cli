package v2action

import (
	"strings"

	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

func (actor Actor) RefreshAccessToken(refreshToken string) (string, error) {

	tokens, err := actor.UAAClient.RefreshAccessToken(refreshToken)
	if err != nil {
		return "", err
	}

	actor.Config.SetAccessToken(tokens.AuthorizationToken())
	actor.Config.SetRefreshToken(tokens.RefreshToken)

	return tokens.AuthorizationToken(), nil
}

func (actor Actor) ParseAccessToken(accessToken string) (jwt.JWT, error) {
	tokenStr := strings.TrimPrefix(accessToken, "bearer ")
	return jws.ParseJWT([]byte(tokenStr))
}
