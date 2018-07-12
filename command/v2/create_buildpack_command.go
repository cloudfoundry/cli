package v2

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/util/progressbar"
)

//go:generate counterfeiter . CreateBuildpackActor

type CreateBuildpackActor interface {
	CreateBuildpack(name string, position int, enabled bool) (v2action.Buildpack, v2action.Warnings, error)
	UploadBuildpack(GUID string, path string, pb progressbar.ProgressBar) (v2action.Warnings, error)
}

type CreateBuildpackCommand struct {
	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"yes"`
	Disable         bool                     `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                     `long:"enable" description:"Enable the buildpack to be used for staging"`
	usage           interface{}              `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--enable|--disable]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}              `related_commands:"buildpacks, push"`

	UI          command.UI
	Actor       CreateBuildpackActor
	SharedActor command.SharedActor
	Config      command.Config
}

func (cmd *CreateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd *CreateBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
		"Buildpack": cmd.RequiredArgs.Buildpack,
		"Username":  user.Name,
	})
	buildpack, warnings, err := cmd.Actor.CreateBuildpack(cmd.RequiredArgs.Buildpack, cmd.RequiredArgs.Position.Value, !cmd.Disable)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
		"Buildpack": cmd.RequiredArgs.Buildpack,
		"Username":  user.Name,
	})

	uploadWarnings, err := cmd.Actor.UploadBuildpack(buildpack.GUID, string(cmd.RequiredArgs.Path), progressbar.ProgressBar{})
	cmd.UI.DisplayWarnings(uploadWarnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayText("Done uploading")
	cmd.UI.DisplayOK()
	return nil
}
