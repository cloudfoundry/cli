package requirements

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type LoginRequirement struct {
	config                 core_config.Reader
	apiEndpointRequirement ApiEndpointRequirement
}

func NewLoginRequirement(config core_config.Reader) LoginRequirement {
	return LoginRequirement{config, ApiEndpointRequirement{config}}
}

func (req LoginRequirement) Execute() error {

	if apiErr := req.apiEndpointRequirement.Execute(); apiErr != nil {
		return apiErr
	}

	if !req.config.IsLoggedIn() {
		return errors.New(terminal.NotLoggedInText())
	}

	return nil
}
