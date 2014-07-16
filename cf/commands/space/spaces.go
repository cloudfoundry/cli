package space

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListSpaces struct {
	ui        terminal.UI
	config    configuration.Reader
	spaceRepo spaces.SpaceRepository
}

func NewListSpaces(ui terminal.UI, config configuration.Reader, spaceRepo spaces.SpaceRepository) (cmd ListSpaces) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRepo = spaceRepo
	return
}

func (cmd ListSpaces) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "spaces",
		Description: T("List all spaces in an org"),
		Usage:       T("CF_NAME spaces"),
	}
}

func (cmd ListSpaces) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd ListSpaces) Run(c *cli.Context) {
	cmd.ui.Say(T("Getting spaces in org {{.TargetOrgName}} as {{.CurrentUser}}...\n",
		map[string]interface{}{
			"TargetOrgName": terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser":   terminal.EntityNameColor(cmd.config.Username()),
		}))

	foundSpaces := false
	table := cmd.ui.Table([]string{T("name")})
	apiErr := cmd.spaceRepo.ListSpaces(func(space models.Space) bool {
		table.Add(space.Name)
		foundSpaces = true
		return true
	})
	table.Print()

	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching spaces.\n{{.ErrorDescription}}",
			map[string]interface{}{
				"ErrorDescription": apiErr.Error(),
			}))
		return
	}

	if !foundSpaces {
		cmd.ui.Say(T("No spaces found"))
	}
}
