package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . DeleteBuildpackActor

type DeleteBuildpackActor interface {
	DeleteBuildpackByNameAndStack(buildpackName string, buildpackStack string) (v7action.Warnings, error)
}

type DeleteBuildpackCommand struct {
	RequiredArgs    flag.BuildpackName `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME delete-buildpack BUILDPACK [-f] [-s STACK]"`
	relatedCommands interface{}        `related_commands:"buildpacks"`
	Force           bool               `long:"force" short:"f" description:"Force deletion without confirmation"`
	Stack           string             `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous"`
	Actor           DeleteBuildpackActor
	UI              command.UI
	Config          command.Config
	SharedActor     command.SharedActor
}

func (cmd *DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)

	return nil
}

func (cmd DeleteBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if !cmd.Force {
		response, uiErr := cmd.UI.DisplayBoolPrompt(false, "Really delete the {{.ModelType}} {{.ModelName}}?", map[string]interface{}{
			"ModelType": "buildpack",
			"ModelName": cmd.RequiredArgs.Buildpack,
		})
		if uiErr != nil {
			return uiErr
		}

		if !response {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	if cmd.Stack == "" {
		cmd.UI.DisplayTextWithFlavor("Deleting buildpack {{.BuildpackName}}...", map[string]interface{}{
			"BuildpackName": cmd.RequiredArgs.Buildpack,
		})

	} else {
		cmd.UI.DisplayTextWithFlavor("Deleting buildpack {{.BuildpackName}} with stack {{.Stack}}...", map[string]interface{}{
			"BuildpackName": cmd.RequiredArgs.Buildpack,
			"Stack":         cmd.Stack,
		})
	}
	warnings, err := cmd.Actor.DeleteBuildpackByNameAndStack(cmd.RequiredArgs.Buildpack, cmd.Stack)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		switch err.(type) {
		case actionerror.BuildpackNotFoundError:
			if cmd.Stack == "" {
				cmd.UI.DisplayWarning("Buildpack {{.BuildpackName}} does not exist.", map[string]interface{}{
					"BuildpackName": cmd.RequiredArgs.Buildpack,
				})
			} else {
				cmd.UI.DisplayWarning("Buildpack {{.BuildpackName}} with stack {{.Stack}} not found.", map[string]interface{}{
					"BuildpackName": cmd.RequiredArgs.Buildpack,
					"Stack":         cmd.Stack,
				})
			}
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}
