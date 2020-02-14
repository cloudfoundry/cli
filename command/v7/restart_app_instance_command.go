package v7

import (
	"code.cloudfoundry.org/cli/command/flag"
)

type RestartAppInstanceCommand struct {
	BaseCommand

	RequiredArgs    flag.AppInstance `positional-args:"yes"`
	ProcessType     string           `long:"process" default:"web" description:"Process to restart"`
	usage           interface{}      `usage:"CF_NAME restart-app-instance APP_NAME INDEX [--process PROCESS]"`
	relatedCommands interface{}      `related_commands:"restart"`
}

func (cmd RestartAppInstanceCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Restarting instance {{.InstanceIndex}} of process {{.ProcessType}} of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"InstanceIndex": cmd.RequiredArgs.Index,
		"ProcessType":   cmd.ProcessType,
		"AppName":       cmd.RequiredArgs.AppName,
		"Username":      user.Name,
		"OrgName":       cmd.Config.TargetedOrganization().Name,
		"SpaceName":     cmd.Config.TargetedSpace().Name,
	})

	warnings, err := cmd.Actor.DeleteInstanceByApplicationNameSpaceProcessTypeAndIndex(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, cmd.ProcessType, cmd.RequiredArgs.Index)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}
