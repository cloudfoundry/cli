package commands

import (
	"cf"
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateSharedDomain struct {
	ui         terminal.UI
	domainRepo api.DomainRepository
}

func NewCreateSharedDomain(ui terminal.UI, domainRepo api.DomainRepository) (cmd CreateSharedDomain) {
	cmd.ui = ui
	cmd.domainRepo = domainRepo
	return
}

func (csd CreateSharedDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	println(len(c.Args()))
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		csd.ui.FailWithUsage(c, "create-domain")
		return
	}
	return
}

func (csd CreateSharedDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	domain := cf.Domain{Name: domainName}

	csd.ui.Say("Creating shared domain %s", domainName)

	_, apiErr := csd.domainRepo.Create(domain)

	if apiErr != nil {
		csd.ui.Failed(apiErr.Error())
		return
	}

	csd.ui.Ok()
}
