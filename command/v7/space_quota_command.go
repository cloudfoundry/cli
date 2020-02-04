package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SpaceQuotaActor

type SpaceQuotaActor interface {
	GetSpaceQuotaByName(spaceQuotaName string, orgGUID string) (v7action.SpaceQuota, v7action.Warnings, error)
}

type SpaceQuotaCommand struct {
	RequiredArgs    flag.SpaceQuota `positional-args:"yes"`
	usage           interface{}     `usage:"CF_NAME space-quota QUOTA"`
	relatedCommands interface{}     `related_commands:"space, space-quotas"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SpaceQuotaActor
}

func (cmd *SpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd SpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	quotaName := cmd.RequiredArgs.SpaceQuota

	cmd.UI.DisplayTextWithFlavor(
		"Getting space quota {{.QuotaName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": quotaName,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	spaceQuota, warnings, err := cmd.Actor.GetSpaceQuotaByName(quotaName, cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplaySingleQuota(v7action.Quota(spaceQuota.Quota))

	return nil
}
