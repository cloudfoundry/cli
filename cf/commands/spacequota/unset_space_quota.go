package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type UnsetSpaceQuota struct {
	ui        terminal.UI
	config    configuration.Reader
	quotaRepo space_quotas.SpaceQuotaRepository
	spaceRepo spaces.SpaceRepository
}

func NewUnsetSpaceQuota(ui terminal.UI, config configuration.Reader, quotaRepo space_quotas.SpaceQuotaRepository, spaceRepo spaces.SpaceRepository) UnsetSpaceQuota {
	return UnsetSpaceQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		spaceRepo: spaceRepo,
	}
}

func (cmd UnsetSpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "unset-space-quota",
		Description: T("Unassign a quota from a space"),
		Usage:       T("CF_NAME unset-space-quota SPACE QUOTA\n\n"),
	}
}

func (cmd UnsetSpaceQuota) GetRequirements(requirementsFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error) {
	if len(c.Args()) != 2 {
		cmd.ui.FailWithUsage(c)
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd UnsetSpaceQuota) Run(c *cli.Context) {
	spaceName := c.Args()[0]
	quotaName := c.Args()[1]

	space, apiErr := cmd.spaceRepo.FindByName(spaceName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	quota, apiErr := cmd.quotaRepo.FindByName(quotaName)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Say(T("Unassigning space quota {{.QuotaName}} from space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": terminal.EntityNameColor(quota.Name),
			"SpaceName": terminal.EntityNameColor(space.Name),
			"Username":  terminal.EntityNameColor(cmd.config.Username())}))

	apiErr = cmd.quotaRepo.UnassignQuotaFromSpace(space.Guid, quota.Guid)
	if apiErr != nil {
		cmd.ui.Failed(apiErr.Error())
		return
	}

	cmd.ui.Ok()
}
