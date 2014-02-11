package configuration

import (
	"cf/configuration"
	"cf/models"
)

func NewRepository() configuration.Repository {
	return configuration.NewRepositoryFromPersistor(NewFakePersistor(), func(err error) {
		panic(err)
	})
}

func NewRepositoryWithAccessToken(tokenInfo configuration.TokenInfo) configuration.Repository {
	accessToken, err := EncodeAccessToken(tokenInfo)
	if err != nil {
		panic(err)
	}

	config := NewRepository()
	config.SetAccessToken(accessToken)
	return config
}

func NewRepositoryWithDefaults() configuration.Repository {
	configRepo := NewRepositoryWithAccessToken(configuration.TokenInfo{
		UserGuid: "my-user-guid",
		Username: "my-user",
	})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"
	spaceFields.Guid = "my-space-guid"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"
	orgFields.Guid = "my-org-guid"

	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	return configRepo
}
