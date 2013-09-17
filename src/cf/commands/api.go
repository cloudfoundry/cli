package commands

import (
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Api struct {
	ui         term.UI
	gateway    net.Gateway
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
}

func NewApi(ui term.UI, gateway net.Gateway, configRepo configuration.ConfigurationRepository) (cmd Api) {
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
		term.EntityNameColor(cmd.config.Target),
		term.EntityNameColor(cmd.config.ApiVersion),
	)
}

func (cmd Api) setNewApiEndpoint(endpoint string) {
	cmd.ui.Say("Setting api endpoint to %s...", term.EntityNameColor(endpoint))

	request, apiErr := cmd.gateway.NewRequest("GET", endpoint+"/v2/info", "", nil)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	scheme := request.URL.Scheme
	if scheme != "http" && scheme != "https" {
		cmd.ui.Failed("API Endpoints should start with https:// or http://")
		return
	}

	serverResponse := new(InfoResponse)
	apiErr = cmd.gateway.PerformRequestForJSONResponse(request, &serverResponse)

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	err := cmd.saveEndpoint(endpoint, serverResponse)

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	if scheme == "http" {
		cmd.ui.Say(term.WarningColor("\nWarning: Insecure http API Endpoint detected. Secure https API Endpoints are recommended.\n"))
	}

	cmd.showApiEndpoint()

	cmd.ui.Say(term.NotLoggedInText())
}

func (cmd Api) saveEndpoint(endpoint string, info *InfoResponse) (err error) {
	cmd.configRepo.ClearSession()
	cmd.config.Target = endpoint
	cmd.config.ApiVersion = info.ApiVersion
	cmd.config.AuthorizationEndpoint = info.AuthorizationEndpoint
	return cmd.configRepo.Save()
}
