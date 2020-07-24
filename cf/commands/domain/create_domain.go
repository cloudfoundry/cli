package domain

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
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

func (cmd *CreateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires org_name, domain_name as arguments\n\n") + commandregistry.Commands.CommandUsage("create-domain"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs, nil
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
