package commands

import (
	/*	"cf"*/
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	/*	"errors"
		"fmt"*/
	"github.com/codegangsta/cli"
	/*	"strings"*/
)

type CreateOrganization struct {
	ui term.UI
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
	name := c.String("name")
	cmd.createOrganization(name)
}

func (cmd CreateOrganization) createOrganization(name string) {
	cmd.ui.Say("Creating organization %s", term.Cyan(name))
	_, apiErr := cmd.orgRepo.Create(name)
	if apiErr != nil {
		cmd.ui.Failed("Error creating organization", apiErr)
		return
	}

	cmd.ui.Ok()
	return
}
