package v7

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	"code.cloudfoundry.org/cli/v9/command/v7/shared"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/types"
)

type ScaleCommand struct {
	BaseCommand

	RequiredArgs        flag.AppName            `positional-args:"yes"`
	Force               bool                    `long:"force" short:"f" description:"Force restart of app without prompt"`
	Instances           flag.Instances          `long:"instances" short:"i" required:"false" description:"Number of instances"`
	DiskLimit           flag.Megabytes          `short:"k" required:"false" description:"Disk limit (e.g. 256M, 1024M, 1G)"`
	LogRateLimit        flag.BytesWithUnlimited `short:"l" required:"false" description:"Log rate limit per second, in bytes (e.g. 128B, 4K, 1M). -l=-1 represents unlimited"`
	MemoryLimit         flag.Megabytes          `short:"m" required:"false" description:"Memory limit (e.g. 256M, 1024M, 1G)"`
	ProcessType         string                  `long:"process" default:"web" description:"App process to scale"`
	usage               interface{}             `usage:"CF_NAME scale APP_NAME [--process PROCESS] [-i INSTANCES] [-k DISK] [-m MEMORY] [-l LOG_RATE_LIMIT] [-f]\n\n   Modifying the app's disk, memory, or log rate will cause the app to restart."`
	relatedCommands     interface{}             `related_commands:"push"`
	envCFStartupTimeout interface{}             `environmentName:"CF_STARTUP_TIMEOUT" environmentDescription:"Max wait time for app instance startup, in minutes" environmentDefault:"5"`
}

func (cmd ScaleCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(true, true)
	if err != nil {
		return err
	}

	user, err := cmd.Actor.GetCurrentUser()
	if err != nil {
		return err
	}

	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	if !cmd.Instances.IsSet && !cmd.DiskLimit.IsSet && !cmd.MemoryLimit.IsSet && !cmd.LogRateLimit.IsSet {
		return cmd.showCurrentScale(user.Name, err)
	}

	scaled, err := cmd.scaleProcess(app.GUID, user.Name)
	if err != nil {
		return err
	}
	if !scaled {
		return nil
	}

	handleInstanceDetails := func(instanceDetails string) {
		cmd.UI.DisplayText(instanceDetails)
	}

	if cmd.shouldRestart() || app.State == constant.ApplicationStarted {
		warnings, err = cmd.Actor.PollStart(app, false, handleInstanceDetails)
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayWarnings(warnings)
	}

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

	shouldRestart := cmd.shouldRestart()
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

	warnings, err := cmd.Actor.ScaleProcessByApplication(appGUID, resources.Process{
		Type:              cmd.ProcessType,
		Instances:         cmd.Instances.NullInt,
		MemoryInMB:        cmd.MemoryLimit.NullUint64,
		DiskInMB:          cmd.DiskLimit.NullUint64,
		LogRateLimitInBPS: types.NullInt(cmd.LogRateLimit),
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

	warnings, err = cmd.Actor.StartApplication(appGUID)
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

	cmd.UI.DisplayNewline()

	summary, warnings, err := cmd.Actor.GetDetailedAppSummary(cmd.RequiredArgs.AppName, cmd.Config.TargetedSpace().GUID, false)
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

func (cmd ScaleCommand) shouldRestart() bool {
	return cmd.DiskLimit.IsSet || cmd.MemoryLimit.IsSet || cmd.LogRateLimit.IsSet
}
