package domain

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateDomain struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewCreateDomain(ui terminal.UI, domainRepo api.DomainRepository) (cmd *CreateDomain) {
	cmd = new(CreateDomain)
	cmd.ui = ui
	cmd.domainRepo = domainRepo
	return
}

func (cmd *CreateDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-domain")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[1])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *CreateDomain) Run(c *cli.Context) {
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
