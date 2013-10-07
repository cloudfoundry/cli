package domain

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ReserveDomain struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewReserveDomain(ui terminal.UI, domainRepo api.DomainRepository) (cmd *ReserveDomain) {
	cmd = new(ReserveDomain)
	cmd.ui = ui
	cmd.domainRepo = domainRepo
	return
}

func (cmd *ReserveDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "reserve-domain")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *ReserveDomain) Run(c *cli.Context) {
	domainName := c.Args()[1]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say("Reserving domain %s for org %s...", domainName, owningOrg.Name)

	domain := cf.Domain{Name: domainName}

	_, apiResponse := cmd.domainRepo.Create(domain, owningOrg)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s map-domain' to assign it to a space", cf.Name)
}
