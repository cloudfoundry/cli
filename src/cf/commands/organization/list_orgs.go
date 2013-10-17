package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type ListOrgs struct {
	ui      terminal.UI
	config  *configuration.Configuration
	orgRepo api.OrganizationRepository
}

func NewListOrgs(ui terminal.UI, config *configuration.Configuration, orgRepo api.OrganizationRepository) (cmd ListOrgs) {
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd ListOrgs) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListOrgs) Run(c *cli.Context) {
	cmd.ui.Say("Getting orgs as %s...", terminal.EntityNameColor(cmd.config.Username()))

	orgs, apiResponse := cmd.orgRepo.FindAll()
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	for _, org := range orgs {
		cmd.ui.Say(org.Name)
	}
}
