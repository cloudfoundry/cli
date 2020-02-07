package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/clock"
)

//go:generate counterfeiter . DeleteSpaceQuotaActor

type DeleteSpaceQuotaActor interface {
	DeleteSpaceQuotaByName(quotaName string, orgGUID string) (v7action.Warnings, error)
}

type DeleteSpaceQuotaCommand struct {
	RequiredArgs    flag.Quota  `positional-args:"yes"`
	Force           bool        `long:"force" short:"f" description:"Force deletion without confirmation"`
	usage           interface{} `usage:"CF_NAME delete-space-quota QUOTA [-f]"`
	relatedCommands interface{} `related_commands:"space-quotas"`

	UI          command.UI
	Config      command.Config
	Actor       DeleteSpaceQuotaActor
	SharedActor command.SharedActor
}

func (cmd *DeleteSpaceQuotaCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd DeleteSpaceQuotaCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	spaceQuotaName := cmd.RequiredArgs.Quota

	if !cmd.Force {
		promptMessage := "Really delete the space quota {{.QuotaName}}?"
		confirmedDelete, promptErr := cmd.UI.DisplayBoolPrompt(false, promptMessage, map[string]interface{}{
			"QuotaName": spaceQuotaName,
		})

		if promptErr != nil {
			return promptErr
		}

		if !confirmedDelete {
			cmd.UI.DisplayText("Space quota '{{.QuotaName}}' has not been deleted.", map[string]interface{}{"QuotaName": spaceQuotaName})
			return nil
		}
	}

	cmd.UI.DisplayTextWithFlavor("Deleting space quota {{.QuotaName}} as {{.User}}...",
		map[string]interface{}{
			"User":      user.Name,
			"QuotaName": spaceQuotaName,
		})

	warnings, err := cmd.Actor.DeleteSpaceQuotaByName(spaceQuotaName, cmd.Config.TargetedOrganization().GUID)

	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		switch err.(type) {
		case actionerror.SpaceQuotaNotFoundByNameError:
			cmd.UI.DisplayWarning(err.Error())
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
