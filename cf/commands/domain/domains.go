package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type ListDomains struct {
	ui         terminal.UI
	config     configuration.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewListDomains(ui terminal.UI, config configuration.Reader, domainRepo api.DomainRepository) (cmd *ListDomains) {
	cmd = new(ListDomains)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = domainRepo
	return
}

func (cmd *ListDomains) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "domains",
		Description: T("List domains in the target org"),
		Usage:       "CF_NAME domains",
	}
}

func (cmd *ListDomains) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 0 {
		cmd.ui.FailWithUsage(c)
	}

	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *ListDomains) Run(c *cli.Context) {
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
