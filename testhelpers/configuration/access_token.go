package configuration

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
)

func EncodeAccessToken(tokenInfo core_config.TokenInfo) (accessToken string, err error) {
	tokenInfoBytes, err := json.Marshal(tokenInfo)
	if err != nil {
		return
	}
	encodedTokenInfo := base64.StdEncoding.EncodeToString(tokenInfoBytes)
	accessToken = fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo)
	return
}
