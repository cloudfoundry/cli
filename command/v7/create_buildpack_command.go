package v7

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccerror"
    "code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/command"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
	"code.cloudfoundry.org/cli/v9/util/download"
)

type CreateBuildpackCommand struct {
	BaseCommand

	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"Yes"`
	usage           interface{}              `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--disable] [--lifecycle buildpack|cnb]\n\nTIP:\n   When using the 'buildpack' lifecycle type, Path should be a zip file, a url to a zip file, or a local directory. When using the 'cnb' lifecycle, Path should be a cnb file or gzipped oci image. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}              `related_commands:"buildpacks, push"`
	Disable         bool                     `long:"disable" description:"Disable the buildpack from being used for staging"`
	Lifecycle       string                   `long:"lifecycle" short:"l" description:"Lifecycle that the buildpack will use ('buildpack' or 'cnb')"`

	ProgressBar v7action.SimpleProgressBar
}

func (cmd *CreateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.ProgressBar = v7action.NewProgressBar()
	return cmd.BaseCommand.Setup(config, ui)
}

func (cmd CreateBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	if cmd.Lifecycle != "" {
		err = command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionBuildpackLifecycleQuery, "--lifecycle")
		if err != nil {
			return err
		}
	}

	cmd.UI.DisplayTextWithFlavor("Creating buildpack {{.BuildpackName}} as {{.Username}}...", map[string]interface{}{
		"Username":      user.Name,
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})

	downloader := download.NewDownloader(time.Second * 30)
	tmpDirPath, err := os.MkdirTemp("", "buildpack-dir-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDirPath)

	pathToBuildpackBits, err := cmd.Actor.PrepareBuildpackBits(string(cmd.RequiredArgs.Path), tmpDirPath, downloader)
	if err != nil {
		return err
	}

	createdBuildpack, warnings, err := cmd.Actor.CreateBuildpack(resources.Buildpack{
		Name:      cmd.RequiredArgs.Buildpack,
		Position:  types.NullInt{IsSet: true, Value: cmd.RequiredArgs.Position},
		Enabled:   types.NullBool{IsSet: true, Value: !cmd.Disable},
		Lifecycle: cmd.Lifecycle,
	})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.BuildpackName}} as {{.Username}}...", map[string]interface{}{
		"Username":      user.Name,
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})
	jobURL, warnings, err := cmd.Actor.UploadBuildpack(createdBuildpack.GUID, pathToBuildpackBits, cmd.ProgressBar)
	if _, ok := err.(ccerror.InvalidAuthTokenError); ok {
		cmd.UI.DisplayWarnings([]string{"Failed to upload buildpack due to auth token expiration, retrying..."})
		jobURL, warnings, err = cmd.Actor.UploadBuildpack(createdBuildpack.GUID, pathToBuildpackBits, cmd.ProgressBar)
	}
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return cmd.wrapWithTip(err)
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Processing uploaded buildpack {{.BuildpackName}}...", map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})
	warnings, err = cmd.Actor.PollUploadBuildpackJob(jobURL)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		return cmd.wrapWithTip(err)
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd CreateBuildpackCommand) wrapWithTip(err error) error {
	return translatableerror.TipDecoratorError{
		BaseError: err,
		Tip:       "A buildpack with name '{{.BuildpackName}}' and nil stack has been created. Use '{{.CfDeleteBuildpackCommand}}' to delete it or '{{.CfUpdateBuildpackCommand}}' to try again.",
		TipKeys: map[string]interface{}{
			"BuildpackName":            cmd.RequiredArgs.Buildpack,
			"CfDeleteBuildpackCommand": cmd.Config.BinaryName() + " delete-buildpack",
			"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack",
		},
	}
}
