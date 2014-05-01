package organization

import (
	"errors"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type RenameOrg struct {
	ui      terminal.UI
	config  configuration.ReadWriter
	orgRepo api.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func NewRenameOrg(ui terminal.UI, config configuration.ReadWriter, orgRepo api.OrganizationRepository) (cmd *RenameOrg) {
	cmd = new(RenameOrg)
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (command *RenameOrg) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "rename-org",
		Description: "Rename an org",
		Usage:       "CF_NAME rename-org ORG NEW_ORG",
	}
}

func (cmd *RenameOrg) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		err = errors.New("Incorrect Usage")
		cmd.ui.FailWithUsage(c, "rename-org")
		return
	}
	cmd.orgReq = requirementsFactory.NewOrganizationRequirement(c.Args()[0])
	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		cmd.orgReq,
	}
	return
}

func (cmd *RenameOrg) Run(c *cli.Context) {
	org := cmd.orgReq.GetOrganization()
	newName := c.Args()[1]

	cmd.ui.Say("Renaming org %s to %s as %s...",
		terminal.EntityNameColor(org.Name),
		terminal.EntityNameColor(newName),
		terminal.EntityNameColor(cmd.config.Username()),
	)

	apiErr := cmd.orgRepo.Rename(org.Guid, newName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}
	cmd.ui.Ok()

	if org.Guid == cmd.config.OrganizationFields().Guid {
		org.Name = newName
		cmd.config.SetOrganizationFields(org.OrganizationFields)
	}
}
