package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type Authenticate struct {
	ui            terminal.UI
	config        *configuration.Configuration
	configRepo    configuration.ConfigurationRepository
	authenticator api.AuthenticationRepository
}

func NewAuthenticate(ui terminal.UI, configRepo configuration.ConfigurationRepository, authenticator api.AuthenticationRepository) (cmd Authenticate) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.config, _ = configRepo.Get()
	cmd.authenticator = authenticator
	return
}

func (cmd Authenticate) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) < 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "auth")
		return
	}
	return
}

func (cmd Authenticate) Run(c *cli.Context) {
	cmd.ui.Say("API endpoint: %s", terminal.EntityNameColor(cmd.config.Target))

	username := c.Args()[0]
	password := c.Args()[1]

	cmd.ui.Say("Authenticating...")

	apiResponse := cmd.doLogin(username, password)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	return
}

func (cmd Authenticate) doLogin(username, password string) (apiResponse net.ApiResponse) {
	apiResponse = cmd.authenticator.Authenticate(username, password)
	if apiResponse.IsSuccessful() {
		cmd.ui.Ok()
		cmd.ui.Say("Use '%s' to view or set your target org and space", terminal.CommandColor(cf.Name+" target"))
	}
	return
}
