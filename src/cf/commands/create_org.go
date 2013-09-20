package commands

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateOrganization struct {
	ui      terminal.UI
	orgRepo api.OrganizationRepository
}

func NewCreateOrganization(ui terminal.UI, orgRepo api.OrganizationRepository) (cmd CreateOrganization) {
	cmd.ui = ui
	cmd.orgRepo = orgRepo
	return
}

func (cmd CreateOrganization) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-org")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd CreateOrganization) Run(c *cli.Context) {
	name := c.Args()[0]

	cmd.ui.Say("Creating organization %s...", terminal.EntityNameColor(name))
	apiErr := cmd.orgRepo.Create(name)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("\nTIP: Use '%s' to target new org.", terminal.CommandColor("cf target -o "+name))
}
