package domain

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteDomain struct {
	ui         terminal.UI
	config     configuration.Reader
	orgReq     requirements.TargetedOrgRequirement
	domainRepo api.DomainRepository
}

func NewDeleteDomain(ui terminal.UI, config configuration.Reader, repo api.DomainRepository) (cmd *DeleteDomain) {
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

	domain, apiErr := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().Guid)

	switch apiErr.(type) {
	case nil: //do nothing
	case errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(apiErr.Error())
		return
	default:
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiErr.Error())
		return
	}

	if !force {
		answer := cmd.ui.Confirm("Are you sure you want to delete the domain %s and all of its associations?", domainName)

		if !answer {
			return
		}
	}

	apiErr = cmd.domainRepo.Delete(domain.Guid)
	if apiErr != nil {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
