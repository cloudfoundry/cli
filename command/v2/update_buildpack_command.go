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
	UpdateBuildpackByNameAndStack(name, stack string, position types.NullInt, locked types.NullBool, enabled types.NullBool) (string, v2action.Warnings, error)
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
	err := cmd.validateFlagCombinations()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	err = cmd.minAPIVersionCheck()
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.printInitialText(user.Name)

	var buildpackBitsPath string

	if cmd.Path != "" {
		var (
			tmpDirPath string
		)
		downloader := download.NewDownloader(time.Second * 30)
		tmpDirPath, err = ioutil.TempDir("", "buildpack-dir-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDirPath)

		buildpackBitsPath, err = cmd.Actor.PrepareBuildpackBits(string(cmd.Path), tmpDirPath, downloader)
		if err != nil {
			return err
		}
	}

	enabled := types.NullBool{
		IsSet: cmd.Enable || cmd.Disable,
		Value: cmd.Enable,
	}

	locked := types.NullBool{
		IsSet: cmd.Lock || cmd.Unlock,
		Value: cmd.Lock,
	}

	buildpackGUID, warnings, err := cmd.Actor.UpdateBuildpackByNameAndStack(cmd.RequiredArgs.Buildpack, cmd.Stack, cmd.Order, locked, enabled)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayText("Done updating")
	cmd.UI.DisplayOK()

	if buildpackBitsPath != "" {
		cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
			"Buildpack": cmd.RequiredArgs.Buildpack,
			"Username":  user.Name,
		})

		warnings, err = cmd.Actor.UploadBuildpack(buildpackGUID, buildpackBitsPath, cmd.ProgressBar)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("Done uploading")
		cmd.UI.DisplayOK()

	}
	return err
}

func (cmd UpdateBuildpackCommand) minAPIVersionCheck() error {
	if cmd.Stack != "" {
		return command.MinimumCCAPIVersionCheck(
			cmd.Actor.CloudControllerAPIVersion(),
			ccversion.MinVersionBuildpackStackAssociationV2,
			"Option '-s'",
		)
	}
	return nil
}

func (cmd UpdateBuildpackCommand) printInitialText(userName string) {
	if cmd.Stack == "" {
		cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   cmd.RequiredArgs.Buildpack,
			"CurrentUser": userName,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} with stack {{.Stack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   cmd.RequiredArgs.Buildpack,
			"CurrentUser": userName,
			"Stack":       cmd.Stack,
		})
	}
}

func (cmd UpdateBuildpackCommand) validateFlagCombinations() error {
	if cmd.Lock && cmd.Unlock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--lock", "--unlock"},
		}
	}

	if cmd.Lock && len(cmd.Path) > 0 {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-p", "--lock"},
		}
	}

	if len(cmd.Path) > 0 && cmd.Unlock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"-p", "--unlock"},
		}
	}

	if cmd.Enable && cmd.Disable {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--enable", "--disable"},
		}
	}

	return nil
}
