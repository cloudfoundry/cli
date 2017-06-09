package v3

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cloudfoundry/bytefmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/v3/shared"
)

//go:generate counterfeiter . V3AppActor

type V3AppActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string) (v3action.ApplicationSummary, v3action.Warnings, error)
}

type V3AppCommand struct {
	usage   interface{} `usage:"CF_NAME v3-app -n APP_NAME"`
	AppName string      `short:"n" long:"name" description:"The application name to display" required:"true"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       V3AppActor
}

func (cmd *V3AppCommand) Setup(config command.Config, ui command.UI) error {
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

func (cmd V3AppCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(cmd.Config, true, true)
	if err != nil {
		return shared.HandleError(err)
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return shared.HandleError(err)
	}

	cmd.UI.DisplayTextWithFlavor("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   cmd.AppName,
		"OrgName":   cmd.Config.TargetedOrganization().Name,
		"SpaceName": cmd.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	cmd.UI.DisplayNewline()

	summary, warnings, err := cmd.Actor.GetApplicationSummaryByNameAndSpace(cmd.AppName, cmd.Config.TargetedSpace().GUID)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return shared.HandleError(err)
	}

	displayAppTable(cmd.UI, summary)

	return nil
}

// Sort processes alphabetically and put web first.
func displayAppTable(ui command.UI, summary v3action.ApplicationSummary) {
	sort.Slice(summary.Processes, func(i int, j int) bool {
		var iScore int
		var jScore int

		switch summary.Processes[i].Type {
		case "web":
			iScore = 0
		default:
			iScore = 1
		}

		switch summary.Processes[j].Type {
		case "web":
			jScore = 0
		default:
			jScore = 1
		}

		if iScore == 1 && jScore == 1 {
			return summary.Processes[i].Type < summary.Processes[j].Type
		}
		return iScore < jScore
	})

	keyValueTable := [][]string{
		{ui.TranslateText("name:"), summary.Application.Name},
		{ui.TranslateText("requested state:"), summary.State},
		{ui.TranslateText("processes:"), processesSummary(summary.Processes)},
		{ui.TranslateText("memory usage:"), usageSummary(summary.Processes)},
		{ui.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		{ui.TranslateText("buildpacks:"), buildpackNames(summary.CurrentDroplet.Buildpacks)},
	}

	ui.DisplayKeyValueTable("", keyValueTable, 3)

	appHasARunningInstance := false

	for processIdx := range summary.Processes {
		if processHasAnInstance(&summary.Processes[processIdx]) {
			appHasARunningInstance = true
			break
		}
	}

	if !appHasARunningInstance {
		ui.DisplayText("There are no running instances of this app.")
		return
	}

	for _, process := range summary.Processes {
		ui.DisplayNewline()

		ui.DisplayTextWithBold("{{.ProcessType}}", map[string]interface{}{
			"ProcessType": process.Type,
		})

		if !processHasAnInstance(&process) {
			ui.DisplayText("There are no running instances of this process.")
			continue
		}

		table := [][]string{
			{
				"",
				ui.TranslateText("state"),
				ui.TranslateText("since"),
				ui.TranslateText("cpu"),
				ui.TranslateText("memory"),
				ui.TranslateText("disk"),
			},
		}

		for _, instance := range process.Instances {
			table = append(table, []string{
				fmt.Sprintf("#%d", instance.Index),
				ui.TranslateText(strings.ToLower(string(instance.State))),
				appInstanceDate(instance.StartTime()),
				fmt.Sprintf("%.1f%%", instance.CPU*100),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(instance.MemoryUsage), bytefmt.ByteSize(instance.MemoryQuota)),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(instance.DiskUsage), bytefmt.ByteSize(instance.DiskQuota)),
			})
		}

		ui.DisplayInstancesTableForApp(table)
	}
}

func processesSummary(processes []v3action.Process) string {
	var processesStrings []string
	for _, process := range processes {
		processesStrings = append(processesStrings, fmt.Sprintf("%s:%d/%d", process.Type, process.HealthyInstanceCount(), process.TotalInstanceCount()))
	}

	return strings.Join(processesStrings, ", ")
}

func usageSummary(processes []v3action.Process) string {
	var usageStrings []string
	for _, process := range processes {
		if process.TotalInstanceCount() > 0 {
			usageStrings = append(usageStrings, fmt.Sprintf("%dM x %d", process.MemoryInMB, process.TotalInstanceCount()))
		}
	}

	return strings.Join(usageStrings, ", ")
}

func buildpackNames(buildpacks []v3action.Buildpack) string {
	var names []string
	for _, buildpack := range buildpacks {
		if buildpack.DetectOutput != "" {
			names = append(names, buildpack.DetectOutput)
		} else {
			names = append(names, buildpack.Name)
		}
	}

	return strings.Join(names, ", ")
}

func appInstanceDate(input time.Time) string {
	return input.Local().Format("2006-01-02 15:04:05 PM")
}

func processHasAnInstance(process *v3action.Process) bool {
	for instanceIdx := range process.Instances {
		if process.Instances[instanceIdx].State != "DOWN" {
			return true
		}
	}

	return false
}
