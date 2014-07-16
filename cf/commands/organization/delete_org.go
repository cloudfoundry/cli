package organization

import (
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type DeleteOrg struct {
	ui      terminal.UI
	config  configuration.ReadWriter
	orgRepo api.OrganizationRepository
	orgReq  requirements.OrganizationRequirement
}

func NewDeleteOrg(ui terminal.UI, config configuration.ReadWriter, orgRepo api.OrganizationRepository) (cmd *DeleteOrg) {
	cmd = new(DeleteOrg)
	cmd.ui = ui
	cmd.config = config
	cmd.orgRepo = orgRepo
	return
}

func (cmd *DeleteOrg) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "delete-org",
		Description: T("Delete an org"),
		Usage:       T("CF_NAME delete-org ORG [-f]"),
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "f", Usage: T("Force deletion without confirmation")},
		},
	}
}

func (cmd *DeleteOrg) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	return
}

func (cmd *DeleteOrg) Run(c *cli.Context) {
	orgName := c.Args()[0]

	if !c.Bool("f") {
		if !cmd.ui.ConfirmDeleteWithAssociations(T("org"), orgName) {
			return
		}
	}

	cmd.ui.Say(T("Deleting org {{.OrgName}} as {{.Username}}...",
		map[string]interface{}{
			"OrgName":  terminal.EntityNameColor(orgName),
			"Username": terminal.EntityNameColor(cmd.config.Username())}))

	org, apiErr := cmd.orgRepo.FindByName(orgName)

	switch apiErr.(type) {
	case nil:
	case *errors.ModelNotFoundError:
		cmd.ui.Ok()
		cmd.ui.Warn(T("Org {{.OrgName}} does not exist.",
			map[string]interface{}{"OrgName": orgName}))
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
