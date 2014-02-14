package requirements

import (
	"cf"
	"cf/configuration"
	"cf/terminal"
)

type ApiEndpointRequirement struct {
	ui     terminal.UI
	config configuration.Reader
}

func NewApiEndpointRequirement(ui terminal.UI, config configuration.Reader) ApiEndpointRequirement {
	return ApiEndpointRequirement{ui, config}
}

func (req ApiEndpointRequirement) Execute() (success bool) {
	if req.config.ApiEndpoint() == "" {
		req.ui.Say("No API endpoint targeted. Use '%s' to target an endpoint.", terminal.CommandColor(cf.Name()+" api"))
		return false
	}
	return true
}
