package configuration

import (
	"cf/configuration"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func CreateAccessTokenWithTokenInfo(tokenInfo configuration.TokenInfo) (accessToken string, err error) {
	tokenInfoBytes, err := json.Marshal(tokenInfo)
	if err != nil {
		return
	}
	encodedTokenInfo := base64.StdEncoding.EncodeToString(tokenInfoBytes)
	accessToken = fmt.Sprintf("BEARER my_access_token.%s.baz", encodedTokenInfo)
	return
}
