package commands

import (
	"cf/api"
	"cf/requirements"
	term "cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type RenameOrg struct {
	ui      term.UI
	orgRepo api.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func NewRenameOrg(ui term.UI, orgRepo api.OrganizationRepository) (cmd *RenameOrg) {
	cmd = new(RenameOrg)
	cmd.ui = ui
	cmd.orgRepo = orgRepo
	return
}

func (cmd *RenameOrg) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-org")
		return
	}
	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *RenameOrg) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	cmd.ui.Say("Renaming org %s...", term.EntityNameColor(org.Name))

	err := cmd.orgRepo.Rename(org, c.Args()[1])
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()
}
