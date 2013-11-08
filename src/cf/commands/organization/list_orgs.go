package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/paginator"
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
	cmd.ui.Say("")

	p := cmd.orgRepo.Paginator()
	noResults := paginator.ForEach(p, cmd.ui.PrintPaginator)

	if noResults {
		cmd.ui.Say("No orgs found")
	}
}
