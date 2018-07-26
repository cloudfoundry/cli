package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/v2v3action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
)

type AppSummaryDisplayer2 struct {
	UI command.UI
}

func NewAppSummaryDisplayer2(ui command.UI) *AppSummaryDisplayer2 {
	return &AppSummaryDisplayer2{
		UI: ui,
	}
}

func (display AppSummaryDisplayer2) GetCreatedTime(summary v2v3action.ApplicationSummary) string {
	if summary.CurrentDroplet.CreatedAt != "" {
		timestamp, _ := time.Parse(time.RFC3339, summary.CurrentDroplet.CreatedAt)
		return display.UI.UserFriendlyDate(timestamp)
	}

	return ""
}

func (display AppSummaryDisplayer2) AppDisplay(summary v2v3action.ApplicationSummary) {
	var isoRow []string
	if name, exists := summary.GetIsolationSegmentName(); exists {
		isoRow = append(isoRow, display.UI.TranslateText("isolation segment:"), name)
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
		{display.UI.TranslateText("routes:"), summary.Routes.Summary()},
		{display.UI.TranslateText("last uploaded:"), display.GetCreatedTime(summary)},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		lifecycleInfo,
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

	display.displayProcessTable(summary.ApplicationSummary)
}

func (display AppSummaryDisplayer2) displayAppInstancesTable(processSummary v3action.ProcessSummary) {
	display.UI.DisplayNewline()

	keyValueTable := [][]string{
		{display.UI.TranslateText("type:"), processSummary.Type},
		{display.UI.TranslateText("instances:"), fmt.Sprintf("%d/%d", processSummary.HealthyInstanceCount(), processSummary.TotalInstanceCount())},
		{display.UI.TranslateText("memory usage:"), fmt.Sprintf("%dM", processSummary.MemoryInMB.Value)},
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

	if !display.processHasAnInstanceUp(&processSummary) {
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

func (display AppSummaryDisplayer2) displayProcessTable(summary v3action.ApplicationSummary) {
	for _, process := range summary.ProcessSummaries {
		display.displayAppInstancesTable(process)

		if !display.processHasAnInstanceUp(&process) || len(process.InstanceDetails) == 0 {
			display.UI.DisplayNewline()
			display.UI.DisplayText("There are no running instances of this process.")
		}
	}
}

func (AppSummaryDisplayer2) usageSummary(processSummaries v3action.ProcessSummaries) string {
	var usageStrings []string
	for _, summary := range processSummaries {
		if summary.TotalInstanceCount() > 0 {
			usageStrings = append(usageStrings, fmt.Sprintf("%dM x %d", summary.MemoryInMB.Value, summary.TotalInstanceCount()))
		}
	}

	return strings.Join(usageStrings, ", ")
}

func (AppSummaryDisplayer2) buildpackNames(buildpacks []v3action.Buildpack) string {
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

func (AppSummaryDisplayer2) appInstanceDate(input time.Time) string {
	return input.Local().Format("2006-01-02 15:04:05 PM")
}

func (AppSummaryDisplayer2) processHasAnInstanceUp(processSummary *v3action.ProcessSummary) bool {
	for _, processInstance := range processSummary.InstanceDetails {
		if processInstance.State != constant.ProcessInstanceDown {
			return true
		}
	}
	return false
}

func (AppSummaryDisplayer2) processInstancesAreAllCrashed(processSummary *v3action.ProcessSummary) bool {
	if len(processSummary.InstanceDetails) < 1 {
		return false
	}

	for _, processInstance := range processSummary.InstanceDetails {
		if processInstance.State != constant.ProcessInstanceDown {
			return false
		}
	}
	return true
}
