package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
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

func (cmd *ListDomains) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) > 0 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "domains")
		return
	}

	cmd.orgReq = reqFactory.NewTargetedOrgRequirement()
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *ListDomains) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganizationFields()

	cmd.ui.Say("Getting domains in org %s as %s...",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	noDomains := true
	table := cmd.ui.Table([]string{"name                              ", "status"})

	apiErr := cmd.domainRepo.ListSharedDomains(domainsCallback(table, &noDomains))

	switch apiErr.(type) {
	case nil:
	case *errors.HttpNotFoundError:
	default:
		cmd.ui.Failed("Failed fetching shared domains.\n%s", apiErr.Error())
		return
	}

	apiErr = cmd.domainRepo.ListDomainsForOrg(org.Guid, domainsCallback(table, &noDomains))
	if apiErr != nil {
		cmd.ui.Failed("Failed fetching private domains.\n%s", apiErr.Error())
		return
	}

	if noDomains {
		cmd.ui.Say("No domains found")
	}
}

func domainsCallback(table terminal.Table, noDomains *bool) func(models.DomainFields) bool {
	return func(domain models.DomainFields) bool {
		table.Print([][]string{{domain.Name, domainStatusString(domain)}})
		*noDomains = false
		return true
	}
}

func domainStatusString(domain models.DomainFields) string {
	if domain.Shared {
		return "shared"
	} else {
		return "owned"
	}
}
