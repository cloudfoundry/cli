package configuration

import (
	"cf"
	"encoding/json"
	"strings"
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

func (c Configuration) UAAEndpoint() string {
	return strings.Replace(c.AuthorizationEndpoint, "login", "uaa", 1)
}

func (c Configuration) UserEmail() (email string) {
	info, err := c.getTokenInfo()

	if err != nil {
		return
	}

	return info.Email
}

func (c Configuration) UserGuid() (guid string) {
	info, err := c.getTokenInfo()

	if err != nil {
		return
	}

	return info.UserGuid
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

type tokenInfo struct {
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	UserGuid string `json:"user_id"`
}

func (c Configuration) getTokenInfo() (info tokenInfo, err error) {
	clearInfo, err := DecodeTokenInfo(c.AccessToken)

	if err != nil {
		return
	}
	info = tokenInfo{}
	err = json.Unmarshal(clearInfo, &info)
	return
}
