package requirements

import (
	"cf"
	"cf/configuration"
	"cf/terminal"
	"fmt"
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
		loginTip := terminal.CommandColor(fmt.Sprintf("%s api", cf.Name()))
		targetTip := terminal.CommandColor(fmt.Sprintf("%s target", cf.Name()))
		req.ui.Say("No API endpoint targeted. Use '%s' or '%s' to target an endpoint.", loginTip, targetTip)
		return false
	}
	return true
}
