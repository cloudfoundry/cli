package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
)

type AppSummaryDisplayer struct {
	UI         command.UI
	Config     command.Config
	Actor      V3AppSummaryActor
	V2AppActor V2AppActor
	AppName    string
}

//go:generate counterfeiter . V2AppActor

type V2AppActor interface {
	GetApplicationRoutes(appGUID string) (v2action.Routes, v2action.Warnings, error)
	GetApplicationInstancesWithStatsByApplication(guid string) ([]v2action.ApplicationInstanceWithStats, v2action.Warnings, error)
}

//go:generate counterfeiter . V3AppSummaryActor

type V3AppSummaryActor interface {
	GetApplicationSummaryByNameAndSpace(appName string, spaceGUID string, withObfuscatedValues bool) (v3action.ApplicationSummary, v3action.Warnings, error)
}

func (display AppSummaryDisplayer) DisplayAppInfo() error {
	summary, warnings, err := display.Actor.GetApplicationSummaryByNameAndSpace(display.AppName, display.Config.TargetedSpace().GUID, false)
	display.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	summary.ProcessSummaries.Sort()

	var routes v2action.Routes
	var appStats []v2action.ApplicationInstanceWithStats
	if len(summary.ProcessSummaries) > 0 {
		var routeWarnings v2action.Warnings
		routes, routeWarnings, err = display.V2AppActor.GetApplicationRoutes(summary.Application.GUID)
		display.UI.DisplayWarnings(routeWarnings)
		if _, ok := err.(ccerror.ResourceNotFoundError); err != nil && !ok {
			return err
		}

		if summary.State == constant.ApplicationStarted {
			var instanceWarnings v2action.Warnings
			appStats, instanceWarnings, err = display.V2AppActor.GetApplicationInstancesWithStatsByApplication(summary.Application.GUID)
			display.UI.DisplayWarnings(instanceWarnings)
			if _, ok := err.(ccerror.ResourceNotFoundError); err != nil && !ok {
				return err
			}
		}
	}

	display.displayAppTable(summary, routes, appStats)

	return nil
}

func (display AppSummaryDisplayer) DisplayAppProcessInfo() error {
	summary, warnings, err := display.Actor.GetApplicationSummaryByNameAndSpace(display.AppName, display.Config.TargetedSpace().GUID, false)
	display.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	summary.ProcessSummaries.Sort()

	display.displayProcessTable(summary)
	return nil
}

func GetCreatedTime(summary v3action.ApplicationSummary) time.Time {
	timestamp, _ := time.Parse(time.RFC3339, summary.CurrentDroplet.CreatedAt)
	return timestamp
}

func (display AppSummaryDisplayer) displayAppTable(summary v3action.ApplicationSummary, routes v2action.Routes, appStats []v2action.ApplicationInstanceWithStats) {
	var isoRow []string
	if len(appStats) > 0 && len(appStats[0].IsolationSegment) > 0 {
		isoRow = append(isoRow, display.UI.TranslateText("isolation segment:"), appStats[0].IsolationSegment)
	}

	var lifecycleInfo []string
	if summary.LifecycleType == constant.AppLifecycleTypeDocker {
		lifecycleInfo = []string{display.UI.TranslateText("docker image:"), summary.CurrentDroplet.Image}
	} else {
		lifecycleInfo = []string{display.UI.TranslateText("buildpacks:"), display.buildpackNames(summary.CurrentDroplet.Buildpacks)}
	}

	keyValueTable := [][]string{
		{display.UI.TranslateText("name:"), summary.Application.Name},
		{display.UI.TranslateText("requested state:"), strings.ToLower(string(summary.State))},
		isoRow,
		{display.UI.TranslateText("routes:"), routes.Summary()},
		{display.UI.TranslateText("last uploaded:"), display.UI.UserFriendlyDate(GetCreatedTime(summary))},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		lifecycleInfo,
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

	display.displayProcessTable(summary)
}

func (display AppSummaryDisplayer) displayAppInstancesTable(processSummary v3action.ProcessSummary) {
	display.UI.DisplayNewline()

	// TODO: figure out how to align key-value output
	keyValueTable := [][]string{
		{display.UI.TranslateText("type:"), processSummary.Type},
		{display.UI.TranslateText("instances:"), fmt.Sprintf("%d/%d", processSummary.HealthyInstanceCount(), processSummary.TotalInstanceCount())},
		{display.UI.TranslateText("memory usage:"), fmt.Sprintf("%dM", processSummary.MemoryInMB.Value)},
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

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
