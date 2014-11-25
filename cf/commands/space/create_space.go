package space

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/organizations"
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/commands/user"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flag_helpers"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type CreateSpace struct {
	ui              terminal.UI
	config          core_config.Reader
	spaceRepo       spaces.SpaceRepository
	orgRepo         organizations.OrganizationRepository
	userRepo        api.UserRepository
	spaceRoleSetter user.SpaceRoleSetter
	spaceQuotaRepo  space_quotas.SpaceQuotaRepository
}

func NewCreateSpace(ui terminal.UI, config core_config.Reader, spaceRoleSetter user.SpaceRoleSetter, spaceRepo spaces.SpaceRepository, orgRepo organizations.OrganizationRepository, userRepo api.UserRepository, spaceQuotaRepo space_quotas.SpaceQuotaRepository) (cmd CreateSpace) {
	cmd.ui = ui
	cmd.config = config
	cmd.spaceRoleSetter = spaceRoleSetter
	cmd.spaceRepo = spaceRepo
	cmd.orgRepo = orgRepo
	cmd.userRepo = userRepo
	cmd.spaceQuotaRepo = spaceQuotaRepo
	return
}

func (cmd CreateSpace) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "create-space",
		Description: T("Create a space"),
		Usage:       T("CF_NAME create-space SPACE [-o ORG] [-q SPACE_QUOTA]"),
		Flags: []cli.Flag{
			flag_helpers.NewStringFlag("o", T("Organization")),
			flag_helpers.NewStringFlag("q", T("Space-Quota to assign to the newly created space. Excluding this option results in assignment of default quota")),
		},
	}
}

func (cmd CreateSpace) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 1 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{requirementsFactory.NewLoginRequirement()}
	if c.String("o") == "" {
		reqs = append(reqs, requirementsFactory.NewTargetedOrgRequirement())
	}

	return
}

func (cmd CreateSpace) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	orgName := c.String("o")
	orgGuid := ""
	if orgName == "" {
		orgName = cmd.config.OrganizationFields().Name
		orgGuid = cmd.config.OrganizationFields().Guid
	}

	cmd.ui.Say(T("Creating space {{.SpaceName}} in org {{.OrgName}} as {{.CurrentUser}}...",
		map[string]interface{}{
			"SpaceName":   terminal.EntityNameColor(spaceName),
			"OrgName":     terminal.EntityNameColor(orgName),
			"CurrentUser": terminal.EntityNameColor(cmd.config.Username()),
		}))

	if orgGuid == "" {
		org, apiErr := cmd.orgRepo.FindByName(orgName)
		switch apiErr.(type) {
		case nil:
		case *errors.ModelNotFoundError:
			cmd.ui.Failed(T("Org {{.OrgName}} does not exist or is not accessible", map[string]interface{}{"OrgName": orgName}))
			return
		default:
			cmd.ui.Failed(T("Error finding org {{.OrgName}}\n{{.ErrorDescription}}",
				map[string]interface{}{
					"OrgName":          orgName,
					"ErrorDescription": apiErr.Error(),
				}))
			return
		}

		orgGuid = org.Guid
	}

	spaceQuotaGuid := ""
	spaceQuotaName := c.String("q")
	if spaceQuotaName != "" {
		spaceQuota, err := cmd.spaceQuotaRepo.FindByName(spaceQuotaName)
		if err != nil {
			cmd.ui.Failed(err.Error())
			return
		}
		spaceQuotaGuid = spaceQuota.Guid
	}

	space, err := cmd.spaceRepo.Create(spaceName, orgGuid, spaceQuotaGuid)
	if err != nil {
		if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.SPACE_EXISTS {
			cmd.ui.Ok()
			cmd.ui.Warn(T("Space {{.SpaceName}} already exists", map[string]interface{}{"SpaceName": spaceName}))
			return
		}
		cmd.ui.Failed(err.Error())
		return
	}
	cmd.ui.Ok()

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.SPACE_MANAGER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	err = cmd.spaceRoleSetter.SetSpaceRole(space, models.SPACE_DEVELOPER, cmd.config.UserGuid(), cmd.config.Username())
	if err != nil {
		cmd.ui.Failed(err.Error())
		return
	}

	if spaceQuotaName != "" {
		cmd.ui.Say(T("Assigning space quota {{.QuotaName}} to space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
			"QuotaName": terminal.EntityNameColor(spaceQuotaName),
			"SpaceName": terminal.EntityNameColor(spaceName),
			"Username":  terminal.EntityNameColor(cmd.config.Username()),
		}))
	}

	cmd.ui.Say(T("\nTIP: Use '{{.CFTargetCommand}}' to target new space",
		map[string]interface{}{
			"CFTargetCommand": terminal.CommandColor(cf.Name() + " target -o " + orgName + " -s " + space.Name),
		}))
}
