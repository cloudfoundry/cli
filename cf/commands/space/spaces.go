package space

import (
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/plugin/models"
)

type ListSpaces struct {
	ui        terminal.UI
	config    core_config.Reader
	spaceRepo spaces.SpaceRepository

	pluginModel *[]plugin_models.GetSpaces_Model
	pluginCall  bool
}

func init() {
	command_registry.Register(&ListSpaces{})
}

func (cmd *ListSpaces) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "spaces",
		Description: T("List all spaces in an org"),
		Usage:       T("CF_NAME spaces"),
	}

}

func (cmd *ListSpaces) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("spaces"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *ListSpaces) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Spaces
	return cmd
}

func (cmd *ListSpaces) Execute(c flags.FlagContext) {
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

		if cmd.pluginCall {
			s := plugin_models.GetSpaces_Model{}
			s.Name = space.Name
			s.Guid = space.Guid
			*(cmd.pluginModel) = append(*(cmd.pluginModel), s)
		}

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
