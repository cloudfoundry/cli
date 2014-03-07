package configuration

import (
	"cf/models"
	"encoding/json"
)

type configJsonV2 struct {
	ConfigVersion         int
	Target                string
	ApiVersion            string
	AuthorizationEndpoint string
	LoggregatorEndpoint   string
	AccessToken           string
	RefreshToken          string
	OrganizationFields    models.OrganizationFields
	SpaceFields           models.SpaceFields
	SSLDisabled           bool
}

func JsonMarshalV2(config *Data) (output []byte, err error) {
	return json.Marshal(configJsonV2{
		ConfigVersion:         2,
		Target:                config.Target,
		ApiVersion:            config.ApiVersion,
		AuthorizationEndpoint: config.AuthorizationEndpoint,
		LoggregatorEndpoint:   config.LoggregatorEndPoint,
		AccessToken:           config.AccessToken,
		RefreshToken:          config.RefreshToken,
		OrganizationFields:    config.OrganizationFields,
		SpaceFields:           config.SpaceFields,
		SSLDisabled:           config.SSLDisabled,
	})
}

func JsonUnmarshalV2(input []byte, config *Data) (err error) {
	configJson := new(configJsonV2)

	err = json.Unmarshal(input, configJson)
	if err != nil {
		return
	}

	if configJson.ConfigVersion != 2 {
		return
	}

	config.Target = configJson.Target
	config.ApiVersion = configJson.ApiVersion
	config.AccessToken = configJson.AccessToken
	config.RefreshToken = configJson.RefreshToken
	config.SpaceFields = configJson.SpaceFields
	config.OrganizationFields = configJson.OrganizationFields
	config.LoggregatorEndPoint = configJson.LoggregatorEndpoint
	config.AuthorizationEndpoint = configJson.AuthorizationEndpoint
	config.SSLDisabled = configJson.SSLDisabled

	return
}
