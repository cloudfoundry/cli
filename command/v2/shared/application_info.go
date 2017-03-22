package shared

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command"
	"github.com/cloudfoundry/bytefmt"
)

// DisplayAppSummary displays the application summary to the UI, and optionally
// the command to start the app.
func DisplayAppSummary(ui command.UI, appSummary v2action.ApplicationSummary, displayStartCommand bool) {
	instances := fmt.Sprintf("%d/%d", len(appSummary.RunningInstances), appSummary.Instances)

	usage := ui.TranslateText(
		"{{.MemorySize}} x {{.NumInstances}} instances",
		map[string]interface{}{
			"MemorySize":   bytefmt.ByteSize(uint64(appSummary.Memory) * bytefmt.MEGABYTE),
			"NumInstances": appSummary.Instances,
		})

	formattedRoutes := []string{}
	for _, route := range appSummary.Routes {
		formattedRoutes = append(formattedRoutes, route.String())
	}
	routes := strings.Join(formattedRoutes, ", ")

	table := [][]string{
		{ui.TranslateText("name:"), appSummary.Name},
		{ui.TranslateText("requested state:"), strings.ToLower(string(appSummary.State))},
		{ui.TranslateText("instances:"), instances},
		{ui.TranslateText("usage:"), usage},
		{ui.TranslateText("routes:"), routes},
		{ui.TranslateText("last uploaded:"), ui.UserFriendlyDate(appSummary.PackageUpdatedAt)},
		{ui.TranslateText("stack:"), appSummary.Stack.Name},
		{ui.TranslateText("buildpack:"), appSummary.Application.CalculatedBuildpack()},
	}

	if displayStartCommand {
		table = append(table, []string{ui.TranslateText("start command:"), appSummary.Application.DetectedStartCommand})
	}

	ui.DisplayTable("", table, 3)
	ui.DisplayNewline()

	if len(appSummary.RunningInstances) == 0 {
		ui.DisplayText("There are no running instances of this app.")
	} else {
		displayAppInstances(ui, appSummary.RunningInstances)
	}
}

func displayAppInstances(ui command.UI, instances []v2action.ApplicationInstanceWithStats) {
	table := [][]string{
		{"", "State", "Since", "CPU", "Memory", "Disk", "Details"},
	}

	for _, instance := range instances {
		table = append(
			table,
			[]string{
				fmt.Sprintf("#%d", instance.ID),
				ui.TranslateText(strings.ToLower(string(instance.State))),
				zuluDate(instance.TimeSinceCreation()),
				fmt.Sprintf("%.1f%%", instance.CPU*100),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(uint64(instance.Memory)), bytefmt.ByteSize(uint64(instance.MemoryQuota))),
				fmt.Sprintf("%s of %s", bytefmt.ByteSize(uint64(instance.Disk)), bytefmt.ByteSize(uint64(instance.DiskQuota))),
				instance.Details,
			})
	}

	ui.DisplayNonWrappingTable("", table, 3)
}

// zuluDate converts the time to UTC and then formats it to ISO8601.
func zuluDate(input time.Time) string {
	// "2006-01-02T15:04:05Z07:00"
	return input.UTC().Format(time.RFC3339)
}
