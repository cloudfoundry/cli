package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/formatters"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
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

	stopChan := make(chan bool)
	defer close(stopChan)

	domainsChan, statusChan := cmd.domainRepo.ListDomainsForOrg(org.Guid, stopChan)

	table := cmd.ui.Table([]string{"name", "status", "spaces"})
	noDomains := true

	for domains := range domainsChan {
		rows := [][]string{}
		for _, domain := range domains {

			var status string
			if domain.Shared {
				status = "shared"
			} else if len(domain.Spaces) == 0 {
				status = "reserved"
			} else {
				status = "owned"
			}

			rows = append(rows, []string{
				domain.Name,
				status,
				strings.Join(formatters.MapStr(domain.Spaces), ", "),
			})
		}
		table.Print(rows)
		noDomains = false
	}

	apiStatus := <-statusChan
	if apiStatus.IsNotSuccessful() {
		cmd.ui.Failed("Failed fetching domains.\n%s", apiStatus.Message)
		return
	}

	if noDomains {
		cmd.ui.Say("No domains found")
	}
}
