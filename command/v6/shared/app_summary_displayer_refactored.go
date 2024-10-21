package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/v7/actor/v2v3action"
	"code.cloudfoundry.org/cli/v7/actor/v3action"
	"code.cloudfoundry.org/cli/v7/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v7/command"
)

type AppSummaryDisplayer2 struct {
	UI command.UI
}

func NewAppSummaryDisplayer2(ui command.UI) *AppSummaryDisplayer2 {
	return &AppSummaryDisplayer2{
		UI: ui,
	}
}

func (display AppSummaryDisplayer2) AppDisplay(summary v2v3action.ApplicationSummary, displayStartCommand bool) {
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

	display.displayProcessTable(summary.ApplicationSummary, displayStartCommand)
}

func (display AppSummaryDisplayer2) displayAppInstancesTable(processSummary v3action.ProcessSummary) {
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

func (display AppSummaryDisplayer2) displayProcessTable(summary v3action.ApplicationSummary, displayStartCommand bool) {
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

func (display AppSummaryDisplayer2) getCreatedTime(summary v2v3action.ApplicationSummary) string {
	if summary.CurrentDroplet.CreatedAt != "" {
		timestamp, _ := time.Parse(time.RFC3339, summary.CurrentDroplet.CreatedAt)
		return display.UI.UserFriendlyDate(timestamp)
	}

	return ""
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
	return input.UTC().Format(time.RFC3339)
}
