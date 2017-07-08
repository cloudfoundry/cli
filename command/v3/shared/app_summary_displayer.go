package shared

import (
	"fmt"
	"sort"
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
	GetApplicationRoutes(appGUID string) ([]v2action.Route, v2action.Warnings, error)
}

//go:generate counterfeiter . V3AppSummaryActor

type V3AppSummaryActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string) (v3action.ApplicationSummary, v3action.Warnings, error)
}

func (cmd AppSummaryDisplayer) DisplayAppInfo() error {
	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return HandleError(err)
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
		return HandleError(err)
	}

	routes, routeWarnings, err := cmd.V2AppRouteActor.GetApplicationRoutes(summary.Application.GUID)
	cmd.UI.DisplayWarnings(routeWarnings)
	if err != nil {
		return sharedV2.HandleError(err)
	}

	cmd.displayAppTable(summary, routes)

	return nil
}

// Sort processes alphabetically and put web first.
func (cmd AppSummaryDisplayer) displayAppTable(summary v3action.ApplicationSummary, routes []v2action.Route) {
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
		{cmd.UI.TranslateText("name:"), summary.Application.Name},
		{cmd.UI.TranslateText("requested state:"), strings.ToLower(summary.State)},
		{cmd.UI.TranslateText("processes:"), cmd.processesSummary(summary.Processes)},
		{cmd.UI.TranslateText("memory usage:"), cmd.usageSummary(summary.Processes)},
		{cmd.UI.TranslateText("routes:"), cmd.routesSummary(routes)},
		{cmd.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		{cmd.UI.TranslateText("buildpacks:"), cmd.buildpackNames(summary.CurrentDroplet.Buildpacks)},
	}

	crashedProcesses := []string{}
	for i := range summary.Processes {
		if cmd.processInstancesAreAllCrashed(&summary.Processes[i]) {
			crashedProcesses = append(crashedProcesses, summary.Processes[i].Type)
		}
	}

	cmd.UI.DisplayKeyValueTableForV3App(keyValueTable, crashedProcesses)

	appHasARunningInstance := false

	for processIdx := range summary.Processes {
		if cmd.processHasAnInstance(&summary.Processes[processIdx]) {
			appHasARunningInstance = true
			break
		}
	}

	if !appHasARunningInstance {
		cmd.UI.DisplayNewline()
		cmd.UI.DisplayText("There are no running instances of this app.")
		return
	}

	for _, process := range summary.Processes {
		cmd.UI.DisplayNewline()

		cmd.UI.DisplayTextWithBold("{{.ProcessType}}:{{.HealthyInstanceCount}}/{{.TotalInstanceCount}}", map[string]interface{}{
			"ProcessType":          process.Type,
			"HealthyInstanceCount": process.HealthyInstanceCount(),
			"TotalInstanceCount":   process.TotalInstanceCount(),
		})

		if !cmd.processHasAnInstance(&process) {
			continue
		}

		table := [][]string{
			{
				"",
				cmd.UI.TranslateText("state"),
				cmd.UI.TranslateText("since"),
				cmd.UI.TranslateText("cpu"),
				cmd.UI.TranslateText("memory"),
				cmd.UI.TranslateText("disk"),
			},
		}

		for _, instance := range process.Instances {
			table = append(table, []string{
				fmt.Sprintf("#%d", instance.Index),
				cmd.UI.TranslateText(strings.ToLower(string(instance.State))),
				cmd.appInstanceDate(instance.StartTime()),
				fmt.Sprintf("%.1f%%", instance.CPU*100),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(instance.MemoryUsage), bytefmt.ByteSize(instance.MemoryQuota)),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(instance.DiskUsage), bytefmt.ByteSize(instance.DiskQuota)),
			})
		}

		cmd.UI.DisplayInstancesTableForApp(table)
	}
}

func (AppSummaryDisplayer) processesSummary(processes []v3action.Process) string {
	var processesStrings []string
	for _, process := range processes {
		processesStrings = append(processesStrings, fmt.Sprintf("%s:%d/%d", process.Type, process.HealthyInstanceCount(), process.TotalInstanceCount()))
	}

	return strings.Join(processesStrings, ", ")
}

func (AppSummaryDisplayer) routesSummary(routes []v2action.Route) string {
	formattedRoutes := []string{}
	for _, route := range routes {
		formattedRoutes = append(formattedRoutes, route.String())
	}
	return strings.Join(formattedRoutes, ", ")
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
