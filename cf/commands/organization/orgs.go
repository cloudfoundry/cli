package organization

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (cmd ListOrgs) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "orgs",
		ShortName:   "o",
		Description: T("List all orgs"),
		Usage:       "CF_NAME orgs",
	}
}

func (cmd ListOrgs) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd ListOrgs) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting orgs as {{.Username}}...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	noOrgs := true
	table := cmd.ui.Table([]string{T("name")})

	apiErr := cmd.orgRepo.ListOrgs(func(org models.Organization) bool {
		table.Add(org.Name)
		noOrgs = false
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching orgs.\n{{.ApiErr}}",
			map[string]interface{}{"ApiErr": apiErr}))
		return
	}

	if noOrgs {
		cmd.ui.Say(T("No orgs found"))
	}
}
