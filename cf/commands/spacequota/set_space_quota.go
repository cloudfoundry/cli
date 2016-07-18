package spacequota

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/api/spacequotas"
	"github.com/cloudfoundry/cli/cf/api/spaces"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/flags"
	. "github.com/cloudfoundry/cli/cf/i18n"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/terminal"
)

type SetSpaceQuota struct {
	ui        terminal.UI
	config    coreconfig.Reader
	spaceRepo spaces.SpaceRepository
	quotaRepo spacequotas.SpaceQuotaRepository
}

func init() {
	commandregistry.Register(&SetSpaceQuota{})
}

func (cmd *SetSpaceQuota) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:        "set-space-quota",
		Description: T("Assign a space quota definition to a space"),
		Usage: []string{
			T("CF_NAME set-space-quota SPACE-NAME SPACE-QUOTA-NAME"),
		},
	}
}

func (cmd *SetSpaceQuota) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) []requirements.Requirement {
	if len(fc.Args()) != 2 {
		cmd.ui.Failed(T("Incorrect Usage. Requires SPACE-NAME and SPACE-QUOTA-NAME as arguments\n\n") + commandregistry.Commands.CommandUsage("set-space-quota"))
	}

	reqs := []requirements.Requirement{
		requirementsFactory.NewLoginRequirement(),
		requirementsFactory.NewTargetedOrgRequirement(),
	}

	return reqs
}

func (cmd *SetSpaceQuota) SetDependency(deps commandregistry.Dependency, pluginCall bool) commandregistry.Command {
	cmd.ui = deps.UI
	cmd.config = deps.Config
	cmd.spaceRepo = deps.RepoLocator.GetSpaceRepository()
	cmd.quotaRepo = deps.RepoLocator.GetSpaceQuotaRepository()
	return cmd
}

func (cmd *SetSpaceQuota) Execute(c flags.FlagContext) error {

	spaceName := c.Args()[0]
	quotaName := c.Args()[1]

	cmd.ui.Say(T("Assigning space quota {{.QuotaName}} to space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"QuotaName": terminal.EntityNameColor(quotaName),
		"SpaceName": terminal.EntityNameColor(spaceName),
		"Username":  terminal.EntityNameColor(cmd.config.Username()),
	}))

	space, err := cmd.spaceRepo.FindByName(spaceName)
	if err != nil {
		return err
	}

	if space.SpaceQuotaGUID != "" {
		return errors.New(T("This space already has an assigned space quota."))
	}

	quota, err := cmd.quotaRepo.FindByName(quotaName)
	if err != nil {
		return err
	}

	err = cmd.quotaRepo.AssociateSpaceWithQuota(space.GUID, quota.GUID)
	if err != nil {
		return err
	}

	cmd.ui.Ok()
	return nil
}
