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
	config  configuration.Reader
	orgRepo api.OrganizationRepository
}

func NewListOrgs(ui terminal.UI, config configuration.Reader, orgRepo api.OrganizationRepository) (cmd ListOrgs) {
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd ListOrgs) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListOrgs) Run(c *cli.Context) {
	cmd.ui.Say("Getting orgs as %s...\n", terminal.EntityNameColor(cmd.config.Username()))

	noOrgs := true
	table := cmd.ui.Table([]string{"name"})

	apiErr := cmd.orgRepo.ListOrgs(func(org models.Organization) bool {
		table.Print([][]string{{org.Name}})
		noOrgs = false
		return true
	})

	if apiErr != nil {
		cmd.ui.Failed("Failed fetching orgs.\n%s", apiErr)
		return
	}

	if noOrgs {
		cmd.ui.Say("No orgs found")
	}
}
