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
	config     *configuration.Configuration
}

func NewApi(ui terminal.UI, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (cmd Api) {
	cmd.ui = ui
	cmd.gateway = gateway
	cmd.configRepo = configRepo
	cmd.config, _ = cmd.configRepo.Get()
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
	cmd.ui.Say(
		"API endpoint: %s (API version: %s)",
		terminal.EntityNameColor(cmd.config.Target),
		terminal.EntityNameColor(cmd.config.ApiVersion),
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
	cmd.config.Target = endpoint
	cmd.config.ApiVersion = info.ApiVersion
	cmd.config.AuthorizationEndpoint = info.AuthorizationEndpoint
	return cmd.configRepo.Save()
}
