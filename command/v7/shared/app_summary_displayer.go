package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
)

type AppSummaryDisplayer struct {
	UI command.UI
}

func NewAppSummaryDisplayer(ui command.UI) *AppSummaryDisplayer {
	return &AppSummaryDisplayer{
		UI: ui,
	}
}

func (display AppSummaryDisplayer) AppDisplay(summary v7action.ApplicationSummary, displayStartCommand bool) {
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
		{display.UI.TranslateText("last uploaded:"), display.getCreatedTime(summary)},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
		lifecycleInfo,
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

	display.displayProcessTable(summary, displayStartCommand)
}

func (display AppSummaryDisplayer) displayAppInstancesTable(processSummary v7action.ProcessSummary) {
	table := [][]string{
		{
			"",
			display.UI.TranslateText("state"),
			display.UI.TranslateText("since"),
			display.UI.TranslateText("cpu"),
			display.UI.TranslateText("memory"),
			display.UI.TranslateText("disk"),
			display.UI.TranslateText("details"),
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
			instance.Details,
		})
	}

	display.UI.DisplayInstancesTableForApp(table)
}

func (display AppSummaryDisplayer) displayProcessTable(summary v7action.ApplicationSummary, displayStartCommand bool) {
	for _, process := range summary.ProcessSummaries {
		display.UI.DisplayNewline()

		var startCommandRow []string
		if displayStartCommand && len(process.Command.Value) > 0 {
			startCommandRow = append(startCommandRow, display.UI.TranslateText("start command:"), process.Command.Value)
		}

		keyValueTable := [][]string{
			{display.UI.TranslateText("type:"), process.Type},
			{display.UI.TranslateText("instances:"), fmt.Sprintf("%d/%d", process.HealthyInstanceCount(), process.TotalInstanceCount())},
			{display.UI.TranslateText("memory usage:"), fmt.Sprintf("%dM", process.MemoryInMB.Value)},
			startCommandRow,
		}

		display.UI.DisplayKeyValueTable("", keyValueTable, 3)

		if len(process.InstanceDetails) == 0 {
			display.UI.DisplayNewline()
			display.UI.DisplayText("There are no running instances of this process.")
			continue
		}
		display.displayAppInstancesTable(process)
	}
}

func (display AppSummaryDisplayer) getCreatedTime(summary v7action.ApplicationSummary) string {
	if summary.CurrentDroplet.CreatedAt != "" {
		timestamp, _ := time.Parse(time.RFC3339, summary.CurrentDroplet.CreatedAt)
		return display.UI.UserFriendlyDate(timestamp)
	}

	return ""
}

func (AppSummaryDisplayer) usageSummary(processSummaries v7action.ProcessSummaries) string {
	var usageStrings []string
	for _, summary := range processSummaries {
		if summary.TotalInstanceCount() > 0 {
			usageStrings = append(usageStrings, fmt.Sprintf("%dM x %d", summary.MemoryInMB.Value, summary.TotalInstanceCount()))
		}
	}

	return strings.Join(usageStrings, ", ")
}

func (AppSummaryDisplayer) buildpackNames(buildpacks []v7action.DropletBuildpack) string {
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
	return input.UTC().Format(time.RFC3339)
}

func (AppSummaryDisplayer) processHasAnInstanceUp(processSummary *v7action.ProcessSummary) bool {
	for _, processInstance := range processSummary.InstanceDetails {
		if processInstance.State != constant.ProcessInstanceDown {
			return true
		}
	}
	return false
}

func (AppSummaryDisplayer) processInstancesAreAllCrashed(processSummary *v7action.ProcessSummary) bool {
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
