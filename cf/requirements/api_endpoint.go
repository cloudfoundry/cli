package requirements

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/terminal"
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
		loginTip := terminal.CommandColor(fmt.Sprintf("%s login", cf.Name()))
		apiTip := terminal.CommandColor(fmt.Sprintf("%s api", cf.Name()))
		req.ui.Say("No API endpoint targeted. Use '%s' or '%s' to target an endpoint.", loginTip, apiTip)
		return false
	}
	return true
}
