package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/ui"
	log "github.com/sirupsen/logrus"
)

type AppSummaryDisplayer struct {
	UI command.UI
}

func NewAppSummaryDisplayer(ui command.UI) *AppSummaryDisplayer {
	return &AppSummaryDisplayer{
		UI: ui,
	}
}

func (display AppSummaryDisplayer) AppDisplay(summary v7action.DetailedApplicationSummary, displayStartCommand bool) {
	var isoRow []string
	var keyValueTable [][]string
	if name, exists := summary.GetIsolationSegmentName(); exists {
		isoRow = append(isoRow, display.UI.TranslateText("isolation segment:"), name)
	}

	if summary.LifecycleType == constant.AppLifecycleTypeDocker {
		keyValueTable = [][]string{
			{display.UI.TranslateText("name:"), summary.Application.Name},
			{display.UI.TranslateText("requested state:"), strings.ToLower(string(summary.State))},
			isoRow,
			{display.UI.TranslateText("routes:"), routeSummary(summary.Routes)},
			{display.UI.TranslateText("last uploaded:"), display.getCreatedTime(summary)},
			{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
			{display.UI.TranslateText("docker image:"), summary.CurrentDroplet.Image},
			isoRow,
		}
	} else {
		keyValueTable = [][]string{
			{display.UI.TranslateText("name:"), summary.Application.Name},
			{display.UI.TranslateText("requested state:"), strings.ToLower(string(summary.State))},
			isoRow,
			{display.UI.TranslateText("routes:"), routeSummary(summary.Routes)},
			{display.UI.TranslateText("last uploaded:"), display.getCreatedTime(summary)},
			{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
			{display.UI.TranslateText("buildpacks:"), ""},
			isoRow,
		}
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, 3)

	if summary.LifecycleType == constant.AppLifecycleTypeBuildpack {
		display.displayBuildpackTable(summary.CurrentDroplet.Buildpacks)
	}

	display.displayProcessTable(summary, displayStartCommand)
}

func routeSummary(rs []resources.Route) string {
	formattedRoutes := []string{}
	for _, route := range rs {
		formattedRoutes = append(formattedRoutes, route.URL)
	}
	return strings.Join(formattedRoutes, ", ")
}

func formatLogRateLimit(limit int64) string {
	if limit == -1 {
		return "unlimited"
	} else {
		return bytefmt.ByteSize(uint64(limit)) + "/s"
	}
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
			display.UI.TranslateText("logging"),
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
			display.UI.TranslateText("{{.LogRate}}/s of {{.LogRateLimit}}", map[string]interface{}{
				"LogRate":      bytefmt.ByteSize(instance.LogRate),
				"LogRateLimit": formatLogRateLimit(instance.LogRateLimit),
			}),
			instance.Details,
		})
	}

	display.UI.DisplayInstancesTableForApp(table)
}

func (display AppSummaryDisplayer) displayProcessTable(summary v7action.DetailedApplicationSummary, displayStartCommand bool) {
	for _, process := range summary.ProcessSummaries {
		display.UI.DisplayNewline()

		var startCommandRow []string
		if displayStartCommand && len(process.Command.Value) > 0 {
			startCommandRow = append(startCommandRow, display.UI.TranslateText("start command:"), process.Command.Value)
		}

		var processSidecars []string
		for _, sidecar := range process.Sidecars {
			processSidecars = append(processSidecars, sidecar.Name)
		}

		keyValueTable := [][]string{
			{display.UI.TranslateText("type:"), process.Type},
			{display.UI.TranslateText("sidecars:"), strings.Join(processSidecars, ", ")},
			{display.UI.TranslateText("instances:"), fmt.Sprintf("%d/%d", process.HealthyInstanceCount(), process.TotalInstanceCount())},
			{display.UI.TranslateText("memory usage:"), fmt.Sprintf("%dM", process.MemoryInMB.Value)},
			startCommandRow,
		}

		display.UI.DisplayKeyValueTable("", keyValueTable, 3)

		if len(process.InstanceDetails) == 0 {
			display.UI.DisplayText("There are no running instances of this process.")
			continue
		}
		display.displayAppInstancesTable(process)
	}
}

func (display AppSummaryDisplayer) getCreatedTime(summary v7action.DetailedApplicationSummary) string {
	if summary.CurrentDroplet.CreatedAt != "" {
		timestamp, err := time.Parse(time.RFC3339, summary.CurrentDroplet.CreatedAt)
		if err != nil {
			log.WithField("createdAt", summary.CurrentDroplet.CreatedAt).Errorln("error parsing created at:", err)
		}

		return display.UI.UserFriendlyDate(timestamp)
	}

	return ""
}

func (AppSummaryDisplayer) appInstanceDate(input time.Time) string {
	return input.UTC().Format(time.RFC3339)
}

func (display AppSummaryDisplayer) displayBuildpackTable(buildpacks []resources.DropletBuildpack) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{
				display.UI.TranslateText("name"),
				display.UI.TranslateText("version"),
				display.UI.TranslateText("detect output"),
				display.UI.TranslateText("buildpack name"),
			},
		}

		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{
				buildpack.Name,
				buildpack.Version,
				buildpack.DetectOutput,
				buildpack.BuildpackName,
			})
		}

		display.UI.DisplayTableWithHeader("\t", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
