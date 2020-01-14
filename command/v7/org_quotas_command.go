package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . OrgQuotasActor

type OrgQuotasActor interface {
	GetOrganizationQuotas() ([]v7action.OrganizationQuota, v7action.Warnings, error)
}

type OrgQuotasCommand struct {
	usage           interface{} `usage:"CF_NAME org-quotas"`
	relatedCommands interface{} `related_commands:"org-quota"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       OrgQuotasActor
}

func (cmd *OrgQuotasCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd OrgQuotasCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting org quotas as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	quotas, warnings, err := cmd.Actor.GetOrganizationQuotas()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.displayTable(quotas)

	return nil
}

func (cmd OrgQuotasCommand) displayTable(orgQuotas []v7action.OrganizationQuota) {
	var keyValueTable = [][]string{
		{"name", "total memory", "instance memory", "routes", "service instances", "paid service plans", "app instances", "route ports"},
	}

	for _, orgQuota := range orgQuotas {
		paidServicesOutput := "disallowed"
		if orgQuota.Services.PaidServicePlans {
			paidServicesOutput = "allowed"
		}

		keyValueTable = append(keyValueTable, []string{
			orgQuota.Name,
			cmd.presentNullOrIntValue(orgQuota.Apps.TotalMemory),
			cmd.presentNullOrIntValue(orgQuota.Apps.InstanceMemory),
			cmd.presentNullOrIntValue(orgQuota.Routes.TotalRoutes),
			cmd.presentNullOrIntValue(orgQuota.Services.TotalServiceInstances),
			paidServicesOutput,
			cmd.presentNullOrIntValue(orgQuota.Apps.TotalAppInstances),
			cmd.presentNullOrIntValue(orgQuota.Routes.TotalRoutePorts),
		})
	}

	cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}

func (cmd OrgQuotasCommand) presentNullOrIntValue(limit types.NullInt) string {
	if !limit.IsSet {
		return "unlimited"
	} else {
		return strconv.Itoa(limit.Value)
	}
}
