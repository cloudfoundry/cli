package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteSharedDomain struct {
	ui         terminal.UI
	config     configuration.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewDeleteSharedDomain(ui terminal.UI, config configuration.Reader, repo api.DomainRepository) (cmd *DeleteSharedDomain) {
	cmd = new(DeleteSharedDomain)
	cmd.ui = ui
	cmd.config = config
	cmd.domainRepo = repo
	return
}

func (cmd *DeleteSharedDomain) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-shared-domain")
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

func (cmd *DeleteSharedDomain) Run(c *cli.Context) {
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
		answer := cmd.ui.Confirm(
			`This domain is shared across all orgs.
Deleting it will remove all associated routes, and will make any app with this domain unreachable.
Are you sure you want to delete the domain %s? `, domainName)

		if !answer {
			return
		}
	}

	apiResponse = cmd.domainRepo.DeleteSharedDomain(domain.Guid)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
