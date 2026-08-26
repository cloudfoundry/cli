package v7action

import (
	"strings"
	"time"

	utiljwt "code.cloudfoundry.org/cli/v9/util/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func (actor Actor) RefreshAccessToken() (string, error) {
	var expiresIn time.Duration

	refreshToken := actor.Config.RefreshToken()

	accessTokenString := strings.TrimPrefix(actor.Config.AccessToken(), "bearer ")
	claims, err := utiljwt.Parse(accessTokenString)

	if err == nil {
		expiration, err := claims.GetExpirationTime()
		if err == nil && expiration != nil {
			expiresIn = time.Until(expiration.Time)
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

func (actor Actor) ParseAccessToken(accessToken string) (jwtv5.MapClaims, error) {
	tokenStr := strings.TrimPrefix(accessToken, "bearer ")
	return utiljwt.Parse(tokenStr)
}
