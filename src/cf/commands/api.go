package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
	"strings"
)

type Api struct {
	ui           terminal.UI
	endpointRepo api.EndpointRepository
	config       *configuration.Configuration
}

type ApiEndpointSetter interface {
	SetApiEndpoint(endpoint string)
}

func NewApi(ui terminal.UI, config *configuration.Configuration, endpointRepo api.EndpointRepository) (cmd Api) {
	cmd.ui = ui
	cmd.config = config
	cmd.endpointRepo = endpointRepo
	return
}

func (cmd Api) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Api) Run(c *cli.Context) {
	if len(c.Args()) == 0 {
		cmd.showApiEndpoint()
		return
	}

	cmd.SetApiEndpoint(c.Args()[0])
}

func (cmd Api) showApiEndpoint() {
	cmd.ui.Say(
		"API endpoint: %s (API version: %s)",
		terminal.EntityNameColor(cmd.config.Target),
		terminal.EntityNameColor(cmd.config.ApiVersion),
	)
}

func (cmd Api) SetApiEndpoint(endpoint string) {
	cmd.ui.Say("Setting api endpoint to %s...", terminal.EntityNameColor(endpoint))

	apiResponse := cmd.endpointRepo.UpdateEndpoint(endpoint)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()

	if !strings.HasPrefix(endpoint, "https://") {
		cmd.ui.Say(terminal.WarningColor("\nWarning: Insecure http API endpoint detected: secure https API endpoints are recommended\n"))
	}

	cmd.showApiEndpoint()

	cmd.ui.Say(terminal.NotLoggedInText())
}
