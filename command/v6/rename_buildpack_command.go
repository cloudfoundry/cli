package v6

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v6/shared"
)

//go:generate counterfeiter . RenameBuildpackActor

type RenameBuildpackActor interface {
	CloudControllerAPIVersion() string
	RenameBuildpack(oldName string, newName string, stackName string) (v2action.Warnings, error)
}

type RenameBuildpackCommand struct {
	RequiredArgs    flag.RenameBuildpackArgs `positional-args:"yes"`
	Stack           string                   `short:"s" description:"Specify which buildpack to rename by stack"`
	usage           interface{}              `usage:"CF_NAME rename-buildpack BUILDPACK_NAME NEW_BUILDPACK_NAME"`
	relatedCommands interface{}              `related_commands:"update-buildpack"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       RenameBuildpackActor
}

func (cmd *RenameBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)

	return nil
}

func (cmd RenameBuildpackCommand) Execute(args []string) error {
	if err := cmd.SharedActor.CheckTarget(false, false); err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	if cmd.stackSpecified() {
		err = command.MinimumCCAPIVersionCheck(cmd.Actor.CloudControllerAPIVersion(), ccversion.MinVersionBuildpackStackAssociationV2, "Option '-s'")
		if err != nil {
			return err
		}
		cmd.UI.DisplayTextWithFlavor("Renaming buildpack {{.OldName}} to {{.NewName}} with stack {{.Stack}} as {{.CurrentUser}}...", map[string]interface{}{
			"OldName":     cmd.RequiredArgs.OldBuildpackName,
			"NewName":     cmd.RequiredArgs.NewBuildpackName,
			"Stack":       cmd.Stack,
			"CurrentUser": user.Name,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Renaming buildpack {{.OldName}} to {{.NewName}} as {{.CurrentUser}}...", map[string]interface{}{
			"OldName":     cmd.RequiredArgs.OldBuildpackName,
			"NewName":     cmd.RequiredArgs.NewBuildpackName,
			"CurrentUser": user.Name,
		})
	}

	warnings, err := cmd.Actor.RenameBuildpack(cmd.RequiredArgs.OldBuildpackName, cmd.RequiredArgs.NewBuildpackName, cmd.Stack)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd RenameBuildpackCommand) stackSpecified() bool {
	return len(cmd.Stack) > 0
}
