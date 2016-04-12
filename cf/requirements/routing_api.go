package requirements

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/cf/i18n"
)

type RoutingAPIRequirement struct {
	config coreconfig.Reader
}

func NewRoutingAPIRequirement(config coreconfig.Reader) RoutingAPIRequirement {
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
