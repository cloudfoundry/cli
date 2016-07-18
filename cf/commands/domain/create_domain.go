package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type CreateDomain struct {
	ui         terminal.UI
	config     coreconfig.Reader
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&CreateDomain{})
}

func (cmd *CreateDomain) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "create-domain",
		Description: T("Create a domain in an org for later use"),
		Usage: []string{
			T("CF_NAME create-domain ORG DOMAIN"),
		},
	}
}

func (cmd *CreateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires org_name, domain_name as arguments\n\n") + commandregistry.Commands.CommandUsage("create-domain"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs
}

func (cmd *CreateDomain) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *CreateDomain) Execute(c flags.FlagContext) error {
	domainName := c.Args()[1]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say(T("Creating domain {{.DomainName}} for org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domainName),
			"OrgName":    terminal.EntityNameColor(owningOrg.Name),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	_, err := cmd.domainRepo.Create(domainName, owningOrg.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
