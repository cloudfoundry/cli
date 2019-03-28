package configuration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
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
	c := jws.Claims{}
	c.SetExpiration(expiration)
	token := jws.NewJWT(c, crypto.Unsecured)
	tokenBytes, _ := token.Serialize(nil)
	return string(tokenBytes)
}
