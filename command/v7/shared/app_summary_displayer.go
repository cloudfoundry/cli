package shared

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/command"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/types"
	"code.cloudfoundry.org/cli/v8/util/ui"
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

	keyValueTable = [][]string{
		{display.UI.TranslateText("name:"), summary.Application.Name},
		{display.UI.TranslateText("requested state:"), strings.ToLower(string(summary.State))},
		isoRow,
		{display.UI.TranslateText("routes:"), routeSummary(summary.Routes)},
		{display.UI.TranslateText("last uploaded:"), display.getCreatedTime(summary)},
		{display.UI.TranslateText("stack:"), summary.CurrentDroplet.Stack},
	}

	if summary.LifecycleType == constant.AppLifecycleTypeDocker {
		keyValueTable = append(keyValueTable, []string{display.UI.TranslateText("docker image:"), summary.CurrentDroplet.Image}, isoRow)
	} else {
		keyValueTable = append(keyValueTable, []string{display.UI.TranslateText("buildpacks:"), ""}, isoRow)
	}

	display.UI.DisplayKeyValueTable("", keyValueTable, ui.DefaultTableSpacePadding)

	if summary.LifecycleType != constant.AppLifecycleTypeDocker {
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

func formatCPUEntitlement(cpuEntitlement types.NullFloat64) string {
	if !cpuEntitlement.IsSet {
		return ""
	}
	return fmt.Sprintf("%.1f%%", cpuEntitlement.Value*100)
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
			display.UI.TranslateText("cpu entitlement"),
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
			formatCPUEntitlement(instance.CPUEntitlement),
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

		display.UI.DisplayKeyValueTable("", keyValueTable, ui.DefaultTableSpacePadding)

		if len(process.InstanceDetails) == 0 {
			display.UI.DisplayText("There are no running instances of this process.")
			continue
		}
		display.displayAppInstancesTable(process)
	}

	if summary.Deployment.StatusValue == constant.DeploymentStatusValueActive {
		display.UI.DisplayNewline()
		display.UI.DisplayText(display.getDeploymentStatusText(summary))

		var maxInFlightRow []string
		var maxInFlight = summary.Deployment.Options.MaxInFlight
		if maxInFlight > 0 {
			maxInFlightRow = append(maxInFlightRow, display.UI.TranslateText("max-in-flight:"), strconv.Itoa(maxInFlight))
		}
		var canaryStepsRow []string
		if summary.Deployment.CanaryStatus.Steps.TotalSteps > 0 {
			stepStatus := summary.Deployment.CanaryStatus.Steps
			canaryStepsRow = []string{display.UI.TranslateText("canary-steps:"), fmt.Sprintf("%d/%d", stepStatus.CurrentStep, stepStatus.TotalSteps)}

		}

		keyValueTable := [][]string{
			{display.UI.TranslateText("strategy:"), strings.ToLower(string(summary.Deployment.Strategy))},
			maxInFlightRow,
			canaryStepsRow,
		}

		display.UI.DisplayKeyValueTable("", keyValueTable, ui.DefaultTableSpacePadding)

		if summary.Deployment.Strategy == constant.DeploymentStrategyCanary && summary.Deployment.StatusReason == constant.DeploymentStatusReasonPaused {
			display.UI.DisplayNewline()
			display.UI.DisplayText(fmt.Sprintf("Please run `cf continue-deployment %s` to promote the canary deployment, or `cf cancel-deployment %s` to rollback to the previous version.", summary.Application.Name, summary.Application.Name))
		}
	}
}

func (display AppSummaryDisplayer) getDeploymentStatusText(summary v7action.DetailedApplicationSummary) string {
	var lastStatusChangeTime = display.getLastStatusChangeTime(summary)
	if lastStatusChangeTime != "" {
		return fmt.Sprintf("Active deployment with status %s (since %s)",
			summary.Deployment.StatusReason,
			lastStatusChangeTime)
	} else {
		return fmt.Sprintf("Active deployment with status %s.",
			summary.Deployment.StatusReason)
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

func (display AppSummaryDisplayer) getLastStatusChangeTime(summary v7action.DetailedApplicationSummary) string {
	if summary.Deployment.LastStatusChange != "" {
		timestamp, err := time.Parse(time.RFC3339, summary.Deployment.LastStatusChange)
		if err != nil {
			log.WithField("last_status_change", summary.Deployment.LastStatusChange).Errorln("error parsing last status change:", err)
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
