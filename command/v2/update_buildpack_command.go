package v2

import (
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/download"
)

//go:generate counterfeiter . UpdateBuildpackActor
type UpdateBuildpackActor interface {
	CloudControllerAPIVersion() string
	UpdateBuildpackByName(name string, position types.NullInt) (string, v2action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v2action.Downloader) (string, error)
	UploadBuildpack(GUID string, path string, progBar v2action.SimpleProgressBar) (v2action.Warnings, error)
}

type UpdateBuildpackCommand struct {
	RequiredArgs    flag.BuildpackName               `positional-args:"yes"`
	Disable         bool                             `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                             `long:"enable" description:"Enable the buildpack to be used for staging"`
	Order           types.NullInt                    `short:"i" description:"The order in which the buildpacks are checked during buildpack auto-detection"`
	Lock            bool                             `long:"lock" description:"Lock the buildpack to prevent updates"`
	Path            flag.PathWithExistenceCheckOrURL `short:"p" description:"Path to directory or zip file"`
	Unlock          bool                             `long:"unlock" description:"Unlock the buildpack to enable updates"`
	Stack           string                           `short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	usage           interface{}                      `usage:"CF_NAME update-buildpack BUILDPACK [-p PATH] [-i POSITION] [-s STACK] [--enable|--disable] [--lock|--unlock]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}                      `related_commands:"buildpacks, rename-buildpack"`

	UI          command.UI
	SharedActor command.SharedActor
	Actor       UpdateBuildpackActor
	Config      command.Config
	ProgressBar v2action.SimpleProgressBar
}

func (cmd *UpdateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	if !config.Experimental() {
		return translatableerror.UnrefactoredCommandError{}
	}

	cmd.UI = ui
	cmd.Config = config

	cmd.SharedActor = sharedaction.NewActor(config)
	ccClient, uaaClient, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient, config)
	cmd.ProgressBar = v2action.NewProgressBar()

	return nil
}

func (cmd UpdateBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if cmd.Stack != "" {
		err = command.MinimumCCAPIVersionCheck(
			cmd.Actor.CloudControllerAPIVersion(),
			ccversion.MinVersionBuildpackStackAssociationV2,
			"Option `-s`",
		)
		if err != nil {
			return err
		}
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} as {{.CurrentUser}}...", map[string]interface{}{
		"Buildpack":   cmd.RequiredArgs.Buildpack,
		"CurrentUser": user.Name,
	})

	buildpackGuid, warnings, err := cmd.Actor.UpdateBuildpackByName(cmd.RequiredArgs.Buildpack, cmd.Order)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Done updating")
	cmd.UI.DisplayOK()

	if cmd.Path != "" {
		var (
			tmpDirPath string
			bitsPath   string
		)
		downloader := download.NewDownloader(time.Second * 30)
		tmpDirPath, err = ioutil.TempDir("", "buildpack-dir-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDirPath)

		bitsPath, err = cmd.Actor.PrepareBuildpackBits(string(cmd.Path), tmpDirPath, downloader)
		if err != nil {
			return err
		}

		cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
			"Buildpack": cmd.RequiredArgs.Buildpack,
			"Username":  user.Name,
		})

		warnings, err = cmd.Actor.UploadBuildpack(buildpackGuid, bitsPath, cmd.ProgressBar)
		cmd.UI.DisplayWarnings(warnings)

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Done uploading")
		cmd.UI.DisplayOK()

	}
	return err
}
