package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type Api struct {
	ui         term.UI
	configRepo configuration.ConfigurationRepository
	config     *configuration.Configuration
}

func NewApi(ui term.UI, configRepo configuration.ConfigurationRepository) (cmd Api) {
	cmd.ui = ui
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
		term.Yellow(cmd.config.Target),
		term.Yellow(cmd.config.ApiVersion),
	)
}

func (cmd Api) setNewApiEndpoint(endpoint string) {
	cmd.ui.Say("Setting api endpoint to %s...", term.Yellow(endpoint))

	request, apiErr := api.NewRequest("GET", endpoint+"/v2/info", "", nil)

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
	apiErr = api.PerformRequestAndParseResponse(request, &serverResponse)

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
		cmd.ui.Say(term.Magenta("\nWarning: Insecure http API Endpoint detected. Secure https API Endpoints are recommended.\n"))
	}

	cmd.showApiEndpoint()
}

func (cmd Api) saveEndpoint(endpoint string, info *InfoResponse) (err error) {
	cmd.configRepo.ClearSession()
	cmd.config.Target = endpoint
	cmd.config.ApiVersion = info.ApiVersion
	cmd.config.AuthorizationEndpoint = info.AuthorizationEndpoint
	return cmd.configRepo.Save()
}
