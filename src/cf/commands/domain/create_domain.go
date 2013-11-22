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

type CreateDomain struct {
	ui         terminal.UI
	config     *configuration.Configuration
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewCreateDomain(ui terminal.UI, config *configuration.Configuration, domainRepo api.DomainRepository) (cmd *CreateDomain) {
	cmd = new(CreateDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = domainRepo
	return
}

func (cmd *CreateDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-domain")
		return
	}

	cmd.orgReq = reqFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *CreateDomain) Run(c *cli.Context) {
	domainName := c.Args()[1]
	owningOrg := cmd.orgReq.GetOrganization()

	cmd.ui.Say("Creating domain %s for org %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(owningOrg.Name),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	_, apiResponse := cmd.domainRepo.Create(domainName, owningOrg.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
	cmd.ui.Say("TIP: Use '%s map-domain' to assign it to a space", cf.Name())
}
