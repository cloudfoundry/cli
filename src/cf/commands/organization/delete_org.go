package organization

import (
	"cf/api"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/requirements"
	"cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteOrg struct {
	ui      terminal.UI
	config  configuration.ReadWriter
	orgRepo api.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func NewDeleteOrg(ui terminal.UI, config configuration.ReadWriter, sR api.OrganizationRepository) (cmd *DeleteOrg) {
	cmd = new(DeleteOrg)
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = sR
	return
}

func (cmd *DeleteOrg) GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "delete-org")
		return

	}
	return
}

func (cmd *DeleteOrg) Run(c *cli.Context) {
	orgName := c.Args()[0]

	force := c.Bool("f")

	if !force {
		response := cmd.ui.Confirm(
			"Really delete org %s and everything associated with it?%s",
			terminal.EntityNameColor(orgName),
			terminal.PromptColor(">"),
		)

		if !response {
			return
		}
	}

	cmd.ui.Say("Deleting org %s as %s...",
		terminal.EntityNameColor(orgName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	org, apiErr := cmd.orgRepo.FindByName(orgName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn("Org %s does not exist.", orgName)
		return
	default:
		cmd.ui.Failed(apiErr.Error())
		return
	}

	apiErr = cmd.orgRepo.Delete(org.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	if org.Guid == cmd.config.OrganizationFields().Guid {
		cmd.config.SetOrganizationFields(models.OrganizationFields{})
		cmd.config.SetSpaceFields(models.SpaceFields{})
	}

	cmd.ui.Ok()
	return
}
