package configuration

import (
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
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
		UserGUID: "my-user-guid",
		Username: "my-user",
		Email:    "my-user-email",
	})

	configRepo.SetSpaceFields(models.SpaceFields{
		Name: "my-space",
		GUID: "my-space-guid",
	})
	configRepo.SetOrganizationFields(models.OrganizationFields{
		Name: "my-org",
		GUID: "my-org-guid",
	})

	return configRepo
}
