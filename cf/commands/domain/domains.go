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
)

type ListDomains struct {
	ui         terminal.UI
	config     core_config.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func init() {
	command_registry.Register(&ListDomains{})
}

func (cmd *ListDomains) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "domains",
		Description: T("List domains in the target org"),
		Usage:       "CF_NAME domains",
	}
}

func (cmd *ListDomains) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 0 {
		cmd.ui.Failed(T("Incorrect Usage. No argument required\n\n") + command_registry.Commands.CommandUsage("domains"))
	}

	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *ListDomains) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	return cmd
}

func (cmd *ListDomains) Execute(c flags.FlagContext) {
	org := cmd.orgReq.GetOrganizationFields()

	cmd.ui.Say(T("Getting domains in org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	domains := cmd.fetchAllDomains(org.Guid)
	cmd.printDomainsTable(domains)

	if len(domains) == 0 {
		cmd.ui.Say(T("No domains found"))
	}
}

func (cmd *ListDomains) fetchAllDomains(orgGuid string) (domains []models.DomainFields) {
	apiErr := cmd.domainRepo.ListDomainsForOrg(orgGuid, func(domain models.DomainFields) bool {
		domains = append(domains, domain)
		return true
	})
	if apiErr != nil {
		cmd.ui.Failed(T("Failed fetching domains.\n{{.ApiErr}}", map[string]interface{}{"ApiErr": apiErr.Error()}))
	}
	return
}

func (cmd *ListDomains) printDomainsTable(domains []models.DomainFields) {
	table := cmd.ui.Table([]string{T("name"), T("status")})

	for _, domain := range domains {
		if domain.Shared {
			table.Add(domain.Name, T("shared"))
		}
	}

	for _, domain := range domains {
		if !domain.Shared {
			table.Add(domain.Name, T("owned"))
		}
	}
	table.Print()
}
