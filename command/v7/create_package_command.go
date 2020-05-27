package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

type CreatePackageCommand struct {
	BaseCommand

	RequiredArgs    flag.AppName                `positional-args:"yes"`
	DockerImage     flag.DockerImage            `long:"docker-image" short:"o" description:"Docker image to use (e.g. user/docker-image-name)"`
	AppPath         flag.PathWithExistenceCheck `short:"p" description:"Path to app directory or to a zip file of the contents of the app directory"`
	usage           interface{}                 `usage:"CF_NAME create-package APP_NAME [-p APP_PATH | --docker-image [REGISTRY_HOST:PORT/]IMAGE[:TAG]]"`
	relatedCommands interface{}                 `related_commands:"app, droplets, packages, push"`

	PackageDisplayer shared.PackageDisplayer
}

func (cmd *CreatePackageCommand) Setup(config command.Config, ui command.UI) error {
	cmd.PackageDisplayer = shared.NewPackageDisplayer(ui, config)
	return cmd.BaseCommand.Setup(config, ui)
}

func (cmd CreatePackageCommand) Execute(args []string) error {
	if cmd.DockerImage.Path != "" && cmd.AppPath != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--docker-image", "-o", "-p"},
		}
	}

	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	isDockerImage := (cmd.DockerImage.Path != "")
	err = cmd.PackageDisplayer.DisplaySetupMessage(cmd.RequiredArgs.AppName, isDockerImage)
	if err != nil {
		return err
	}

	var (
		pkg      v7action.Package
		warnings v7action.Warnings
	)
	if isDockerImage {
		pkg, warnings, err = cmd.Actor.CreateDockerPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, v7action.DockerImageCredentials{Path: cmd.DockerImage.Path})
	} else {
		pkg, warnings, err = cmd.Actor.CreateAndUploadBitsPackageByApplicationNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, string(cmd.AppPath))
	}

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Package with guid '{{.PackageGuid}}' has been created.", map[string]interface{}{
		"PackageGuid": pkg.GUID,
	})
	cmd.UI.DisplayOK()

	return nil
}
