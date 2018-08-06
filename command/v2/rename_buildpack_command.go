package v2

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . RenameBuildpackActor

type RenameBuildpackActor interface {
	GetBuildpackByName(name string) (v2action.Buildpack, v2action.Warnings, error)
	UpdateBuildpack(buildpack v2action.Buildpack) (v2action.Buildpack, v2action.Warnings, error)
}

type RenameBuildpackCommand struct {
	RequiredArgs    flag.RenameBuildpackArgs `positional-args:"yes"`
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

	cmd.UI.DisplayTextWithFlavor("Renaming buildpack {{.OldName}} to {{.NewName}}...", map[string]interface{}{
		"OldName": cmd.RequiredArgs.OldBuildpackName,
		"NewName": cmd.RequiredArgs.NewBuildpackName,
	})

	oldBp, warnings, err := cmd.Actor.GetBuildpackByName(cmd.RequiredArgs.OldBuildpackName)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		if _, notFound := err.(actionerror.BuildpackNotFoundError); notFound {
			return translatableerror.ConvertToTranslatableError(err)
		} else if _, ambiguousBuildpacks := err.(actionerror.MultipleBuildpacksFoundError); ambiguousBuildpacks {
			return translatableerror.ConvertToTranslatableError(err)
		}
		return err
	}

	oldBp.Name = cmd.RequiredArgs.NewBuildpackName

	_, warnings, err = cmd.Actor.UpdateBuildpack(oldBp)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
