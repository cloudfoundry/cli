package configuration

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
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

	configRepo.SetSpaceFields(models.SpaceFields{
		Name: "my-space",
		Guid: "my-space-guid",
	})
	configRepo.SetOrganizationFields(models.OrganizationFields{
		Name: "my-org",
		Guid: "my-org-guid",
	})

	return configRepo
}
