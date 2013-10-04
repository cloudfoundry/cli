package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/net"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

const maxLoginTries = 3

type Login struct {
	ui            terminal.UI
	config        *configuration.Configuration
	configRepo    configuration.ConfigurationRepository
	authenticator api.AuthenticationRepository
}

func NewLogin(ui terminal.UI, configRepo configuration.ConfigurationRepository, authenticator api.AuthenticationRepository) (cmd Login) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.config, _ = configRepo.Get()
	cmd.authenticator = authenticator
	return
}

func (cmd Login) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Login) Run(c *cli.Context) {
	cmd.ui.Say("API endpoint: %s", terminal.EntityNameColor(cmd.config.Target))

	var (
		username string
		password string
	)

	if len(c.Args()) > 0 {
		username = c.Args()[0]
	} else {
		username = cmd.ui.Ask("Username%s", terminal.PromptColor(">"))
	}

	if len(c.Args()) > 1 {
		password = c.Args()[1]
		cmd.ui.Say("Authenticating...")

		apiStatus := cmd.doLogin(username, password)
		if apiStatus.NotSuccessful() {
			cmd.ui.Failed(apiStatus.Message)
			return
		}

	} else {
		for i := 0; i < maxLoginTries; i++ {
			password = cmd.ui.AskForPassword("Password%s", terminal.PromptColor(">"))
			cmd.ui.Say("Authenticating...")

			apiStatus := cmd.doLogin(username, password)
			if apiStatus.NotSuccessful() {
				cmd.ui.Failed(apiStatus.Message)
				continue
			}

			return
		}
	}
	return
}

func (cmd Login) doLogin(username, password string) (apiStatus net.ApiStatus) {
	apiStatus = cmd.authenticator.Authenticate(username, password)
	if !apiStatus.NotSuccessful() {
		cmd.ui.Ok()
		cmd.ui.Say("Use '%s' to view or set your target organization and space", terminal.CommandColor(cf.Name+" target"))
	}
	return
}
