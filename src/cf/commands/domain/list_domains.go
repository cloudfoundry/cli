package domain

import (
	"cf"
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ListDomains struct {
	ui         terminal.UI
	config     *configuration.Configuration
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewListDomains(ui terminal.UI, config *configuration.Configuration, domainRepo api.DomainRepository) (cmd *ListDomains) {
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
	apiResponse := cmd.domainRepo.ListSharedDomains(domainsCallback("shared", table, &noDomains))
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching shared domains.\n%s", apiResponse.Message)
		return
	}

	apiResponse = cmd.domainRepo.ListDomainsForOrg(org.Guid, domainsCallback("owned", table, &noDomains))
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching private domains.\n%s", apiResponse.Message)
		return
	}

	if noDomains {
		cmd.ui.Say("No domains found")
	}
}

func domainsCallback(status string, table terminal.Table, noDomains *bool) api.ListDomainsCallback {
	return api.ListDomainsCallback(func(domains []cf.Domain) bool {
		rows := [][]string{}
		for _, domain := range domains {
			rows = append(rows, []string{domain.Name, status})
		}
		table.Print(rows)
		*noDomains = false
		return true
	})
}
