package configuration

import (
	"cf/models"
	"encoding/json"
	"time"
)

type Configuration struct {
	ConfigVersion           int
	Target                  string
	ApiVersion              string
	AuthorizationEndpoint   string
	LoggregatorEndPoint     string
	AccessToken             string
	RefreshToken            string
	OrganizationFields      models.OrganizationFields
	SpaceFields             models.SpaceFields
	ApplicationStartTimeout time.Duration // will be used as seconds
}

func newConfiguration() (config *Configuration) {
	config = new(Configuration)
	config.ApplicationStartTimeout = 30
	return
}

func (c *Configuration) UserEmail() (email string) {
	return c.getTokenInfo().Email
}

func (c *Configuration) UserGuid() (guid string) {
	return c.getTokenInfo().UserGuid
}

func (c *Configuration) Username() (guid string) {
	return c.getTokenInfo().Username
}

func (c *Configuration) IsLoggedIn() bool {
	return c.AccessToken != ""
}

func (c *Configuration) ClearTokens() {
	c.AccessToken = ""
	c.RefreshToken = ""
}

func (c *Configuration) SetOrganizationFields(org models.OrganizationFields) {
	c.OrganizationFields = org
}

func (c *Configuration) SetSpaceFields(space models.SpaceFields) {
	c.SpaceFields = space
}

func (c *Configuration) ClearSession() {
	c.ClearTokens()
	c.OrganizationFields = models.OrganizationFields{}
	c.SpaceFields = models.SpaceFields{}
}

func (c *Configuration) HasOrganization() bool {
	return c.OrganizationFields.Guid != "" && c.OrganizationFields.Name != ""
}

func (c *Configuration) HasSpace() bool {
	return c.SpaceFields.Guid != "" && c.SpaceFields.Name != ""
}

type TokenInfo struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	UserGuid string `json:"user_id"`
}

func (c *Configuration) getTokenInfo() (info TokenInfo) {
	clearInfo, err := DecodeTokenInfo(c.AccessToken)

	if err != nil {
		return
	}
	info = TokenInfo{}
	err = json.Unmarshal(clearInfo, &info)
	return
}
