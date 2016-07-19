package configuration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

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
