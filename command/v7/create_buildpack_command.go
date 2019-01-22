package v7

import (
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
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . CreateBuildpackActor

type CreateBuildpackActor interface {
	CreateBuildpack(buildpack v7action.Buildpack) (v7action.Buildpack, v7action.Warnings, error)
	UploadBuildpack(guid string, pathToBuildpackBits string, progressBar v7action.SimpleProgressBar) (v7action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v7action.Downloader) (string, error)
}

type CreateBuildpackCommand struct {
	RequiredArgs    flag.CreateBuildpackArgs `positional-args:"Yes"`
	usage           interface{}              `usage:"CF_NAME create-buildpack"`
	relatedCommands interface{}              `related_commands:"push"`

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
	cmd.UI.DisplayNewline()

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
		Position: cmd.RequiredArgs.Position,
	})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()
	cmd.UI.DisplayNewline()

	cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.BuildpackName}} as {{.Username}}...", map[string]interface{}{
		"Username":      user.Name,
		"BuildpackName": cmd.RequiredArgs.Buildpack,
	})
	warnings, err = cmd.Actor.UploadBuildpack(createdBuildpack.GUID, pathToBuildpackBits, cmd.ProgressBar)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayText("Done uploading")
	cmd.UI.DisplayOK()

	return nil
}

func (cmd CreateBuildpackCommand) displayTable(buildpacks []v7action.Buildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "filename"},
		}
		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{strconv.Itoa(buildpack.Position), buildpack.Name, buildpack.Stack, strconv.FormatBool(buildpack.Enabled), strconv.FormatBool(buildpack.Locked), buildpack.Filename})
		}

		cmd.UI.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
