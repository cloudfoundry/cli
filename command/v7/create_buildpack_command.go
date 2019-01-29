package v7

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/util/download"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . CreateBuildpackActor

type CreateBuildpackActor interface {
	CreateBuildpack(buildpack v7action.Buildpack) (v7action.Buildpack, v7action.Warnings, error)
	UploadBuildpack(guid string, pathToBuildpackBits string, progressBar v7action.SimpleProgressBar) (ccv3.JobURL, v7action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v7action.Downloader) (string, error)
	PollUploadBuildpackJob(jobURL ccv3.JobURL) (v7action.Warnings, error)
}

type CreateBuildpackCommand struct {
	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"Yes"`
	usage           interface{}              `usage:"CF_NAME create-buildpack BUILDPACK PATH POSITION [--disable]\n\nTIP:\n   Path should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest."`
	relatedCommands interface{}              `related_commands:"buildpacks, push"`
	Disable         bool                     `long:"disable" description:"Disable the buildpack from being used for staging"`

	UI          command.UI
	Config      command.Config
	ProgressBar v7action.SimpleProgressBar
	SharedActor command.SharedActor
	Actor       CreateBuildpackActor
}

func (cmd *CreateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.ProgressBar = v7action.NewProgressBar()

	return nil
}

func (cmd CreateBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Creating buildpack {{.BuildpackName}} as {{.Username}}...", map[string]interface{}{
		"Username":      user.Name,
		"BuildpackName": cmd.RequiredArgs.Buildpack,
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

	createdBuildpack, warnings, err := cmd.Actor.CreateBuildpack(v7action.Buildpack{
		Name:     cmd.RequiredArgs.Buildpack,
		Position: types.NullInt{IsSet: true, Value: cmd.RequiredArgs.Position},
		Enabled:  types.NullBool{IsSet: true, Value: !cmd.Disable},
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
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		cmd.displayTip()
		return err
	}
	cmd.UI.DisplayOK()

	cmd.UI.DisplayTextWithFlavor("Processing uploaded buildpack {{.BuildpackName}}...", map[string]interface{}{
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})
	warnings, err = cmd.Actor.PollUploadBuildpackJob(jobURL)
	cmd.UI.DisplayWarnings(warnings)

	if err != nil {
		cmd.displayTip()
		return err
	}

	cmd.UI.DisplayOK()
	return nil
}

func (cmd CreateBuildpackCommand) displayTip() {
	cmd.UI.DisplayTextWithFlavor(
		"TIP: A buildpack with name '{{.BuildpackName}}' and nil stack has been created. "+
			"Use '{{.CfDeleteBuildpackCommand}}' to delete it or "+
			"'{{.CfUpdateBuildpackCommand}}' to try again.",
		map[string]interface{}{
			"BuildpackName":            cmd.RequiredArgs.Buildpack,
			"CfDeleteBuildpackCommand": cmd.Config.BinaryName() + " delete-buildpack " + cmd.RequiredArgs.Buildpack,
			"CfUpdateBuildpackCommand": cmd.Config.BinaryName() + " update-buildpack " + cmd.RequiredArgs.Buildpack +
				" --assign-stack STACK --path " + string(cmd.RequiredArgs.Path),
		})
}

func (cmd CreateBuildpackCommand) displayTable(buildpacks []v7action.Buildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "filename"},
		}
		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{
				strconv.Itoa(buildpack.Position.Value),
				buildpack.Name,
				buildpack.Stack,
				strconv.FormatBool(buildpack.Enabled.Value),
				strconv.FormatBool(buildpack.Locked.Value),
				buildpack.Filename,
			})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
