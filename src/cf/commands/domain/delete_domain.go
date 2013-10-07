package domain

import (
	"cf/api"
	"cf/requirements"
	"cf/terminal"
	"errors"
	"github.com/codegangsta/cli"
)

type DeleteDomain struct {
	ui         terminal.UI
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewDeleteDomain(ui terminal.UI, repo api.DomainRepository) (cmd *DeleteDomain) {
	cmd = &DeleteDomain{ui: ui, domainRepo: repo}
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

	cmd.ui.Say("Deleting domain %s...", domainName)

	domain, apiResponse := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganization())
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
		var answer bool
		if domain.Shared {
			answer = cmd.ui.Confirm("This domain is shared across all orgs.\nDeleting it will remove all associated routes, and will make any app with this domain unreachable.\nAre you sure you want to delete the domain %s? ", domainName)
		} else {
			answer = cmd.ui.Confirm("Are you sure you want to delete the domain %s and all of its associations?", domainName)
		}

		if !answer {
			return
		}
	}

	apiResponse = cmd.domainRepo.DeleteDomain(domain)
	if apiResponse.IsNotSuccessful() {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiResponse.Message)
		return
	}

	cmd.ui.Ok()
}
