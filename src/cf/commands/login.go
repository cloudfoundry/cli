package commands

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

const maxLoginTries = 3

type Login struct {
	ui            term.UI
	config        *configuration.Configuration
	configRepo    configuration.ConfigurationRepository
	orgRepo       api.OrganizationRepository
	spaceRepo     api.SpaceRepository
	authenticator api.Authenticator
}

func NewLogin(ui term.UI, configRepo configuration.ConfigurationRepository, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository, authenticator api.Authenticator) (l Login) {
	l.ui = ui
	l.configRepo = configRepo
	l.config, _ = configRepo.Get()
	l.orgRepo = orgRepo
	l.spaceRepo = spaceRepo
	l.authenticator = authenticator
	return
}

func (cmd Login) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (l Login) Run(c *cli.Context) {
	l.ui.Say("API endpoint: %s", term.EntityNameColor(l.config.Target))

	var (
		username string
		password string
	)

	if len(c.Args()) > 0 {
		username = c.Args()[0]
	} else {
		username = l.ui.Ask("Username%s", term.PromptColor(">"))
	}

	for i := 0; i < maxLoginTries; i++ {
		if len(c.Args()) > 1 {
			password = c.Args()[1]
		} else {
			password = l.ui.AskForPassword("Password%s", term.PromptColor(">"))
		}
		l.ui.Say("Authenticating...")

		err := l.authenticator.Authenticate(username, password)

		if err != nil {
			l.ui.Failed(err.Error())
			continue
		}

		l.ui.Ok()

		l.ui.Say("Use '%s' to view or set your target organization and space", term.CommandColor("cf target"))

		return
	}
}
