package configuration

import (
	"cf/models"
)

type Data struct {
	ConfigVersion         int
	Target                string
	ApiVersion            string
	AuthorizationEndpoint string
	LoggregatorEndPoint   string
	AccessToken           string
	RefreshToken          string
	OrganizationFields    models.OrganizationFields
	SpaceFields           models.SpaceFields
}

func NewData() (data *Data) {
	data = new(Data)
	return
}
