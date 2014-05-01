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

func (command *DeleteSharedDomain) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-shared-domain",
		Description: "Delete a shared domain",
		Usage:       "CF_NAME delete-shared-domain DOMAIN [-f]",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: "Force deletion without confirmation"},
		},
	}
}

func (cmd *DeleteSharedDomain) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-shared-domain")
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

func (cmd *DeleteSharedDomain) Run(c *cli.Context) {
	domainName := c.Args()[0]
	force := c.Bool("f")

	cmd.ui.Say("Deleting domain %s as %s...",
		terminal.EntityNameColor(domainName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	domain, apiErr := cmd.domainRepo.FindByNameInOrg(domainName, cmd.orgReq.GetOrganizationFields().Guid)
	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(apiErr.Error())
		return
	default:
		cmd.ui.Failed("Error finding domain %s\n%s", domainName, apiErr.Error())
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

	apiErr = cmd.domainRepo.DeleteSharedDomain(domain.Guid)
	if apiErr != nil {
		cmd.ui.Failed("Error deleting domain %s\n%s", domainName, apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
