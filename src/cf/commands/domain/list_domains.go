package domain

import (
	"cf/api"
	"cf/commands/application"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
	"strings"
)

type ListDomains struct {
	ui         terminal.UI
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewListDomains(ui terminal.UI, domainRepo api.DomainRepository) (cmd *ListDomains) {
	cmd = new(ListDomains)
	cmd.ui = ui
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

	cmd.ui.Say("Getting domains in org %s...", org.Name)

	domains, apiResponse := cmd.domainRepo.FindAllByOrg(org)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

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
			strings.Join(application.MapStr(domain.Spaces), ", "),
		})
	}

	cmd.ui.Ok()
	cmd.ui.DisplayTable(table)
}
