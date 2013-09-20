package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListOrganizations struct {
	ui      terminal.UI
	orgRepo api.OrganizationRepository
}

func NewListOrganizations(ui terminal.UI, orgRepo api.OrganizationRepository) (cmd ListOrganizations) {
	cmd.ui = ui
	cmd.orgRepo = orgRepo
	return
}

func (cmd ListOrganizations) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListOrganizations) Run(c *cli.Context) {
	cmd.ui.Say("Getting organizations...")

	orgs, err := cmd.orgRepo.FindAll()
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()

	for _, org := range orgs {
		cmd.ui.Say(org.Name)
	}
}
