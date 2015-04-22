package configuration

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
)

func NewRepository() core_config.Repository {
	return core_config.NewRepositoryFromPersistor(NewFakePersistor(), func(err error) {
		panic(err)
	})
}

func NewRepositoryWithAccessToken(tokenInfo core_config.TokenInfo) core_config.Repository {
	accessToken, err := EncodeAccessToken(tokenInfo)
	if err != nil {
		panic(err)
	}

	config := NewRepository()
	config.SetAccessToken(accessToken)
	return config
}

func NewRepositoryWithDefaults() core_config.Repository {
	configRepo := NewRepositoryWithAccessToken(core_config.TokenInfo{
		UserGuid: "my-user-guid",
		Username: "my-user",
		Email:    "my-user-email",
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
