package configuration

import (
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
)

func NewRepository() coreconfig.Repository {
	return coreconfig.NewRepositoryFromPersistor(NewFakePersistor(), func(err error) {
		panic(err)
	})
}

func NewRepositoryWithAccessToken(tokenInfo coreconfig.TokenInfo) coreconfig.Repository {
	accessToken, err := EncodeAccessToken(tokenInfo)
	if err != nil {
		panic(err)
	}

	config := NewRepository()
	config.SetAccessToken(accessToken)
	return config
}

func NewRepositoryWithDefaults() coreconfig.Repository {
	configRepo := NewRepositoryWithAccessToken(coreconfig.TokenInfo{
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
