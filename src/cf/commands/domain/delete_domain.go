package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteDomain struct {
	ui         terminal.UI
	config     *configuration.Configuration
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewDeleteDomain(ui terminal.UI, config *configuration.Configuration, repo api.DomainRepository) (cmd *DeleteDomain) {
	cmd = new(DeleteDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = repo
	return
}

func (cmd *DeleteDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-domain")
		return
	}

	loginReq := reqFactory.NewLoginRequirement()
	cmd.orgReq = reqFactory.NewTargetedOrgRequirement()

	reqs = []requirements.Requirement{
		loginReq,
		cmd.orgReq,
	}

	return
}

func (cmd *DeleteDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	force := c.Bool("f")

	cmd.ui.Say("Deleting domain %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	domain, apiResponse := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().Guid)
	if apiResponse.IsError() {
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiResponse.Message)
		return
	}
	if apiResponse.IsNotFound() {
		cmd.ui.Ok()
		cmd.ui.Warn(apiResponse.Message)
		return
	}

	if !force {
		answer := cmd.ui.Confirm("Are you sure you want to delete the domain %s and all of its associations?", domainName)

		if !answer {
			return
		}
	}

	apiResponse = cmd.domainRepo.Delete(domain.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
