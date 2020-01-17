package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . OrgQuotaActor

type OrgQuotaActor interface {
	GetOrganizationQuotaByName(orgQuotaName string) (v7action.OrganizationQuota, v7action.Warnings, error)
}

type OrgQuotaCommand struct {
	RequiredArgs    flag.OrganizationQuota `positional-args:"yes"`
	usage           interface{}            `usage:"CF_NAME quota QUOTA"`
	relatedCommands interface{}            `related_commands:"org, quotas"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       OrgQuotaActor
}

func (cmd *OrgQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd OrgQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	quotaName := cmd.RequiredArgs.OrganizationQuota

	cmd.UI.DisplayTextWithFlavor(
		"Getting org quota {{.QuotaName}} as {{.Username}}...",
		map[string]interface{}{
			"QuotaName": quotaName,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	orgQuota, warnings, err := cmd.Actor.GetOrganizationQuotaByName(quotaName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplaySingleQuota(orgQuota)

	return nil
}
