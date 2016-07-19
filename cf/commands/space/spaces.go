package space

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/spaces"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/plugin/models"
)

type ListSpaces struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceRepo spaces.SpaceRepository

	pluginModel *[]plugin_models.GetSpaces_Model
	pluginCall  bool
}

func init() {
	commandregistry.Register(&ListSpaces{})
}

func (cmd *ListSpaces) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "spaces",
		Description: T("List all spaces in an org"),
		Usage: []string{
			T("CF_NAME spaces"),
		},
	}

}

func (cmd *ListSpaces) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	return reqs, nil
}

func (cmd *ListSpaces) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.pluginCall = pluginCall
	cmd.pluginModel = deps.PluginModels.Spaces
	return cmd
}

func (cmd *ListSpaces) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting spaces in org {{.TargetOrgName}} as {{.CurrentUser}}...\n",
		map[string]interface{}{
			"TargetOrgName": terminal.EntityNameColor(cmd.config.OrganizationFields().Name),
			"CurrentUser":   terminal.EntityNameColor(cmd.config.Username()),
		}))

	foundSpaces := false
	table := cmd.ui.Table([]string{T("name")})
	err := cmd.spaceRepo.ListSpaces(func(space models.Space) bool {
		table.Add(space.Name)
		foundSpaces = true

		if cmd.pluginCall {
			s := plugin_models.GetSpaces_Model{}
			s.Name = space.Name
			s.Guid = space.GUID
			*(cmd.pluginModel) = append(*(cmd.pluginModel), s)
		}

		return true
	})
	err = table.Print()
	if err != nil {
		return err
	}

	if err != nil {
		return errors.New(T("Failed fetching spaces.\n{{.ErrorDescription}}",
			map[string]interface{}{
				"ErrorDescription": err.Error(),
			}))
	}

	if !foundSpaces {
		cmd.ui.Say(T("No spaces found"))
	}
	return nil
}
