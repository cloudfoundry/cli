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
	org := cmd.orgReq.GetOrganization()

	cmd.ui.Say("Getting domains in org %s...",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	domains, apiResponse := cmd.domainRepo.FindAllByOrg(org)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("")

	table := [][]string{
		[]string{"name", "status", "spaces"},
	}
	for _, domain := range domains {
		var status string
		if domain.Shared {
			status = "shared"
		} else if len(domain.Spaces) == 0 {
			status = "reserved"
		} else {
			status = "owned"
		}

		table = append(table, []string{
			domain.Name,
			status,
			strings.Join(formatters.MapStr(domain.Spaces), ", "),
		})
	}

	cmd.ui.DisplayTable(table)
}
