package spacequota

import (
	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
	"github.com/cloudfoundry/cli/flags"
)

type UnsetSpaceQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	quotaRepo space_quotas.SpaceQuotaRepository
	spaceRepo spaces.SpaceRepository
}

func init() {
	command_registry.Register(&UnsetSpaceQuota{})
}

func (cmd *UnsetSpaceQuota) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "unset-space-quota",
		Description: T("Unassign a quota from a space"),
		Usage:       T("CF_NAME unset-space-quota SPACE QUOTA\n\n"),
	}
}

func (cmd *UnsetSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE and QUOTA as arguments\n\n") + command_registry.Commands.CommandUsage("unset-space-quota"))
	}

	reqs = []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}
	return
}

func (cmd *UnsetSpaceQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *UnsetSpaceQuota) Execute(c flags.FlagContext) {
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
