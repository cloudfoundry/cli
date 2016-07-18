package domain

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type ListDomains struct {
	ui             terminal.UI
	config         coreconfig.Reader
	domainRepo     api.DomainRepository
	routingAPIRepo api.RoutingAPIRepository
}

func init() {
	commandregistry.Register(&ListDomains{})
}

func (cmd *ListDomains) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "domains",
		Description: T("List domains in the target org"),
		Usage: []string{
			"CF_NAME domains",
		},
	}
}

func (cmd *ListDomains) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	usageReq := requirements.NewUsageRequirement(commandregistry.CLICommandUsagePresenter(cmd),
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

func (cmd *ListDomains) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.domainRepo = deps.RepoLocator.GetDomainRepository()
	cmd.routingAPIRepo = deps.RepoLocator.GetRoutingAPIRepository()

	return cmd
}

func (cmd *ListDomains) Execute(c flags.FlagContext) error {
	org := cmd.config.OrganizationFields()

	cmd.ui.Say(T("Getting domains in org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(org.Name),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	domains, err := cmd.getDomains(org.GUID)
	if err != nil {
		return errors.New(T("Failed fetching domains.\n{{.Error}}", map[string]interface{}{"Error": err.Error()}))
	}

	table := cmd.ui.Table([]string{T("name"), T("status"), T("type")})

	for _, domain := range domains {
		if domain.Shared {
			table.Add(domain.Name, T("shared"), domain.RouterGroupType)
		}
	}

	for _, domain := range domains {
		if !domain.Shared {
			table.Add(domain.Name, T("owned"), domain.RouterGroupType)
		}
	}

	table.Print()

	if len(domains) == 0 {
		cmd.ui.Say(T("No domains found"))
	}
	return nil
}

func (cmd *ListDomains) getDomains(orgGUID string) ([]models.DomainFields, error) {
	domains := []models.DomainFields{}
	err := cmd.domainRepo.ListDomainsForOrg(orgGUID, func(domain models.DomainFields) bool {
		domains = append(domains, domain)
		return true
	})

	if err != nil {
		return []models.DomainFields{}, err
	}

	return domains, nil
}
