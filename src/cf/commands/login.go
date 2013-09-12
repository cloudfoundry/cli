package commands

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
	"strconv"
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
	l.ui.Say("target: %s", term.Cyan(l.config.Target))

	var (
		username string
		password string
	)

	if len(c.Args()) > 0 {
		username = c.Args()[0]
	} else {
		username = l.ui.Ask("Username%s", term.Cyan(">"))
	}

	for i := 0; i < maxLoginTries; i++ {
		if len(c.Args()) > 1 {
			password = c.Args()[1]
		} else {
			password = l.ui.AskForPassword("Password%s", term.Cyan(">"))
		}
		l.ui.Say("Authenticating...")

		err := l.authenticator.Authenticate(username, password)

		if err != nil {
			l.ui.Failed(err.Error())
			continue
		}

		l.ui.Ok()

		organizations, err := l.orgRepo.FindAll()

		if err != nil {
			l.ui.Failed(err.Error())
			return
		}

		if len(organizations) == 0 {
			l.ui.Say("No orgs found. Use 'cf create-organization' as an Administrator.")
			return
		}

		l.targetOrganization(l.config, organizations)

		spaces, err := l.spaceRepo.FindAll()

		if err != nil {
			l.ui.Failed(err.Error())
			return
		}

		if len(spaces) == 0 {
			l.ui.ShowConfiguration(l.config)
			l.ui.Say("No spaces found. Use 'cf create-space' as an Org Manager.")
			return
		}

		l.targetSpace(l.config, spaces)
		l.ui.ShowConfiguration(l.config)

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
		l.ui.Failed("Invalid number")
		return l.chooseOrg(orgs)
	}

	return orgs[index-1]
}

func (l Login) saveOrg(config *configuration.Configuration, org cf.Organization) (err error) {
	config.Organization = org
	err = l.configRepo.Save()

	if err != nil {
		l.ui.Failed(err.Error())
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
		l.ui.Failed("Invalid number")
		return l.chooseSpace(spaces)
	}

	return spaces[index-1]
}

func (l Login) saveSpace(config *configuration.Configuration, space cf.Space) (err error) {
	config.Space = space
	err = l.configRepo.Save()

	if err != nil {
		l.ui.Failed(err.Error())
		return
	}

	return
}
