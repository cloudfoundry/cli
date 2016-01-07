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

type SetSpaceQuota struct {
	ui        terminal.UI
	config    core_config.Reader
	spaceRepo spaces.SpaceRepository
	quotaRepo space_quotas.SpaceQuotaRepository
}

func init() {
	command_registry.Register(&SetSpaceQuota{})
}

func (cmd *SetSpaceQuota) MetaData() command_registry.CommandMetadata {
	return command_registry.CommandMetadata{
		Name:        "set-space-quota",
		Description: T("Assign a space quota definition to a space"),
		Usage:       T("CF_NAME set-space-quota SPACE-NAME SPACE-QUOTA-NAME"),
	}
}

func (cmd *SetSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) (reqs []requirements.Requirement, err error) {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE-NAME and SPACE-QUOTA-NAME as arguments\n\n") + command_registry.Commands.CommandUsage("set-space-quota"))
	}

	return []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}, nil
}

func (cmd *SetSpaceQuota) SetDependency(deps command_registry.Dependency, pluginCall bool) command_registry.Command {
	cmd.ui = deps.Ui
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *SetSpaceQuota) Execute(c flags.FlagContext) {

	spaceName := c.Args()[0]
	quotaName := c.Args()[1]

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
