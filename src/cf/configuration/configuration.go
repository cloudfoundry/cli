package configuration

import (
	"cf"
	"encoding/json"
	"time"
)

type Configuration struct {
	Target                  string
	ApiVersion              string
	AuthorizationEndpoint   string
	AccessToken             string
	Organization            cf.Organization
	Space                   cf.Space
	ApplicationStartTimeout time.Duration // will be used as seconds
}

func (c Configuration) UserEmail() (email string) {
	clearInfo, err := DecodeTokenInfo(c.AccessToken)

	if err != nil {
		return
	}

	type TokenInfo struct {
		UserName string `json:"user_name"`
		Email    string `json:"email"`
	}

	tokenInfo := new(TokenInfo)
	err = json.Unmarshal(clearInfo, &tokenInfo)

	if err != nil {
		return
	}

	return tokenInfo.Email
}

func (c Configuration) IsLoggedIn() bool {
	return c.AccessToken != ""
}

func (c Configuration) HasOrganization() bool {
	return c.Organization.Guid != "" && c.Organization.Name != ""
}

func (c Configuration) HasSpace() bool {
	return c.Space.Guid != "" && c.Space.Name != ""
}
