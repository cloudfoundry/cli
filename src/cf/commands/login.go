package commands

import (
	"cf/api"
	"cf/configuration"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
)

const maxLoginTries = 3

type Login struct {
	ui            term.UI
	orgRepo       api.OrganizationRepository
	spaceRepo     api.SpaceRepository
	authenticator api.Authenticator
}

func NewLogin(ui term.UI, orgRepo api.OrganizationRepository, spaceRepo api.SpaceRepository, autenticator api.Authenticator) (l Login) {
	l.ui = ui
	l.orgRepo = orgRepo
	l.spaceRepo = spaceRepo
	l.authenticator = autenticator
	return
}

func (l Login) Run(c *cli.Context) {
	config, err := configuration.Load()
	if err != nil {
		l.ui.Failed("Error loading configuration", err)
		return
	}

	l.ui.Say("target: %s", term.Cyan(config.Target))
	email := l.ui.Ask("Email%s", term.Cyan(">"))

	for i := 0; i < maxLoginTries; i++ {
		password := l.ui.Ask("Password%s", term.Cyan(">"))
		l.ui.Say("Authenticating...")

		err := l.authenticator.Authenticate(config, email, password)

		if err != nil {
			l.ui.Failed("Error Authenticating", err)
			continue
		}

		l.ui.Ok()
		l.targetOrganization(config)
		l.targetSpace(config)
		l.ui.ShowConfiguration(config)

		return
	}
}

func (l Login) targetOrganization(config *configuration.Configuration) {
	organizations, err := l.orgRepo.FindOrganizations(config)

	if err != nil {
		l.ui.Failed("Error fetching organizations.", err)
		return
	}

	if len(organizations) < 2 {
		return
	}

	for i, org := range organizations {
		l.ui.Say("%s: %s", term.Green(strconv.Itoa(i+1)), org.Name)
	}

	index, err := strconv.Atoi(l.ui.Ask("Organization%s", term.Cyan(">")))

	if err != nil || index > len(organizations) {
		l.ui.Failed("Invalid number", err)
		l.targetOrganization(config)
		return
	}

	selectedOrg := organizations[index-1]
	config.Organization = selectedOrg
	err = config.Save()

	if err != nil {
		l.ui.Failed("Error saving organization: %s", err)
		return
	}

	l.ui.Say("Targeting org %s...", term.Cyan(selectedOrg.Name))
	l.ui.Ok()
}

func (l Login) targetSpace(config *configuration.Configuration) {
	spaces, err := l.spaceRepo.FindSpaces(config)

	if len(spaces) < 2 {
		return
	}

	if err != nil {
		l.ui.Failed("Error fetching spaces.", err)
		return
	}

	for i, space := range spaces {
		l.ui.Say("%s: %s", term.Green(strconv.Itoa(i+1)), space.Name)
	}

	index, err := strconv.Atoi(l.ui.Ask("Space%s", term.Cyan(">")))

	if err != nil || index > len(spaces) {
		l.ui.Failed("Invalid number", err)
		l.targetSpace(config)
		return
	}

	selectedSpace := spaces[index-1]
	config.Space = selectedSpace
	err = config.Save()

	if err != nil {
		l.ui.Failed("Error saving organization: %s", err)
		return
	}

	l.ui.Say("Targeting space %s...", term.Cyan(selectedSpace.Name))
	l.ui.Ok()
}
