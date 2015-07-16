package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type CreateDomain struct {
	ui         terminal.UI
	config     core_config.Reader
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&CreateDomain{})
}

func (cmd *CreateDomain) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "create-domain",
		Description: T("Create a domain in an org for later use"),
		Usage:       T("CF_NAME create-domain ORG DOMAIN"),
	}
}

func (cmd *CreateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires org_name, domain_name as arguments\n\n") + command_registry.Commands.CommandUsage("create-domain"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *CreateDomain) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *CreateDomain) Execute(c flags.FlagContext) {
	domainName := c.Args()[1]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Creating domain {{.DomainName}} for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"OrgName":    terminal.EntityNameColor(owningOrg.Name),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	_, apiErr := cmd.domainRepo.Create(domainName, owningOrg.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
