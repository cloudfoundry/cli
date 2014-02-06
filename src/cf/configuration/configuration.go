package configuration

import (
	"cf/models"
	"time"
)

type config struct {
	config *Configuration
}

func NewConfigReadWriteCloser(c *Configuration) ConfigReadWriteCloser {
	result := new(config)
	result.config = c
	result.SetApplicationStartTimeout(30) // TODO - make sure this is seconds
	return result
}

type ConfigReader interface {
	ApiEndpoint() string
	ApiVersion() string
	AuthorizationEndpoint() string
	LoggregatorEndpoint() string
	AccessToken() string
	RefreshToken() string
	OrganizationFields() models.OrganizationFields
	SpaceFields() models.SpaceFields
	ApplicationStartTimeout() time.Duration

	HasSpace() bool
	HasOrganization() bool
	IsLoggedIn() bool
	Username() string
	UserGuid() string
}

type ConfigReadWriter interface {
	ConfigReader
	SetApiEndpoint(string)
	SetApiVersion(string)
	SetAuthorizationEndpoint(string)
	SetLoggregatorEndpoint(string)
	SetAccessToken(string)
	SetRefreshToken(string)
	SetOrganizationFields(models.OrganizationFields)
	SetSpaceFields(models.SpaceFields)
	SetApplicationStartTimeout(time.Duration)
}

type ConfigReadWriteCloser interface {
	ConfigReadWriter
	GetOldConfig() *Configuration
	Close()
}

func (c *config) ApiVersion() string {
	return c.config.ApiVersion
}

func (c *config) AuthorizationEndpoint() string {
	return c.config.AuthorizationEndpoint
}

func (c *config) LoggregatorEndpoint() string {
	return c.config.LoggregatorEndPoint
}

func (c *config) ApiEndpoint() string {
	return c.config.Target
}

func (c *config) AccessToken() string {
	return c.config.AccessToken
}

func (c *config) RefreshToken() string {
	return c.config.RefreshToken
}

func (c *config) OrganizationFields() models.OrganizationFields {
	return c.config.OrganizationFields
}

func (c *config) SpaceFields() models.SpaceFields {
	return c.config.SpaceFields
}

func (c *config) ApplicationStartTimeout() time.Duration {
	return c.config.ApplicationStartTimeout
}

func (c *config) UserEmail() (email string) {
	return c.config.getTokenInfo().Email
}

func (c *config) UserGuid() (guid string) {
	return c.config.getTokenInfo().UserGuid
}

func (c *config) Username() (guid string) {
	return c.config.getTokenInfo().Username
}

func (c *config) IsLoggedIn() bool {
	return c.config.AccessToken != ""
}

func (c *config) HasOrganization() bool {
	return c.config.OrganizationFields.Guid != "" && c.config.OrganizationFields.Name != ""
}

func (c *config) HasSpace() bool {
	return c.config.SpaceFields.Guid != "" && c.config.SpaceFields.Name != ""
}

// ConfigReadWriter

func (c *config) SetApiEndpoint(endpoint string) {
	c.config.Target = endpoint
}

func (c *config) SetApiVersion(version string) {
	c.config.ApiVersion = version
}

func (c *config) SetAuthorizationEndpoint(endpoint string) {
	c.config.AuthorizationEndpoint = endpoint
}

func (c *config) SetLoggregatorEndpoint(endpoint string) {
	c.config.LoggregatorEndPoint = endpoint
}

func (c *config) SetAccessToken(token string) {
	c.config.AccessToken = token
}

func (c *config) SetRefreshToken(token string) {
	c.config.RefreshToken = token
}

func (c *config) SetOrganizationFields(org models.OrganizationFields) {
	c.config.OrganizationFields = org
}

func (c *config) SetSpaceFields(space models.SpaceFields) {
	c.config.SpaceFields = space
}

func (c *config) SetApplicationStartTimeout(timeout time.Duration) {
	c.config.ApplicationStartTimeout = timeout
}

// ConfigReadWriteCloser

func (c *config) Close() {
}

func (c *config) GetOldConfig() *Configuration {
	return c.config
}
