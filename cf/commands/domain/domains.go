package domain

import (
	"strings"

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
	ui             terminal.UI
	config         core_config.Reader
	domainRepo     api.DomainRepository
	routingApiRepo api.RoutingApiRepository
}

func init() {
	command_registry.Register(&ListDomains{})
}

func (cmd *ListDomains) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "domains",
		Description: T("List domains in the target org"),
		Usage: []string{
			"CF_NAME domains",
		},
	}
}

func (cmd *ListDomains) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(command_registry.CliCommandUsagePresenter(cmd),
		T("No argument required"),
		func() bool {
			return len(fc.Args()) != 0
		},
	)

	reqs := []requirements.Requirement{
		usageReq,
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	return reqs
}

func (cmd *ListDomains) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routingApiRepo = deps.RepoLocator.GetRoutingApiRepository()

	return cmd
}

func (cmd *ListDomains) Execute(c flags.FlagContext) {
	org := cmd.config.OrganizationFields()

	cmd.ui.Say(T("Getting domains in org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	domains, err := cmd.getDomains(org.Guid)
	if err != nil {
		cmd.ui.Failed(T("Failed fetching domains.\n{{.Error}}", map[string]interface{}{"Error": err.Error()}))
	}

	table := cmd.ui.Table([]string{T("name"), T("status"), T("type")})

	for _, domain := range domains {
		if domain.Shared {
			table.Add(domain.Name, T("shared"), strings.Join(domain.RouterGroupTypes, ", "))
		}
	}

	for _, domain := range domains {
		if !domain.Shared {
			table.Add(domain.Name, T("owned"), strings.Join(domain.RouterGroupTypes, ", "))
		}
	}

	table.Print()

	if len(domains) == 0 {
		cmd.ui.Say(T("No domains found"))
	}
}

func (cmd *ListDomains) getDomains(orgGuid string) ([]models.DomainFields, error) {
	domains := []models.DomainFields{}
	err := cmd.domainRepo.ListDomainsForOrg(orgGuid, func(domain models.DomainFields) bool {
		domains = append(domains, domain)
		return true
	})

	if err != nil {
		return []models.DomainFields{}, err
	}

	return domains, nil
}
