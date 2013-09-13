package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateOrganization struct {
	ui      term.UI
	orgRepo api.OrganizationRepository
}

func NewCreateOrganization(ui term.UI, orgRepo api.OrganizationRepository) (cmd CreateOrganization) {
	cmd.ui = ui
	cmd.orgRepo = orgRepo
	return
}

func (cmd CreateOrganization) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	return
}

func (cmd CreateOrganization) Run(c *cli.Context) {
	argsCount := len(c.Args())

	if argsCount == 0 {
		cmd.ui.Say("No name provided. Use 'cf create-org <name>' create an organization.")
		return
	}

	name := c.Args()[0]

	cmd.ui.Say("Creating organization %s", term.Cyan(name))
	_, apiErr := cmd.orgRepo.Create(name)
	if apiErr != nil {
		cmd.ui.Failed("Error creating organization", apiErr)
		return
	}

	cmd.ui.Ok()
	return
}
