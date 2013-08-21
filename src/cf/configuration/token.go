package configuration

import (
	"encoding/base64"
	"strings"
)

func DecodeTokenInfo(accessToken string) (clearTokenInfo []byte, err error) {
	tokenParts := strings.Split(accessToken, " ")

	if len(tokenParts) < 2 {
		return
	}

	token := tokenParts[1]
	encodedInfoParts := strings.Split(token, ".")

	if len(encodedInfoParts) < 3 {
		return
	}

	encodedInfo := encodedInfoParts[1]
	return base64Decode(encodedInfo)
}

func base64Decode(encodedInfo string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(restorePadding(encodedInfo))
}

func restorePadding(seg string) string {
	switch len(seg) % 4 {
	case 2:
		seg = seg + "=="
	case 3:
		seg = seg + "==="
	}
	return seg
}
