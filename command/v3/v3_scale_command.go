package v3

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v3/shared"
	"github.com/cloudfoundry/bytefmt"
)

//go:generate counterfeiter . V3ScaleActor

type V3ScaleActor interface {
	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetProcessByApplication(appGUID string) (ccv3.Process, v3action.Warnings, error)
	ScaleProcessByApplication(appGUID string, process ccv3.Process) (ccv3.Process, v3action.Warnings, error)
}

type V3ScaleCommand struct {
	RequiredArgs    flag.AppName   `positional-args:"yes"`
	Instances       int            `short:"i" description:"Number of instances"`
	DiskLimit       flag.Megabytes `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit     flag.Megabytes `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	usage           interface{}    `usage:"CF_NAME v3-scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY]"`
	relatedCommands interface{}    `related_commands:"v3-push"`

	UI          command.UI
	Config      command.Config
	Actor       V3ScaleActor
	SharedActor command.SharedActor
}

func (cmd *V3ScaleCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, _, err := shared.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.Actor = v3action.NewActor(ccClient, config)

	return nil
}

func (cmd V3ScaleCommand) Execute(args []string) error {
	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	var actorErr error
	// TODO: distinguish between user provided instance value and default
	if cmd.Instances == 0 && cmd.DiskLimit.Size == 0 && cmd.MemoryLimit.Size == 0 {
		actorErr = cmd.getAndDisplayProcess(app.GUID, user.Name)
	} else {
		actorErr = cmd.scaleAndDisplayProcess(app.GUID, user.Name)
	}
	if actorErr != nil {
		return actorErr
	}

	return nil
}

func (cmd V3ScaleCommand) getAndDisplayProcess(appGUID string, username string) error {
	cmd.UI.DisplayTextWithFlavor("Showing current scale of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	process, warnings, err := cmd.Actor.GetProcessByApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.displayProcessSummary(process)
	return nil
}

func (cmd V3ScaleCommand) scaleAndDisplayProcess(appGUID string, username string) error {
	cmd.UI.DisplayTextWithFlavor("Scaling app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	ccv3Process := ccv3.Process{
		Type:       "web",
		Instances:  cmd.Instances,
		MemoryInMB: int(cmd.MemoryLimit.Size),
		DiskInMB:   int(cmd.DiskLimit.Size),
	}
	process, warnings, err := cmd.Actor.ScaleProcessByApplication(appGUID, ccv3Process)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.displayProcessSummary(process)
	return nil
}

func (cmd V3ScaleCommand) displayProcessSummary(process ccv3.Process) {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayKeyValueTable("", [][]string{
		{cmd.UI.TranslateText("instances:"), strconv.Itoa(process.Instances)},
		{cmd.UI.TranslateText("memory:"), bytefmt.ByteSize(uint64(process.MemoryInMB) * bytefmt.MEGABYTE)},
		{cmd.UI.TranslateText("disk:"), bytefmt.ByteSize(uint64(process.DiskInMB) * bytefmt.MEGABYTE)},
	}, 3)
}
