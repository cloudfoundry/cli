package requirements

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type RoutingAPIRequirement struct {
	ui     terminal.UI
	config core_config.Reader
}

func NewRoutingAPIRequirement(ui terminal.UI, config core_config.Reader) RoutingAPIRequirement {
	return RoutingAPIRequirement{
		ui,
		config,
	}
}

func (req RoutingAPIRequirement) Execute() bool {
	if len(req.config.RoutingApiEndpoint()) == 0 {
		req.ui.Failed(T("Routing API uri missing. Please log in again and retry."))
		return false
	}

	return true
}
