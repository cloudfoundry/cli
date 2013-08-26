package commands

import (
	"cf"
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

		organizations, err := l.orgRepo.FindAll(config)

		if err != nil {
			l.ui.Failed("Error fetching organizations.", err)
			return
		}

		if len(organizations) == 0 {
			l.ui.Say("No orgs found. Use 'cf create-organization' as an Administrator.")
			return
		}

		l.targetOrganization(config, organizations)

		spaces, err := l.spaceRepo.FindAll(config)

		if err != nil {
			l.ui.Failed("Error fetching spaces.", err)
			return
		}

		if len(spaces) == 0 {
			l.ui.ShowConfigurationNoSpacesAvailable(config)
			return
		}

		l.targetSpace(config, spaces)
		l.ui.ShowConfiguration(config)

		return
	}
}

func (l Login) targetOrganization(config *configuration.Configuration, organizations []cf.Organization) {
	var selectedOrg cf.Organization

	if len(organizations) == 1 {
		selectedOrg = organizations[0]
	} else {
		selectedOrg = l.chooseOrg(organizations)
	}

	l.ui.Say("Targeting org %s...", term.Cyan(selectedOrg.Name))
	err := l.saveOrg(config, selectedOrg)

	if err == nil {
		l.ui.Ok()
	}
}

func (l Login) chooseOrg(orgs []cf.Organization) (org cf.Organization) {
	for i, org := range orgs {
		l.ui.Say("%s: %s", term.Green(strconv.Itoa(i+1)), org.Name)
	}

	index, err := strconv.Atoi(l.ui.Ask("Organization%s", term.Cyan(">")))

	if err != nil || index > len(orgs) {
		l.ui.Failed("Invalid number", err)
		return l.chooseOrg(orgs)
	}

	return orgs[index-1]
}

func (l Login) saveOrg(config *configuration.Configuration, org cf.Organization) (err error) {
	config.Organization = org
	err = config.Save()

	if err != nil {
		l.ui.Failed("Error saving organization: %s", err)
		return
	}

	return
}

func (l Login) targetSpace(config *configuration.Configuration, spaces []cf.Space) {
	if len(spaces) == 1 {
		l.saveSpace(config, spaces[0])
	} else {
		selectedSpace := l.chooseSpace(spaces)
		l.ui.Say("Targeting space %s...", term.Cyan(selectedSpace.Name))
		err := l.saveSpace(config, selectedSpace)

		if err == nil {
			l.ui.Ok()
		}
	}
}

func (l Login) chooseSpace(spaces []cf.Space) (space cf.Space) {
	for i, space := range spaces {
		l.ui.Say("%s: %s", term.Green(strconv.Itoa(i+1)), space.Name)
	}

	index, err := strconv.Atoi(l.ui.Ask("Space%s", term.Cyan(">")))

	if err != nil || index > len(spaces) {
		l.ui.Failed("Invalid number", err)
		return l.chooseSpace(spaces)
	}

	return spaces[index-1]
}

func (l Login) saveSpace(config *configuration.Configuration, space cf.Space) (err error) {
	config.Space = space
	err = config.Save()

	if err != nil {
		l.ui.Failed("Error saving organization: %s", err)
		return
	}

	return
}
