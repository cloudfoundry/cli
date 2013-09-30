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
	configRepo    configuration.ConfigurationRepository
	authenticator api.Authenticator
}

func NewLogin(ui terminal.UI, configRepo configuration.ConfigurationRepository, authenticator api.Authenticator) (cmd Login) {
	cmd.ui = ui
	cmd.configRepo = configRepo
	cmd.authenticator = authenticator
	return
}

func (cmd Login) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd Login) Run(c *cli.Context) {
	config, err := cmd.configRepo.Get()
	if err != nil {
		cmd.ui.ConfigFailure(err)
	}

	cmd.ui.Say("API endpoint: %s", terminal.EntityNameColor(config.Target))

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

		apiErr := cmd.doLogin(username, password)
		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}

	} else {
		for i := 0; i < maxLoginTries; i++ {
			password = cmd.ui.AskForPassword("Password%s", terminal.PromptColor(">"))
			cmd.ui.Say("Authenticating...")

			apiErr := cmd.doLogin(username, password)
			if apiErr != nil {
				cmd.ui.Failed(apiErr.Error())
				continue
			}

			return
		}
	}
	return
}

func (cmd Login) doLogin(username, password string) (apiErr *net.ApiError) {
	apiErr = cmd.authenticator.Authenticate(username, password)
	if apiErr == nil {
		cmd.ui.Ok()
		cmd.ui.Say("Use '%s' to view or set your target organization and space", terminal.CommandColor(cf.Name+" target"))
	}
	return
}
