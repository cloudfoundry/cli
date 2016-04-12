package requirements

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type LoginRequirement struct {
	config                 coreconfig.Reader
	apiEndpointRequirement ApiEndpointRequirement
}

func NewLoginRequirement(config coreconfig.Reader) LoginRequirement {
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
