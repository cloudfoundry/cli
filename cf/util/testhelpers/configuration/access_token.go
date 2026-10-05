package configuration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	"code.cloudfoundry.org/cli/v9/cf/configuration/coreconfig"
)

func EncodeAccessToken(tokenInfo coreconfig.TokenInfo) (accessToken string, err error) {
	tokenInfoBytes, err := json.Marshal(tokenInfo)
	if err != nil {
		return
	}
	encodedTokenInfo := base64.StdEncoding.EncodeToString(tokenInfoBytes)
	accessToken = fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo)
	return
}

// BuildTokenString builds a minimal JWT with the given time as expiration claim.
func BuildTokenString(expiration time.Time) string {
	claims := jwtv5.MapClaims{"exp": jwtv5.NewNumericDate(expiration)}
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodNone, claims)
	tokenBytes, _ := token.SignedString(jwtv5.UnsafeAllowNoneSignatureType)
	return string(tokenBytes)
}
