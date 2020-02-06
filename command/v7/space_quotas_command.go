package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . SpaceQuotasActor

type SpaceQuotasActor interface {
	GetSpaceQuotasByOrgGUID(orgGUID string) ([]v7action.SpaceQuota, v7action.Warnings, error)
}

type SpaceQuotasCommand struct {
	usage           interface{} `usage:"CF_NAME space-quotas"`
	relatedCommands interface{} `related_commands:"space-quota, set-space-quota"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       SpaceQuotasActor
}

func (cmd *SpaceQuotasCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient, clock.NewClock())

	return nil
}

func (cmd SpaceQuotasCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting space quotas as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	orgQuotas, warnings, err := cmd.Actor.GetSpaceQuotasByOrgGUID(cmd.Config.TargetedOrganization().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	var quotas []v7action.Quota
	for _, orgQuota := range orgQuotas {
		quotas = append(quotas, v7action.Quota(orgQuota.Quota))
	}

	quotaDisplayer := shared.NewQuotaDisplayer(cmd.UI)
	quotaDisplayer.DisplayQuotasTable(quotas, "No space quotas found.")

	return nil
}
