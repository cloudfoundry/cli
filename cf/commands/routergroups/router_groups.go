package routergroups

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type RouterGroups struct {
	ui             terminal.UI
	routingApiRepo api.RoutingApiRepository
	config         core_config.Reader
}

func init() {
	command_registry.Register(&RouterGroups{})
}

func (cmd *RouterGroups) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "router-groups",
		Description: T("List router groups"),
		Usage:       "CF_NAME router-groups",
	}
}

func (cmd *RouterGroups) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("router-groups"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewRoutingAPIRequirement(),
	}, nil
}

func (cmd *RouterGroups) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.routingApiRepo = deps.RepoLocator.GetRoutingApiRepository()
	return cmd
}

func (cmd *RouterGroups) Execute(c flags.FlagContext) {
	cmd.ui.Say(T("Getting router groups as {{.Username}} ...\n",
		map[string]interface{}{"Username": terminal.EntityNameColor(cmd.config.Username())}))

	table := cmd.ui.Table([]string{T("name"), T("type")})

	noRouterGroups := true
	cb := func(group models.RouterGroup) bool {
		noRouterGroups = false
		table.Add(group.Name, group.Type)
		return true
	}

	apiErr := cmd.routingApiRepo.ListRouterGroups(cb)
	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching router groups.\n{{.Err}}", map[string]interface{}{"Err": apiErr.Error()}))
		return
	}

	if noRouterGroups {
		cmd.ui.Say(T("No router groups found"))
	}

	table.Print()
}
