package commands

import (
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type Api struct {
	ui         terminal.UI
	gateway    net.Gateway
	configRepo configuration.ConfigurationRepository
}

func NewApi(ui terminal.UI, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (cmd Api) {
	cmd.ui = ui
	cmd.gateway = gateway
	cmd.configRepo = configRepo
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

	cmd.setNewApiEndpoint(c.Args()[0])
}

func (cmd Api) showApiEndpoint() {
	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.Failed("Error reading config.\n%s", err.Error())
	}
	cmd.ui.Say(
		"API endpoint: %s (API version: %s)",
		terminal.EntityNameColor(config.Target),
		terminal.EntityNameColor(config.ApiVersion),
	)
}

func (cmd Api) setNewApiEndpoint(endpoint string) {
	cmd.ui.Say("Setting api endpoint to %s...", terminal.EntityNameColor(endpoint))

	request, apiStatus := cmd.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	scheme := request.URL.Scheme
	if scheme != "http" && scheme != "https" {
		cmd.ui.Failed("API Endpoints should start with https:// or http://")
		return
	}

	serverResponse := new(InfoResponse)
	_, apiStatus = cmd.gateway.PerformRequestForJSONResponse(request, &serverResponse)

	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	err := cmd.saveEndpoint(endpoint, serverResponse)

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	if scheme == "http" {
		cmd.ui.Say(terminal.WarningColor("\nWarning: Insecure http API Endpoint detected. Secure https API Endpoints are recommended.\n"))
	}

	cmd.showApiEndpoint()

	cmd.ui.Say(terminal.NotLoggedInText())
}

func (cmd Api) saveEndpoint(endpoint string, info *InfoResponse) (err error) {
	cmd.configRepo.ClearSession()
	config, err := cmd.configRepo.Get()
	if err != nil {
		return
	}
	config.Target = endpoint
	config.ApiVersion = info.ApiVersion
	config.AuthorizationEndpoint = info.AuthorizationEndpoint
	err = cmd.configRepo.Save(config)
	return
}
