package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type CreateSharedDomain struct {
	ui             terminal.UI
	config         core_config.Reader
	domainRepo     api.DomainRepository
	routingApiRepo api.RoutingApiRepository
	orgReq         requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&CreateSharedDomain{})
}

func (cmd *CreateSharedDomain) MetaData() command_registry.CommandMetadata {
	fs := make(map[string]flags.FlagSet)
	fs["r"] = &cliFlags.StringFlag{Name: "r", Usage: T("Routes for this domain will be configured only on the specified router group.")}
	return command_registry.CommandMetadata{
		Name:        "create-shared-domain",
		Description: T("Create a domain that can be used by all orgs (admin-only)"),
		Usage:       T("CF_NAME create-shared-domain DOMAIN"),
		Flags:       fs,
	}
}

func (cmd *CreateSharedDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 1 {
		cmd.ui.Failed(T("Incorrect Usage. Requires an argument\n\n") + command_registry.Commands.CommandUsage("create-shared-domain"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *CreateSharedDomain) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routingApiRepo = deps.RepoLocator.GetRoutingApiRepository()
	return cmd
}

func (cmd *CreateSharedDomain) Execute(c flags.FlagContext) {
	var routerGroup models.RouterGroup
	domainName := c.Args()[0]

	if c.String("r") != "" {
		apiErr := cmd.routingApiRepo.ListRouterGroups(func(group models.RouterGroup) bool {
			if group.Name == c.String("r") {
				routerGroup = group
				return false
			}

			return true
		})

		if apiErr != nil {
			cmd.ui.Failed(apiErr.Error())
			return
		}

		if routerGroup.Guid == "" {
			cmd.ui.Failed(T("Router group not found"))
		}
	}

	cmd.ui.Say(T("Creating shared domain {{.DomainName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	apiErr := cmd.domainRepo.CreateSharedDomain(domainName, routerGroup.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
