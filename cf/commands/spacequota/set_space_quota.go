package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/codegangsta/cli"
)

type SetSpaceQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	spaceRepo spaces.SpaceRepository
	quotaRepo space_quotas.SpaceQuotaRepository
}

func NewSetSpaceQuota(ui terminal.UI, config core_config.Reader, spaceRepo spaces.SpaceRepository, quotaRepo space_quotas.SpaceQuotaRepository) SetSpaceQuota {
	return SetSpaceQuota{
		ui:        ui,
		config:    config,
		quotaRepo: quotaRepo,
		spaceRepo: spaceRepo,
	}
}

func (cmd SetSpaceQuota) Metadata() command_metadata.CommandMetadata {
	return command_metadata.CommandMetadata{
		Name:        "set-space-quota",
		Description: T("Assign a space quota definition to a space"),
		Usage:       T("CF_NAME set-space-quota SPACE-NAME SPACE-QUOTA-NAME"),
	}
}

func (cmd SetSpaceQuota) GetRequirements(requirementsFactory requirements.Factory, context *cli.Context) ([]requirements.Requirement, error) {
	if len(context.Args()) != 2 {
		cmd.ui.FailWithUsage(context)
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}, nil
}

func (cmd SetSpaceQuota) Run(context *cli.Context) {

	spaceName := context.Args()[0]
	quotaName := context.Args()[1]

	cmd.ui.Say(T("Assigning space quota {{.QuotaName}} to space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(quotaName),
		"SpaceName": terminal.EntityNameColor(spaceName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	space, err := cmd.spaceRepo.FindByName(spaceName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	if space.SpaceQuotaGuid != "" {
		cmd.ui.Failed(T("This space already has an assigned space quota."))
	}

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	err = cmd.quotaRepo.AssociateSpaceWithQuota(space.Guid, quota.Guid)
	if err != nil {
		cmd.ui.Failed(err.Error())
	}

	cmd.ui.Ok()
}
