package v7action

import (
	"strings"
	"time"

	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
)

func (actor Actor) RefreshAccessToken() (string, error) {
	var expiresIn time.Duration

	refreshToken := actor.Config.RefreshToken()

	accessTokenString := strings.TrimPrefix(actor.Config.AccessToken(), "bearer ")
	token, err := jws.ParseJWT([]byte(accessTokenString))

	if err == nil {
		expiration, ok := token.Claims().Expiration()
		if ok {
			expiresIn = time.Until(expiration)
		}
	}

	if err != nil || expiresIn < time.Minute {
		tokens, err := actor.UAAClient.RefreshAccessToken(refreshToken)
		if err != nil {
			return "", err
		}

		actor.Config.SetAccessToken(tokens.AuthorizationToken())
		actor.Config.SetRefreshToken(tokens.RefreshToken)

		return tokens.AuthorizationToken(), nil
	}
	return actor.Config.AccessToken(), nil
}

func (actor Actor) ParseAccessToken(accessToken string) (jwt.JWT, error) {
	tokenStr := strings.TrimPrefix(accessToken, "bearer ")
	return jws.ParseJWT([]byte(tokenStr))
}
