package routergroups

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type RouterGroups struct {
	ui             terminal.UI
	routingAPIRepo api.RoutingAPIRepository
	config         coreconfig.Reader
}

func init() {
	commandregistry.Register(&RouterGroups{})
}

func (cmd *RouterGroups) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "router-groups",
		Description: T("List router groups"),
		Usage: []string{
			"CF_NAME router-groups",
		},
	}
}

func (cmd *RouterGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	return []requirements.Requirement{
		requirementsFactory.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
			T("No argument required"),
			func() bool {
				return len(fc.Args()) != 0
			},
		),
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewRoutingAPIRequirement(),
	}, nil
}

func (cmd *RouterGroups) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.routingAPIRepo = deps.RepoLocator.GetRoutingAPIRepository()
	return cmd
}

func (cmd *RouterGroups) Execute(c flags.FlagContext) error {
	cmd.ui.Say(T("Getting router groups as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("name"), T("type")})

	noRouterGroups := true
	cb := func(group models.RouterGroup) bool {
		noRouterGroups = false
		table.Add(group.Name, group.Type)
		return true
	}

	apiErr := cmd.routingAPIRepo.ListRouterGroups(cb)
	if apiErr != nil {
		return errors.New(T("Failed fetching router groups.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
	}

	if noRouterGroups {
		cmd.ui.Say(T("No router groups found"))
	}

	err := table.Print()
	if err != nil {
		return err
	}
	return nil
}
