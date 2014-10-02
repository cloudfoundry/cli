package core_config

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

type TokenInfo struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	UserGuid string `json:"user_id"`
}

func NewTokenInfo(accessToken string) (info TokenInfo) {
	tokenJson, err := DecodeAccessToken(accessToken)

	if err != nil {
		return
	}
	info = TokenInfo{}
	err = json.Unmarshal(tokenJson, &info)
	return
}

func DecodeAccessToken(accessToken string) (tokenJson []byte, err error) {
	tokenParts := strings.Split(accessToken, " ")

	if len(tokenParts) < 2 {
		return
	}

	token := tokenParts[1]
	encodedParts := strings.Split(token, ".")

	if len(encodedParts) < 3 {
		return
	}

	encodedTokenJson := encodedParts[1]
	return base64Decode(encodedTokenJson)
}

func base64Decode(encodedData string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(restorePadding(encodedData))
}

func restorePadding(seg string) string {
	switch len(seg) % 4 {
	case 2:
		seg = seg + "=="
	case 3:
		seg = seg + "="
	}
	return seg
}
