package v2

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/bytefmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	oldCmd "code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/v2/shared"
)

//go:generate counterfeiter . AppActor

type AppActor interface {
	GetApplicationByNameAndSpace(name string, spaceGUID string) (v2action.Application, v2action.Warnings, error)
	GetApplicationSummaryByNameAndSpace(name string, spaceGUID string) (v2action.ApplicationSummary, v2action.Warnings, error)
}

type AppCommand struct {
	RequiredArgs    flag.AppName `positional-args:"yes"`
	GUID            bool         `long:"guid" description:"Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."`
	usage           interface{}  `usage:"CF_NAME app APP_NAME"`
	relatedCommands interface{}  `related_commands:"apps, events, logs, map-route, unmap-route, push"`

	Config      command.Config
	SharedActor SharedActor
	Actor       AppActor
	UI          command.UI
}

func (cmd *AppCommand) Setup(config command.Config, ui command.UI) error {
	cmd.Config = config
	cmd.UI = ui
	cmd.SharedActor = sharedaction.NewActor()

	ccClient, uaaClient, err := shared.NewClients(config, ui)
	if err != nil {
		return err
	}
	cmd.Actor = v2action.NewActor(ccClient, uaaClient)

	return nil
}

func (cmd AppCommand) Execute(args []string) error {
	if cmd.Config.Experimental() == false {
		oldCmd.Main(os.Getenv("CF_TRACE"), os.Args)
		return nil
	}

	cmd.UI.DisplayText(command.ExperimentalWarning)
	cmd.UI.DisplayNewline()

	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...",
		map[string]interface{}{
			"AppName":   cmd.RequiredArgs.AppName,
			"OrgName":   cmd.Config.TargetedOrganization().Name,
			"SpaceName": cmd.Config.TargetedSpace().Name,
			"Username":  user.Name,
		})
	cmd.UI.DisplayNewline()

	if cmd.GUID {
		return cmd.DisplayAppGUID()
	}

	appSummary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	err = ShowApp(appSummary, cmd.UI)
	if err != nil {
		return shared.HandleError(err)
	}

	return nil
}

func (cmd *AppCommand) DisplayAppGUID() error {
	app, warnings, err := cmd.Actor.GetApplicationByNameAndSpace(
		cmd.RequiredArgs.AppName,
		cmd.Config.TargetedSpace().GUID,
	)

	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayText(app.GUID)
	return nil
}

func ShowApp(appSummary v2action.ApplicationSummary, ui command.UI) error {
	// Application Summary Table
	instances := fmt.Sprintf("%d/%d", len(appSummary.RunningInstances), appSummary.Instances)

	usage := ui.TranslateText("{{.MemorySize}} x {{.NumInstances}} instances",
		map[string]interface{}{
			"MemorySize":   bytefmt.ByteSize(uint64(appSummary.Memory) * bytefmt.MEGABYTE),
			"NumInstances": appSummary.Instances,
		})

	formattedRoutes := []string{}
	for _, route := range appSummary.Routes {
		formattedRoutes = append(formattedRoutes, route.String())
	}
	routes := strings.Join(formattedRoutes, ", ")

	table := [][]string{
		{ui.TranslateText("Name:"), appSummary.Name},
		{ui.TranslateText("Instances:"), instances},
		{ui.TranslateText("Usage:"), usage},
		{ui.TranslateText("Routes:"), routes},
		{ui.TranslateText("Last uploaded:"), ui.UserFriendlyDate(appSummary.PackageUpdatedAt)},
		{ui.TranslateText("Stack:"), appSummary.Stack.Name},
		{ui.TranslateText("Buildpack:"), appSummary.Application.CalculatedBuildpack()},
	}

	ui.DisplayTable("", table, 3)
	ui.DisplayNewline()

	// Instance List Table
	table = [][]string{
		{"", "State", "Since", "CPU", "Memory", "Disk"},
	}

	for _, instance := range appSummary.RunningInstances {
		table = append(table,
			[]string{
				fmt.Sprintf("#%d", instance.ID),
				ui.TranslateText(strings.ToLower(string(instance.State))),
				ui.UserFriendlyDate(instance.StartTime()),
				fmt.Sprintf("%.1f%%", instance.CPU*100),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(uint64(instance.Memory)), bytefmt.ByteSize(uint64(instance.MemoryQuota))),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(uint64(instance.Disk)), bytefmt.ByteSize(uint64(instance.DiskQuota))),
			})
	}
	ui.DisplayTable("", table, 3)

	return nil
}
