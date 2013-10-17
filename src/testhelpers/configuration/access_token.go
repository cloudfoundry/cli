package configuration

import (
	"encoding/base64"
	"fmt"
	"cf/configuration"
	"encoding/json"
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
