package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
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
	summary, warnings, err := display.Actor.GetApplicationSummaryByNameAndSpace(display.AppName, display.Config.TargetedSpace().GUID)
	display.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	summary.ProcessSummaries.Sort()

	var routes v2action.Routes
	if len(summary.ProcessSummaries) > 0 {
		var routeWarnings v2action.Warnings
		routes, routeWarnings, err = display.V2AppRouteActor.GetApplicationRoutes(summary.Application.GUID)
		display.UI.DisplayWarnings(routeWarnings)
		if err != nil {
			return err
		}
	}

	display.displayAppTable(summary, routes)

	return nil
}

func (display AppSummaryDisplayer) displayAppInstancesTable(processSummary v3action.ProcessSummary) {
	display.UI.DisplayNewline()

	display.UI.DisplayTextWithBold("{{.ProcessType}}:{{.HealthyInstanceCount}}/{{.TotalInstanceCount}}", map[string]interface{}{
		"ProcessType":          processSummary.Type,
		"HealthyInstanceCount": processSummary.HealthyInstanceCount(),
		"TotalInstanceCount":   processSummary.TotalInstanceCount(),
	})

	if !display.processHasAnInstance(&processSummary) {
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

	for _, instance := range processSummary.InstanceDetails {
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

func (display AppSummaryDisplayer) DisplayAppProcessInfo() error {
	summary, warnings, err := display.Actor.GetApplicationSummaryByNameAndSpace(display.AppName, display.Config.TargetedSpace().GUID)
	display.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	summary.ProcessSummaries.Sort()

	display.displayProcessTable(summary)
	return nil
}

func (display AppSummaryDisplayer) displayAppTable(summary v3action.ApplicationSummary, routes v2action.Routes) {
	keyValueTable := [][]string{
		{display.UI.TranslateText("name:"), summary.Application.Name},
		{display.UI.TranslateText("requested state:"), strings.ToLower(string(summary.State))},
		{display.UI.TranslateText("processes:"), summary.ProcessSummaries.String()},
		{display.UI.TranslateText("memory usage:"), display.usageSummary(summary.ProcessSummaries)},
		{display.UI.TranslateText("routes:"), routes.Summary()},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
	}

	var lifecycleInfo []string

	if summary.Lifecycle.Type == constant.DockerAppLifecycleType {
		lifecycleInfo = []string{display.UI.TranslateText("docker image:"), summary.CurrentDroplet.Image}
	} else {
		lifecycleInfo = []string{display.UI.TranslateText("buildpacks:"), display.buildpackNames(summary.CurrentDroplet.Buildpacks)}
	}

	keyValueTable = append(keyValueTable, lifecycleInfo)

	crashedProcesses := []string{}
	for i := range summary.ProcessSummaries {
		if display.processInstancesAreAllCrashed(&summary.ProcessSummaries[i]) {
			crashedProcesses = append(crashedProcesses, summary.ProcessSummaries[i].Type)
		}
	}

	display.UI.DisplayKeyValueTableForV3App(keyValueTable, crashedProcesses)

	display.displayProcessTable(summary)
}

func (display AppSummaryDisplayer) displayProcessTable(summary v3action.ApplicationSummary) {
	appHasARunningInstance := false

	for processIdx := range summary.ProcessSummaries {
		if display.processHasAnInstance(&summary.ProcessSummaries[processIdx]) {
			appHasARunningInstance = true
			break
		}
	}

	if !appHasARunningInstance {
		display.UI.DisplayNewline()
		display.UI.DisplayText("There are no running instances of this app.")
		return
	}

	for _, process := range summary.ProcessSummaries {
		display.displayAppInstancesTable(process)
	}
}

func (AppSummaryDisplayer) usageSummary(processSummaries v3action.ProcessSummaries) string {
	var usageStrings []string
	for _, summary := range processSummaries {
		if summary.TotalInstanceCount() > 0 {
			usageStrings = append(usageStrings, fmt.Sprintf("%dM x %d", summary.MemoryInMB.Value, summary.TotalInstanceCount()))
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

func (AppSummaryDisplayer) processHasAnInstance(processSummary *v3action.ProcessSummary) bool {
	for instanceIdx := range processSummary.InstanceDetails {
		if processSummary.InstanceDetails[instanceIdx].State != constant.ProcessInstanceDown {
			return true
		}
	}

	return false
}

func (AppSummaryDisplayer) processInstancesAreAllCrashed(processSummary *v3action.ProcessSummary) bool {
	if len(processSummary.InstanceDetails) < 1 {
		return false
	}

	for instanceIdx := range processSummary.InstanceDetails {
		if processSummary.InstanceDetails[instanceIdx].State != constant.ProcessInstanceCrashed {
			return false
		}
	}

	return true
}
