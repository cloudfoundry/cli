package configuration

import (
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/cf/models"
	"code.cloudfoundry.org/cli/v8/util/configv3"
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

// NewConfigWithDefaults returns a configv3.Config initialized with the same
// default values as NewRepositoryWithDefaults for testing purposes.
func NewConfigWithDefaults() *configv3.Config {
	// Create a proper JWT token with user info for configv3
	// For UAA users, the user_name field contains the email
	accessToken := BuildTokenStringWithUserInfo("my-user-guid", "my-user-email", "my-user-email", "uaa")

	jsonConfig := configv3.JSONConfig{
		AccessToken: accessToken,
		TargetedOrganization: configv3.Organization{
			Name: "my-org",
			GUID: "my-org-guid",
		},
		TargetedSpace: configv3.Space{
			Name: "my-space",
			GUID: "my-space-guid",
		},
	}

	return &configv3.Config{
		ConfigFile: jsonConfig,
		UserConfig: configv3.DefaultUserConfig{
			ConfigFile: &jsonConfig,
		},
	}
}
