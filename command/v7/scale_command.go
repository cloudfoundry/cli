package v7

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	sharedV2 "code.cloudfoundry.org/cli/command/v6/shared"
	"code.cloudfoundry.org/cli/command/v7/shared"
)

//go:generate counterfeiter . ScaleActor

type ScaleActor interface {
	AppActor

	ScaleProcessByApplication(appGUID string, process v7action.Process) (v7action.Warnings, error)
	StopApplication(appGUID string) (v7action.Warnings, error)
	StartApplication(appGUID string) (v7action.Application, v7action.Warnings, error)
	PollStart(appGUID string) (v7action.Warnings, error)
}

type ScaleCommand struct {
	RequiredArgs        flag.AppName   `positional-args:"yes"`
	Force               bool           `short:"f" description:"Force restart of app without prompt"`
	Instances           flag.Instances `short:"i" required:"false" description:"Number of instances"`
	DiskLimit           flag.Megabytes `short:"k" required:"false" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	MemoryLimit         flag.Megabytes `short:"m" required:"false" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	ProcessType         string         `long:"process" default:"web" description:"App process to scale"`
	usage               interface{}    `usage:"CF_NAME scale APP_NAME [--process PROCESS] [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"`
	relatedCommands     interface{}    `related_commands:"v3-push"`
	envCFStartupTimeout interface{}    `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`

	UI          command.UI
	Config      command.Config
	Actor       ScaleActor
	SharedActor command.SharedActor
	RouteActor  v7action.RouteActor
}

func (cmd *ScaleCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)

	ccClientV2, _, err := sharedV2.NewClients(config, ui, true)
	if err != nil {
		return err
	}
	cmd.RouteActor = v2action.NewActor(ccClientV2, uaaClient, config)

	return nil
}

func (cmd ScaleCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if !cmd.Instances.IsSet && !cmd.DiskLimit.IsSet && !cmd.MemoryLimit.IsSet {
		return cmd.showCurrentScale(user.Name, err)
	}

	scaled, err := cmd.scaleProcess(app.GUID, user.Name)
	if err != nil {
		return err
	}
	if !scaled {
		return nil
	}

	warnings, err = cmd.Actor.PollStart(app.GUID)
	cmd.UI.DisplayWarnings(warnings)

	showErr := cmd.showCurrentScale(user.Name, err)
	if showErr != nil {
		return showErr
	}

	return cmd.translateErrors(err)
}

func (cmd ScaleCommand) translateErrors(err error) error {
	if _, ok := err.(actionerror.StartupTimeoutError); ok {
		return translatableerror.StartupTimeoutError{
			AppName:    cmd.RequiredArgs.AppName,
			BinaryName: cmd.Config.BinaryName(),
		}
	} else if _, ok := err.(actionerror.AllInstancesCrashedError); ok {
		return translatableerror.ApplicationUnableToStartError{
			AppName:    cmd.RequiredArgs.AppName,
			BinaryName: cmd.Config.BinaryName(),
		}
	}

	return err
}

func (cmd ScaleCommand) scaleProcess(appGUID string, username string) (bool, error) {
	cmd.UI.DisplayTextWithFlavor("Scaling app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})
	cmd.UI.DisplayNewline()

	shouldRestart := cmd.DiskLimit.IsSet || cmd.MemoryLimit.IsSet
	if shouldRestart && !cmd.Force {
		shouldScale, err := cmd.UI.DisplayBoolPrompt(
			false,
			"This will cause the app to restart. Are you sure you want to scale {{.AppName}}?",
			map[string]interface{}{"AppName": cmd.RequiredArgs.AppName})
		if err != nil {
			return false, err
		}

		if !shouldScale {
			cmd.UI.DisplayText("Scaling cancelled")
			return false, nil
		}
		cmd.UI.DisplayNewline()
	}

	warnings, err := cmd.Actor.ScaleProcessByApplication(appGUID, v7action.Process{
		Type:       cmd.ProcessType,
		Instances:  cmd.Instances.NullInt,
		MemoryInMB: cmd.MemoryLimit.NullUint64,
		DiskInMB:   cmd.DiskLimit.NullUint64,
	})
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return false, err
	}

	if shouldRestart {
		err := cmd.restartApplication(appGUID, username)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (cmd ScaleCommand) restartApplication(appGUID string, username string) error {
	cmd.UI.DisplayTextWithFlavor("Stopping app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})
	cmd.UI.DisplayNewline()

	warnings, err := cmd.Actor.StopApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Starting app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  username,
	})
	cmd.UI.DisplayNewline()

	_, warnings, err = cmd.Actor.StartApplication(appGUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	return nil
}

func (cmd ScaleCommand) showCurrentScale(userName string, runningErr error) error {
	if !shouldShowCurrentScale(runningErr) {
		return nil
	}

	cmd.UI.DisplayTextWithFlavor("Showing current scale of app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.RequiredArgs.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  userName,
	})

	summary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, false, cmd.RouteActor)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	appSummaryDisplayer := shared.NewAppSummaryDisplayer(cmd.UI)
	appSummaryDisplayer.AppDisplay(summary, false)
	return nil
}

func shouldShowCurrentScale(err error) bool {
	if err == nil {
		return true
	}

	if _, ok := err.(actionerror.AllInstancesCrashedError); ok {
		return true
	}

	return false
}
