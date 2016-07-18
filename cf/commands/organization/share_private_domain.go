package organization

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (cmd *SharePrivateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires ORG and DOMAIN as arguments\n\n") + commandregistry.Commands.CommandUsage("share-private-domain"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}

	return reqs
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
