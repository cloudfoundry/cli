package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type ShareDomain struct {
	ui         terminal.UI
	config     *configuration.Configuration
	domainRepo api.DomainRepository
	orgReq     requirements.OrganizationRequirement
}

func NewShareDomain(ui terminal.UI, config *configuration.Configuration, domainRepo api.DomainRepository) (cmd *ShareDomain) {
	cmd = new(ShareDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = domainRepo
	return
}

func (cmd *ShareDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "share-domain")
		return
	}

	reqs = []requirements.Requirement{
		reqFactory.NewLoginRequirement(),
	}
	return
}

func (cmd *ShareDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]

	cmd.ui.Say("Sharing domain %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiResponse := cmd.domainRepo.CreateSharedDomain(domainName)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed(apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
