package v3

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3ScaleActor

type V3ScaleActor interface {
	shared.V3AppSummaryActor

	GetApplicationByNameAndSpace(appName string, spaceGUID string) (v3action.Application, v3action.Warnings, error)
	GetInstancesByApplicationAndProcessType(appGUID string, processType string) (v3action.Process, v3action.Warnings, error)
	ScaleProcessByApplication(appGUID string, process ccv3.Process) (v3action.Process, v3action.Warnings, error)
	StopApplication(appGUID string) (v3action.Warnings, error)
	StartApplication(appGUID string) (v3action.Application, v3action.Warnings, error)
	PollStart(appGUID string, warnings chan<- v3action.Warnings) error
}

type V3ScaleCommand struct {
	RequiredArgs        flag.AppName   `positional-args:"yes"`
	Instances           flag.Instances `short:"i" description:"Number of instances"`
	DiskLimit           flag.Megabytes `short:"k" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit         flag.Megabytes `short:"m" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	usage               interface{}    `usage:"CF_NAME v3-scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY]"`
	relatedCommands     interface{}    `related_commands:"v3-push"`
	envCFStartupTimeout interface{}    `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI                  command.UI
	Config              command.Config
	Actor               V3ScaleActor
	SharedActor         command.SharedActor
	AppSummaryDisplayer shared.AppSummaryDisplayer
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

	cmd.AppSummaryDisplayer = shared.AppSummaryDisplayer{
		UI:              ui,
		Config:          config,
		Actor:           cmd.Actor,
		V2AppRouteActor: nil,
		AppName:         cmd.RequiredArgs.AppName,
	}

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

	if cmd.Instances.IsSet == false && cmd.DiskLimit.Size == 0 && cmd.MemoryLimit.Size == 0 {
		return cmd.getAndDisplayProcess(app.GUID, user.Name)
	}

	process, actorErr := cmd.scaleProcess(app.GUID, user.Name)

	cmd.UI.DisplayText("Waiting for app to start...")

	pollWarnings := make(chan v3action.Warnings)
	done := make(chan bool)
	go func() {
		for {
			select {
			case message := <-pollWarnings:
				cmd.UI.DisplayWarnings(message)
			case <-done:
				return
			}
		}
	}()

	err = cmd.Actor.PollStart(app.GUID, pollWarnings)
	done <- true

	if err != nil {
		if _, ok := err.(v3action.StartupTimeoutError); ok {
			return translatableerror.StartupTimeoutError{
				AppName:    cmd.RequiredArgs.AppName,
				BinaryName: cmd.Config.BinaryName(),
			}
		} else {
			return shared.HandleError(err)
		}
	}

	cmd.AppSummaryDisplayer.DisplayAppInstancesTable(process)
	return actorErr
}

func (cmd V3ScaleCommand) getAndDisplayProcess(appGUID string, username string) error {
	cmd.UI.DisplayTextWithFlavor("Showing current scale of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	process, warnings, err := cmd.Actor.GetInstancesByApplicationAndProcessType(appGUID, "web")
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.AppSummaryDisplayer.DisplayAppInstancesTable(process)
	return nil
}

func (cmd V3ScaleCommand) scaleProcess(appGUID string, username string) (v3action.Process, error) {
	cmd.UI.DisplayTextWithFlavor("Scaling app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	shouldRestart := cmd.DiskLimit.Size != 0 || cmd.MemoryLimit.Size != 0
	if shouldRestart {
		cmd.UI.DisplayNewline()
		shouldScale, err := cmd.UI.DisplayBoolPrompt(
			false,
			"This will cause the app to restart. Are you sure you want to scale {{.AppName}}?",
			map[string]interface{}{"AppName": cmd.RequiredArgs.AppName})
		if err != nil {
			return v3action.Process{}, err
		}

		if !shouldScale {
			cmd.UI.DisplayText("Scaling cancelled")
			return v3action.Process{}, nil
		}
	}

	ccv3Process := ccv3.Process{
		Type:       "web",
		Instances:  cmd.Instances.Value,
		MemoryInMB: int(cmd.MemoryLimit.Size),
		DiskInMB:   int(cmd.DiskLimit.Size),
	}
	process, warnings, err := cmd.Actor.ScaleProcessByApplication(appGUID, ccv3Process)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return v3action.Process{}, err
	}

	if shouldRestart {
		err := cmd.restartApplication(appGUID, username)
		if err != nil {
			return v3action.Process{}, err
		}
	}

	return process, nil
}

func (cmd V3ScaleCommand) restartApplication(appGUID string, username string) error {
	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTextWithFlavor("Stopping app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	warnings, err := cmd.Actor.StopApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayNewline()
	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})

	_, warnings, err = cmd.Actor.StartApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	return nil
}
