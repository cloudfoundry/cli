package requirements

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
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
		loginTip := terminal.CommandColor(fmt.Sprintf(T("{{.CFName}} login", map[string]interface{}{"CFName": cf.Name()})))
		apiTip := terminal.CommandColor(fmt.Sprintf(T("{{.CFName}} api", map[string]interface{}{"CFName": cf.Name()})))
		req.ui.Say(T("No API endpoint targeted. Use '{{.LoginTip}}' or '{{.APITip}}' to target an endpoint.",
			map[string]interface{}{
				"LoginTip": loginTip,
				"APITip":   apiTip,
			}))
		return false
	}
	return true
}
