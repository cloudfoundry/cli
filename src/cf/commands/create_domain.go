package commands

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
	orgRepo    api.OrganizationRepository
}

func NewCreateDomain(ui terminal.UI, domainRepo api.DomainRepository, orgRepo api.OrganizationRepository) (cmd CreateDomain) {
	cmd.ui = ui
	cmd.domainRepo = domainRepo
	cmd.orgRepo = orgRepo
	return
}

func (cmd CreateDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-domain")
		return
	}
	return
}

func (cmd CreateDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	orgName := c.Args()[1]

	cmd.ui.Say("Creating domain %s on %s...", domainName, orgName)

	domain := cf.Domain{Name: domainName}

	owningOrg, found, apiErr := cmd.orgRepo.FindByName(orgName)
	if !found {
		cmd.ui.Failed("Org %s could not be found.", orgName)
		return
	}

	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	_, apiErr = cmd.domainRepo.Create(domain, owningOrg)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s map-domain' to assign it to a space.", cf.Name)
}
