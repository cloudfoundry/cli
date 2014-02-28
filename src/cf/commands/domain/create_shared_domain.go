package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type CreateSharedDomain struct {
	ui         terminal.UI
	config     configuration.Reader
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewCreateSharedDomain(ui terminal.UI, config configuration.Reader, domainRepo api.DomainRepository) (cmd *CreateSharedDomain) {
	cmd = new(CreateSharedDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = domainRepo
	return
}

func (cmd *CreateSharedDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "create-shared-domain")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *CreateSharedDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]

	cmd.ui.Say("Creating shared domain %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.domainRepo.CreateSharedDomain(domainName)
	if apiResponse != nil {
		cmd.ui.Failed(apiResponse.Error())
		return
	}

	cmd.ui.Ok()
}
