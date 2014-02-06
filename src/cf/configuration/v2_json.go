package configuration

import (
	"cf/models"
	"encoding/json"
	"time"
)

type configJsonV2 struct {
	ConfigVersion           int
	Target                  string
	ApiVersion              string
	AuthorizationEndpoint   string
	LoggregatorEndpoint     string
	AccessToken             string
	RefreshToken            string
	OrganizationFields      models.OrganizationFields
	SpaceFields             models.SpaceFields
	ApplicationStartTimeout time.Duration // will be used as seconds
}

func ConfigToJsonV2(config ConfigReader) (output []byte, err error) {
	configJson := configJsonV2{
		ConfigVersion:           2,
		Target:                  config.ApiEndpoint(),
		ApiVersion:              config.ApiVersion(),
		AuthorizationEndpoint:   config.AuthorizationEndpoint(),
		LoggregatorEndpoint:     config.LoggregatorEndpoint(),
		AccessToken:             config.AccessToken(),
		RefreshToken:            config.RefreshToken(),
		OrganizationFields:      config.OrganizationFields(),
		SpaceFields:             config.SpaceFields(),
		ApplicationStartTimeout: config.ApplicationStartTimeout(),
	}

	return json.Marshal(configJson)
}

func ConfigFromJsonV2(input []byte) (config ConfigReadWriteCloser, err error) {
	configJson := new(configJsonV2)
	err = json.Unmarshal(input, configJson)
	if err != nil {
		return
	}

	config = NewConfigReadWriteCloser(newConfiguration())
	if configJson.ConfigVersion != 2 {
		return
	}

	config.SetApiEndpoint(configJson.Target)
	config.SetApiVersion(configJson.ApiVersion)
	config.SetAccessToken(configJson.AccessToken)
	config.SetApplicationStartTimeout(configJson.ApplicationStartTimeout)
	config.SetRefreshToken(configJson.RefreshToken)
	config.SetSpaceFields(configJson.SpaceFields)
	config.SetOrganizationFields(configJson.OrganizationFields)
	config.SetLoggregatorEndpoint(configJson.LoggregatorEndpoint)
	config.SetAuthorizationEndpoint(configJson.AuthorizationEndpoint)

	return
}
