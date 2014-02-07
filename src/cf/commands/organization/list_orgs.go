package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/models"
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
	cmd.ui.Say("Getting orgs as %s...\n", terminal.EntityNameColor(cmd.config.Username()))

	noOrgs := true
	table := cmd.ui.Table([]string{"name"})

	apiStatus := cmd.orgRepo.ListOrgs(func(org models.Organization) bool {
		table.Print([][]string{{org.Name}})
		noOrgs = false
		return true
	})

	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching orgs.\n%s", apiStatus.Message)
		return
	}

	if noOrgs {
		cmd.ui.Say("No orgs found")
	}
}
