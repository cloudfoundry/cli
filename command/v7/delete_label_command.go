package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
)

//go:generate counterfeiter . DeleteLabelActor

type DeleteLabelActor interface {
	UpdateApplicationLabelsByApplicationName(string, string, map[string]types.NullString) (v7action.Warnings, error)
}

type DeleteLabelCommand struct {
	RequiredArgs flag.DeleteLabelArgs `positional-args:"yes"`
	usage        interface{}          `usage:"CF_NAME delete-label RESOURCE RESOURCE_NAME KEY\n\n EXAMPLES:\n   cf delete-label app dora ci_signature_sha2\n\nRESOURCES:\n   APP\n\nSEE ALSO:\n   set-label, labels"`
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
	Actor        DeleteLabelActor
}

func (cmd *DeleteLabelCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, _, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)
	return nil
}

func (cmd DeleteLabelCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Deleting label(s) for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.ResourceName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})

	labels := make(map[string]types.NullString)
	for _, value := range cmd.RequiredArgs.LabelKeys {
		labels[value] = types.NewNullString()
	}

	warnings, err := cmd.Actor.UpdateApplicationLabelsByApplicationName(cmd.RequiredArgs.ResourceName, cmd.Config.TargetedSpace().GUID, labels)

	for _, warning := range warnings {
		cmd.UI.DisplayWarning(warning)
	}

	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
