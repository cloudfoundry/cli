package domain

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
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

func (command *DeleteDomain) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-domain",
		Description: "Delete a domain",
		Usage:       "CF_NAME delete-domain DOMAIN [-f]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
}

func (cmd *DeleteDomain) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-domain")
		return
	}

	loginReq := requirementsFactory.NewLoginRequirement()
	cmd.orgReq = requirementsFactory.NewTargetedOrgRequirement()

	reqs = []requirements.Requirement{
		loginReq,
		cmd.orgReq,
	}

	return
}

func (cmd *DeleteDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	domain, apiErr := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().Guid)

	switch apiErr.(type) {
	case nil: //do nothing
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(apiErr.Error())
		return
	default:
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiErr.Error())
		return
	}

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDelete("domain", domainName) {
			return
		}
	}

	cmd.ui.Say("Deleting domain %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr = cmd.domainRepo.Delete(domain.Guid)
	if apiErr != nil {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
