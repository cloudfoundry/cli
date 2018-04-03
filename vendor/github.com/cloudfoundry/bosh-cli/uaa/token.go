package uaa

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type TokenImpl struct {
	type_ string
	value string
}

func (t TokenImpl) Type() string  { return t.type_ }
func (t TokenImpl) Value() string { return t.value }

type AccessTokenImpl struct {
	client Client

	type_        string
	accessValue  string
	refreshValue string
}

func (t AccessTokenImpl) Type() string  { return t.type_ }
func (t AccessTokenImpl) Value() string { return t.accessValue }

func (t AccessTokenImpl) RefreshToken() Token {
	return TokenImpl{type_: t.type_, value: t.refreshValue}
}

func (t AccessTokenImpl) Refresh() (AccessToken, error) {
	resp, err := t.client.RefreshTokenGrant(t.refreshValue)
	if err != nil {
		return nil, err
	}

	token := AccessTokenImpl{
		client: t.client,

		type_:        resp.Type,
		accessValue:  resp.AccessToken,
		refreshValue: resp.RefreshToken,
	}

	return token, nil
}

type TokenInfo struct {
	Username  string   `json:"user_name"` // e.g. "admin",
	Scopes    []string `json:"scope"`     // e.g. ["openid","bosh.admin"]
	ExpiredAt int      `json:"exp"`
	// ...snip...
}

func NewTokenInfoFromValue(value string) (TokenInfo, error) {
	var info TokenInfo

	segments := strings.Split(value, ".")
	if len(segments) != 3 {
		return info, bosherr.Error("Expected token value to have 3 segments")
	}

	bytes, err := base64.RawURLEncoding.DecodeString(segments[1])
	if err != nil {
		return info, bosherr.WrapError(err, "Decoding token info")
	}

	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return info, bosherr.WrapError(err, "Unmarshaling token info")
	}

	return info, nil
}
