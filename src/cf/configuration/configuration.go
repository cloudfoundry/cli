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
	RefreshToken            string
	Organization            cf.Organization
	Space                   cf.Space
	ApplicationStartTimeout time.Duration // will be used as seconds
}

func (c Configuration) UserEmail() (email string) {
	return c.getTokenInfo().Email
}

func (c Configuration) UserGuid() (guid string) {
	return c.getTokenInfo().UserGuid
}

func (c Configuration) Username() (guid string) {
	return c.getTokenInfo().Username
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

type TokenInfo struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	UserGuid string `json:"user_id"`
}

func (c Configuration) getTokenInfo() (info TokenInfo) {
	clearInfo, err := DecodeTokenInfo(c.AccessToken)

	if err != nil {
		return
	}
	info = TokenInfo{}
	err = json.Unmarshal(clearInfo, &info)
	return
}
