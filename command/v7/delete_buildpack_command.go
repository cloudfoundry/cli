package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
    "code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
)

type DeleteBuildpackCommand struct {
	BaseCommand

	RequiredArgs    flag.BuildpackName `positional-args:"yes"`
	usage           interface{}        `usage:"CF_NAME delete-buildpack BUILDPACK [-f] [-s STACK] [-l LIFECYCLE]"`
	relatedCommands interface{}        `related_commands:"buildpacks"`
	Force           bool               `long:"force" short:"f" description:"Force deletion without confirmation"`
	Stack           string             `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous"`
	Lifecycle       string             `long:"lifecycle" short:"l" description:"Specify lifecycle to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous"`
}

func (cmd DeleteBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.Lifecycle != "" {
		err = command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionBuildpackLifecycleQuery, "--lifecycle")
		if err != nil {
			return err
		}
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

	cmd.displayBuildpackDeletingMessage()
	warnings, err := cmd.Actor.DeleteBuildpackByNameAndStackAndLifecycle(cmd.RequiredArgs.Buildpack, cmd.Stack, cmd.Lifecycle)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		switch err.(type) {
		case actionerror.BuildpackNotFoundError:
			cmd.displayBuildpackNotFoundWarning()
		default:
			return err
		}
	}

	cmd.UI.DisplayOK()

	return nil
}

func (cmd DeleteBuildpackCommand) displayBuildpackNotFoundWarning() {
	warning := "Buildpack '{{.BuildpackName}}'"
	if cmd.Stack != "" {
		warning += " with stack '{{.Stack}}'"
	}

	if cmd.Lifecycle != "" {
		warning += " with lifecycle '{{.Lifecycle}}'"
	}

	if cmd.Stack != "" && cmd.Lifecycle != "" {
		warning += " does not exist."
	} else {
		warning += " not found."
	}
	cmd.UI.DisplayWarning(warning, map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.Buildpack,
		"Stack":         cmd.Stack,
		"Lifecycle":     cmd.Lifecycle,
	})
}

func (cmd DeleteBuildpackCommand) displayBuildpackDeletingMessage() {
	message := "Deleting buildpack {{.BuildpackName}}"
	if cmd.Stack != "" {
		message += " with stack {{.Stack}}"
	}

	if cmd.Lifecycle != "" {
		message += " with lifecycle {{.Lifecycle}}"
	}

	cmd.UI.DisplayTextWithFlavor(message+"...", map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.Buildpack,
		"Stack":         cmd.Stack,
		"Lifecycle":     cmd.Lifecycle,
	})
}
