package domain

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ParkDomain struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewParkDomain(ui terminal.UI, domainRepo api.DomainRepository) (cmd *ParkDomain) {
	cmd = new(ParkDomain)
	cmd.ui = ui
	cmd.domainRepo = domainRepo
	return
}

func (cmd *ParkDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "park-domain")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[1])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *ParkDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say("Creating domain %s on %s...", domainName, owningOrg.Name)

	domain := cf.Domain{Name: domainName}

	_, apiStatus := cmd.domainRepo.Create(domain, owningOrg)
	if apiStatus.IsError() {
		cmd.ui.Failed(apiStatus.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s map-domain' to assign it to a space.", cf.Name)
}
