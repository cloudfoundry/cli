package requirements

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RoutingAPIRequirement struct {
	config core_config.Reader
}

func NewRoutingAPIRequirement(config core_config.Reader) RoutingAPIRequirement {
	return RoutingAPIRequirement{
		config,
	}
}

func (req RoutingAPIRequirement) Execute() error {
	if len(req.config.RoutingApiEndpoint()) == 0 {
		return errors.New(T("Routing API URI missing. Please log in again to set the URI automatically."))
	}

	return nil
}
