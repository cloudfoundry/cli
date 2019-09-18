package v6

import (
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/util/download"
)

//go:generate counterfeiter . Downloader

type Downloader interface {
	Download(string) (string, error)
}

//go:generate counterfeiter . CreateBuildpackActor

type CreateBuildpackActor interface {
	CreateBuildpack(name string, position int, enabled bool) (v2action.Buildpack, v2action.Warnings, error)
	UploadBuildpack(GUID string, path string, progBar v2action.SimpleProgressBar) (v2action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v2action.Downloader) (string, error)
}

type CreateBuildpackCommand struct {
	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"yes"`
	Disable         bool                     `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                     `long:"enable" description:"Enable the buildpack to be used for staging"`
	usage           interface{}              `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--enable|--disable]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}              `related_commands:"buildpacks, push"`

	UI          command.UI
	Actor       CreateBuildpackActor
	ProgressBar v2action.SimpleProgressBar
	SharedActor command.SharedActor
	Config      command.Config
}

func (cmd *CreateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	ccClient, uaaClient, err := shared.GetNewClientsAndConnectToCF(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.ProgressBar = v2action.NewProgressBar()

	return nil
}

func (cmd *CreateBuildpackCommand) Execute(args []string) error {
	if cmd.Enable && cmd.Disable {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--enable", "--disable"},
		}
	}

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

	downloader := download.NewDownloader(time.Second * 30)
	tmpDirPath, err := ioutil.TempDir("", "buildpack-dir-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDirPath)

	pathToBuildpackBits, err := cmd.Actor.PrepareBuildpackBits(string(cmd.RequiredArgs.Path), tmpDirPath, downloader)
	if err != nil {
		return err
	}

	buildpack, warnings, err := cmd.Actor.CreateBuildpack(cmd.RequiredArgs.Buildpack, cmd.RequiredArgs.Position, !cmd.Disable)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return cmd.displayIfNameCollisionError(err)
	}

	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
		"Buildpack": cmd.RequiredArgs.Buildpack,
		"Username":  user.Name,
	})

	uploadWarnings, err := cmd.Actor.UploadBuildpack(buildpack.GUID, pathToBuildpackBits, cmd.ProgressBar)
	cmd.UI.DisplayWarnings(uploadWarnings)
	if err != nil {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayNewline()
		if _, ok := err.(actionerror.BuildpackAlreadyExistsForStackError); ok {
			cmd.displayNameAndStackCollisionError(err)
			return nil
		} else if httpErr, ok := err.(download.RawHTTPStatusError); ok {
			return translatableerror.HTTPStatusError{Status: httpErr.Status}
		}
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Done uploading")
	cmd.UI.DisplayOK()

	return nil
}

func (cmd CreateBuildpackCommand) displayNameAndStackCollisionError(err error) {
	cmd.UI.DisplayWarning(err.Error())
	cmd.UI.DisplayTextWithFlavor("TIP: use '{{.CfUpdateBuildpackCommand}}' to update this buildpack",
		map[string]interface{}{
			"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack",
		})
}

func (cmd CreateBuildpackCommand) displayIfNameCollisionError(err error) error {
	if _, ok := err.(actionerror.BuildpackAlreadyExistsWithoutStackError); ok {
		cmd.displayAlreadyExistingBuildpackWithoutStack(err)
		return nil
	} else if _, ok := err.(actionerror.BuildpackNameTakenError); ok {
		cmd.displayAlreadyExistingBuildpack(err)
		return nil
	}
	return err
}

func (cmd CreateBuildpackCommand) displayAlreadyExistingBuildpackWithoutStack(err error) {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayWarning(err.Error())
	cmd.UI.DisplayTextWithFlavor("TIP: use '{{.CfBuildpacksCommand}}' and '{{.CfDeleteBuildpackCommand}}' to delete buildpack {{.BuildpackName}} without a stack",
		map[string]interface{}{
			"CfBuildpacksCommand":      cmd.Config.BinaryName() + " buildpacks",
			"CfDeleteBuildpackCommand": cmd.Config.BinaryName() + " delete-buildpack",
			"BuildpackName":            cmd.RequiredArgs.Buildpack,
		})
}

func (cmd CreateBuildpackCommand) displayAlreadyExistingBuildpack(err error) {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayWarning(err.Error())
	cmd.UI.DisplayTextWithFlavor("TIP: use '{{.CfUpdateBuildpackCommand}}' to update this buildpack",
		map[string]interface{}{
			"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack",
		})
}
