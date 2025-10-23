package v7

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/command/flag"
	"code.cloudfoundry.org/cli/v8/command/translatableerror"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/download"
)

type UpdateBuildpackActor interface {
	UpdateBuildpackByNameAndStack(buildpackName string, buildpackStack string, buildpack resources.Buildpack) (resources.Buildpack, v7action.Warnings, error)
	UploadBuildpack(guid string, pathToBuildpackBits string, progressBar v7action.SimpleProgressBar) (ccv3.JobURL, v7action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v7action.Downloader) (string, error)
	PollUploadBuildpackJob(jobURL ccv3.JobURL) (v7action.Warnings, error)
}

type UpdateBuildpackCommand struct {
	BaseCommand

	RequiredArgs    flag.BuildpackName               `positional-args:"Yes"`
	usage           interface{}                      `usage:"CF_NAME update-buildpack BUILDPACK [-p PATH | -s STACK | --assign-stack NEW_STACK] [-i POSITION] [--rename NEW_NAME] [--enable|--disable] [--lock|--unlock]\n\nTIP:\nWhen using the 'buildpack' lifecycle type, Path should be a zip file, a url to a zip file, or a local directory. When using the 'cnb' lifecycle, Path should be a cnb file or gzipped oci image. Position is a positive integer, sets priority, and is sorted from lowest to highest.\n\nUse '--assign-stack' with caution. Associating a buildpack with a stack that it does not support may result in undefined behavior. Additionally, changing this association once made may require a local copy of the buildpack.\n\n"`
	relatedCommands interface{}                      `related_commands:"buildpacks, create-buildpack, delete-buildpack"`
	NewStack        string                           `long:"assign-stack" description:"Assign a stack to a buildpack that does not have a stack association"`
	Disable         bool                             `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                             `long:"enable" description:"Enable the buildpack to be used for staging"`
	Lock            bool                             `long:"lock" description:"Lock the buildpack to prevent updates"`
	Path            flag.PathWithExistenceCheckOrURL `long:"path" short:"p" description:"Path to directory or zip file"`
	Position        types.NullInt                    `long:"position" short:"i" description:"The order in which the buildpacks are checked during buildpack auto-detection"`
	NewName         string                           `long:"rename" description:"Rename an existing buildpack"`
	CurrentStack    string                           `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	Lifecycle       string                           `long:"lifecycle" short:"l" description:"Specify lifecycle to disambiguate buildpacks with the same name"`
	Unlock          bool                             `long:"unlock" description:"Unlock the buildpack to enable updates"`

	ProgressBar v7action.SimpleProgressBar
}

func (cmd *UpdateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.ProgressBar = v7action.NewProgressBar()
	return cmd.BaseCommand.Setup(config, ui)
}

func (cmd UpdateBuildpackCommand) Execute(args []string) error {
	var buildpackBitsPath, tmpDirPath string

	user, err := cmd.validateSetup()
	if err != nil {
		return err
	}

	if cmd.Lifecycle != "" {
		err = command.MinimumCCAPIVersionCheck(cmd.Config.APIVersion(), ccversion.MinVersionBuildpackLifecycleQuery, "--lifecycle")
		if err != nil {
			return err
		}
	}

	cmd.printInitialText(user.Name)

	if cmd.Path != "" {
		buildpackBitsPath, tmpDirPath, err = cmd.prepareBuildpackBits()
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDirPath)
	}

	updatedBuildpack, err := cmd.updateBuildpack()
	if err != nil {
		return err
	}

	if buildpackBitsPath != "" {
		return cmd.uploadBits(user, updatedBuildpack, buildpackBitsPath)
	}

	return nil
}

func (cmd UpdateBuildpackCommand) validateSetup() (configv3.User, error) {
	var user configv3.User

	err := cmd.validateFlagCombinations()
	if err != nil {
		return user, err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return user, err
	}

	return cmd.Actor.GetCurrentUser()
}

func (cmd UpdateBuildpackCommand) prepareBuildpackBits() (string, string, error) {
	downloader := download.NewDownloader(time.Second * 30)
	tmpDirPath, err := os.MkdirTemp("", "buildpack-dir-")
	if err != nil {
		return "", "", err
	}

	buildpackBits, err := cmd.Actor.PrepareBuildpackBits(string(cmd.Path), tmpDirPath, downloader)
	return buildpackBits, tmpDirPath, err
}

func (cmd UpdateBuildpackCommand) updateBuildpack() (resources.Buildpack, error) {
	var desiredBuildpack resources.Buildpack

	desiredBuildpack.Enabled = types.NullBool{IsSet: cmd.Enable || cmd.Disable, Value: cmd.Enable}
	desiredBuildpack.Locked = types.NullBool{IsSet: cmd.Lock || cmd.Unlock, Value: cmd.Lock}
	desiredBuildpack.Position = cmd.Position

	if cmd.NewStack != "" {
		desiredBuildpack.Stack = cmd.NewStack
	}

	if cmd.NewName != "" {
		desiredBuildpack.Name = cmd.NewName
	}

	updatedBuildpack, warnings, err := cmd.Actor.UpdateBuildpackByNameAndStackAndLifecycle(
		cmd.RequiredArgs.Buildpack,
		cmd.CurrentStack,
		cmd.Lifecycle,
		desiredBuildpack,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return updatedBuildpack, err
	}
	cmd.UI.DisplayOK()

	return updatedBuildpack, nil
}

func (cmd UpdateBuildpackCommand) uploadBits(user configv3.User, updatedBuildpack resources.Buildpack, buildpackBitsPath string) error {
	cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
		"Buildpack": cmd.RequiredArgs.Buildpack,
		"Username":  user.Name,
	})

	jobURL, warnings, err := cmd.Actor.UploadBuildpack(
		updatedBuildpack.GUID,
		buildpackBitsPath,
		cmd.ProgressBar,
	)
	if _, ok := err.(ccerror.InvalidAuthTokenError); ok {
		cmd.UI.DisplayWarnings([]string{"Failed to upload buildpack due to auth token expiration, retrying..."})
		jobURL, warnings, err = cmd.Actor.UploadBuildpack(updatedBuildpack.GUID, buildpackBitsPath, cmd.ProgressBar)
	}
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Processing uploaded buildpack {{.BuildpackName}}...", map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})

	warnings, err = cmd.Actor.PollUploadBuildpackJob(jobURL)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd UpdateBuildpackCommand) printInitialText(userName string) {
	var originalBuildpackName = cmd.RequiredArgs.Buildpack
	var buildpackName = originalBuildpackName

	if cmd.NewName != "" {
		buildpackName = cmd.NewName
		cmd.UI.DisplayTextWithFlavor("Renaming buildpack {{.Buildpack}} to {{.DesiredBuildpackName}} as {{.CurrentUser}}...\n", map[string]interface{}{
			"Buildpack":            originalBuildpackName,
			"CurrentUser":          userName,
			"DesiredBuildpackName": cmd.NewName,
		})
	}

	if cmd.NewStack != "" {
		cmd.UI.DisplayTextWithFlavor("Assigning stack {{.Stack}} to {{.Buildpack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   buildpackName,
			"CurrentUser": userName,
			"Stack":       cmd.NewStack,
		})
		if cmd.Position.IsSet || cmd.Lock || cmd.Unlock || cmd.Enable || cmd.Disable {
			cmd.UI.DisplayNewline()
			cmd.displayUpdatingBuildpackMessage(buildpackName, userName, cmd.NewStack)
		}
	} else {
		cmd.displayUpdatingBuildpackMessage(buildpackName, userName, cmd.CurrentStack)
	}
}

func (cmd UpdateBuildpackCommand) displayUpdatingBuildpackMessage(buildpackName, userName, stack string) {
	message := "Updating buildpack {{.Buildpack}}"
	if stack != "" {
		message += " with stack {{.Stack}}"
	}
	if cmd.Lifecycle != "" {
		message += " with lifecycle {{.Lifecycle}}"
	}
	message += " as {{.CurrentUser}}..."
	cmd.UI.DisplayTextWithFlavor(message, map[string]interface{}{
		"Buildpack":   buildpackName,
		"CurrentUser": userName,
		"Stack":       stack,
		"Lifecycle":   cmd.Lifecycle,
	})
}

func (cmd UpdateBuildpackCommand) validateFlagCombinations() error {
	if cmd.Lock && cmd.Unlock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--lock", "--unlock"},
		}
	}

	if cmd.Enable && cmd.Disable {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--enable", "--disable"},
		}
	}

	if cmd.Path != "" && cmd.Lock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--path", "--lock"},
		}
	}

	if cmd.Path != "" && cmd.NewStack != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--path", "--assign-stack"},
		}
	}

	if cmd.CurrentStack != "" && cmd.NewStack != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--stack", "--assign-stack"},
		}
	}
	return nil
}
