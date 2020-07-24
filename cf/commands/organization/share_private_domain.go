package organization

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/organizations"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/cf/i18n"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/terminal"
)

type SharePrivateDomain struct {
	ui         terminal.UI
	config     coreconfig.Reader
	orgRepo    organizations.OrganizationRepository
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func init() {
	commandregistry.Register(&SharePrivateDomain{})
}

func (cmd *SharePrivateDomain) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "share-private-domain",
		Description: T("Share a private domain with an org"),
		Usage: []string{
			T("CF_NAME share-private-domain ORG DOMAIN"),
		},
	}
}

func (cmd *SharePrivateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires ORG and DOMAIN as arguments\n\n") + commandregistry.Commands.CommandUsage("share-private-domain"))
		return nil, fmt.Errorf("Incorrect usage: %d arguments of %d required", len(fc.Args()), 2)
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs, nil
}

func (cmd *SharePrivateDomain) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *SharePrivateDomain) Execute(c flags.FlagContext) error {
	org := cmd.orgReq.GetOrganization()
	domainName := c.Args()[1]
	domain, err := cmd.domainRepo.FindPrivateByName(domainName)

	if err != nil {
		return err
	}

	cmd.ui.Say(T("Sharing domain {{.DomainName}} with org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domain.Name),
			"OrgName":    terminal.EntityNameColor(org.Name),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	err = cmd.orgRepo.SharePrivateDomain(org.GUID, domain.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
