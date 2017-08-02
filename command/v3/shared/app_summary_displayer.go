package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	sharedV2 "code.cloudfoundry.org/cli/command/v2/shared"
	"github.com/cloudfoundry/bytefmt"
)

type AppSummaryDisplayer struct {
	UI              command.UI
	Config          command.Config
	Actor           V3AppSummaryActor
	V2AppRouteActor V2AppRouteActor
	AppName         string
}

//go:generate counterfeiter . V2AppRouteActor

type V2AppRouteActor interface {
	GetApplicationRoutes(appGUID string) (v2action.Routes, v2action.Warnings, error)
}

//go:generate counterfeiter . V3AppSummaryActor

type V3AppSummaryActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string) (v3action.ApplicationSummary, v3action.Warnings, error)
}

func (display AppSummaryDisplayer) DisplayAppInfo() error {
	user, err := display.Config.CurrentUser()
	if err != nil {
		return HandleError(err)
	}

	display.UI.DisplayTextWithFlavor("Showing health and status for app {{.AppName}} in org {{.OrgName}} / space {{.SpaceName}} as {{.Username}}...", map[string]interface{}{
		"AppName":   display.AppName,
		"OrgName":   display.Config.TargetedOrganization().Name,
		"SpaceName": display.Config.TargetedSpace().Name,
		"Username":  user.Name,
	})
	display.UI.DisplayNewline()

	summary, warnings, err := display.Actor.GetApplicationSummaryByNameAndSpace(display.AppName, display.Config.TargetedSpace().GUID)
	display.UI.DisplayWarnings(warnings)
	if err != nil {
		return HandleError(err)
	}

	var routes v2action.Routes
	if len(summary.Processes) > 0 {
		var routeWarnings v2action.Warnings
		routes, routeWarnings, err = display.V2AppRouteActor.GetApplicationRoutes(summary.Application.GUID)
		display.UI.DisplayWarnings(routeWarnings)
		if err != nil {
			return sharedV2.HandleError(err)
		}
	}

	display.displayAppTable(summary, routes)

	return nil
}

// Sort processes alphabetically and put web first.
func (display AppSummaryDisplayer) displayAppTable(summary v3action.ApplicationSummary, routes v2action.Routes) {
	summary.Processes.Sort()

	keyValueTable := [][]string{
		{display.UI.TranslateText("name:"), summary.Application.Name},
		{display.UI.TranslateText("requested state:"), strings.ToLower(summary.State)},
		{display.UI.TranslateText("processes:"), display.processesSummary(summary.Processes)},
		{display.UI.TranslateText("memory usage:"), display.usageSummary(summary.Processes)},
		{display.UI.TranslateText("routes:"), routes.Summary()},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		{display.UI.TranslateText("buildpacks:"), display.buildpackNames(summary.CurrentDroplet.Buildpacks)},
	}

	crashedProcesses := []string{}
	for i := range summary.Processes {
		if display.processInstancesAreAllCrashed(&summary.Processes[i]) {
			crashedProcesses = append(crashedProcesses, summary.Processes[i].Type)
		}
	}

	display.UI.DisplayKeyValueTableForV3App(keyValueTable, crashedProcesses)

	appHasARunningInstance := false

	for processIdx := range summary.Processes {
		if display.processHasAnInstance(&summary.Processes[processIdx]) {
			appHasARunningInstance = true
			break
		}
	}

	if !appHasARunningInstance {
		display.UI.DisplayNewline()
		display.UI.DisplayText("There are no running instances of this app.")
		return
	}

	for _, process := range summary.Processes {
		display.DisplayAppInstancesTable(process)
	}
}

func (display AppSummaryDisplayer) DisplayAppInstancesTable(process v3action.Process) {
	display.UI.DisplayNewline()

	display.UI.DisplayTextWithBold("{{.ProcessType}}:{{.HealthyInstanceCount}}/{{.TotalInstanceCount}}", map[string]interface{}{
		"ProcessType":          process.Type,
		"HealthyInstanceCount": process.HealthyInstanceCount(),
		"TotalInstanceCount":   process.TotalInstanceCount(),
	})

	if !display.processHasAnInstance(&process) {
		return
	}

	table := [][]string{
		{
			"",
			display.UI.TranslateText("state"),
			display.UI.TranslateText("since"),
			display.UI.TranslateText("cpu"),
			display.UI.TranslateText("memory"),
			display.UI.TranslateText("disk"),
		},
	}

	for _, instance := range process.Instances {
		table = append(table, []string{
			fmt.Sprintf("#%d", instance.Index),
			display.UI.TranslateText(strings.ToLower(string(instance.State))),
			display.appInstanceDate(instance.StartTime()),
			fmt.Sprintf("%.1f%%", instance.CPU*100),
			display.UI.TranslateText("{{.MemUsage}} of {{.MemQuota}}", map[string]interface{}{
				"MemUsage": bytefmt.ByteSize(instance.MemoryUsage),
				"MemQuota": bytefmt.ByteSize(instance.MemoryQuota),
			}),
			display.UI.TranslateText("{{.DiskUsage}} of {{.DiskQuota}}", map[string]interface{}{
				"DiskUsage": bytefmt.ByteSize(instance.DiskUsage),
				"DiskQuota": bytefmt.ByteSize(instance.DiskQuota),
			}),
		})
	}

	display.UI.DisplayInstancesTableForApp(table)
}

func (AppSummaryDisplayer) processesSummary(processes []v3action.Process) string {
	var processesStrings []string
	for _, process := range processes {
		processesStrings = append(processesStrings, fmt.Sprintf("%s:%d/%d", process.Type, process.HealthyInstanceCount(), process.TotalInstanceCount()))
	}

	return strings.Join(processesStrings, ", ")
}

func (AppSummaryDisplayer) usageSummary(processes []v3action.Process) string {
	var usageStrings []string
	for _, process := range processes {
		if process.TotalInstanceCount() > 0 {
			usageStrings = append(usageStrings, fmt.Sprintf("%dM x %d", process.MemoryInMB, process.TotalInstanceCount()))
		}
	}

	return strings.Join(usageStrings, ", ")
}

func (AppSummaryDisplayer) buildpackNames(buildpacks []v3action.Buildpack) string {
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

func (AppSummaryDisplayer) appInstanceDate(input time.Time) string {
	return input.Local().Format("2006-01-02 15:04:05 PM")
}

func (AppSummaryDisplayer) processHasAnInstance(process *v3action.Process) bool {
	for instanceIdx := range process.Instances {
		if process.Instances[instanceIdx].State != "DOWN" {
			return true
		}
	}

	return false
}

func (AppSummaryDisplayer) processInstancesAreAllCrashed(process *v3action.Process) bool {
	if len(process.Instances) < 1 {
		return false
	}

	for instanceIdx := range process.Instances {
		if process.Instances[instanceIdx].State != "CRASHED" {
			return false
		}
	}

	return true
}
