package organization

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type SharePrivateDomain struct {
	ui         terminal.UI
	config     core_config.Reader
	orgRepo    organizations.OrganizationRepository
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func init() {
	command_registry.Register(&SharePrivateDomain{})
}

func (cmd *SharePrivateDomain) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "share-private-domain",
		Description: T("Share a private domain with an org"),
		Usage:       T("CF_NAME share-private-domain ORG DOMAIN"),
	}
}

func (cmd *SharePrivateDomain) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires ORG and DOMAIN as arguments\n\n") + command_registry.Commands.CommandUsage("share-private-domain"))
	}

	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(fc.Args()[0])

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}, nil
}

func (cmd *SharePrivateDomain) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.orgRepo = deps.RepoLocator.GetOrganizationRepository()
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *SharePrivateDomain) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganization()
	domainName := c.Args()[1]
	domain, err := cmd.domainRepo.FindPrivateByName(domainName)

	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Say(T("Sharing domain {{.DomainName}} with org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"DomainName": terminal.EntityNameColor(domain.Name),
			"OrgName":    terminal.EntityNameColor(org.Name),
			"Username":   terminal.EntityNameColor(cmd.config.Username())}))

	err = cmd.orgRepo.SharePrivateDomain(org.Guid, domain.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	cmd.ui.Ok()
}
