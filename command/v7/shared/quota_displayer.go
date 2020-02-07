package shared

import (
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/ui"
)

const (
	MEGABYTE = 1024 * 1024
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

type QuotaDisplayer struct {
	ui command.UI
}

func NewQuotaDisplayer(ui command.UI) QuotaDisplayer {
	return QuotaDisplayer{ui: ui}
}

func (displayer QuotaDisplayer) DisplayQuotasTable(quotas []resources.Quota, emptyMessage string) {
	if len(quotas) == 0 {
		displayer.ui.DisplayText(emptyMessage)
		return
	}

	var keyValueTable = [][]string{
		{"name", "total memory", "instance memory", "routes", "service instances", "paid service plans", "app instances", "route ports"},
	}

	for _, quota := range quotas {
		keyValueTable = append(keyValueTable, []string{
			quota.Name,
			displayer.presentQuotaMemoryValue(*quota.Apps.TotalMemory),
			displayer.presentQuotaMemoryValue(*quota.Apps.InstanceMemory),
			displayer.presentQuotaValue(*quota.Routes.TotalRoutes),
			displayer.presentQuotaValue(*quota.Services.TotalServiceInstances),
			displayer.presentBooleanValue(*quota.Services.PaidServicePlans),
			displayer.presentQuotaValue(*quota.Apps.TotalAppInstances),
			displayer.presentQuotaValue(*quota.Routes.TotalReservedPorts),
		})
	}

	displayer.ui.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
}

func (displayer QuotaDisplayer) DisplaySingleQuota(quota resources.Quota) {
	quotaTable := [][]string{
		{displayer.ui.TranslateText("total memory:"), displayer.presentQuotaMemoryValue(*quota.Apps.TotalMemory)},
		{displayer.ui.TranslateText("instance memory:"), displayer.presentQuotaMemoryValue(*quota.Apps.InstanceMemory)},
		{displayer.ui.TranslateText("routes:"), displayer.presentQuotaValue(*quota.Routes.TotalRoutes)},
		{displayer.ui.TranslateText("service instances:"), displayer.presentQuotaValue(*quota.Services.TotalServiceInstances)},
		{displayer.ui.TranslateText("paid service plans:"), displayer.presentBooleanValue(*quota.Services.PaidServicePlans)},
		{displayer.ui.TranslateText("app instances:"), displayer.presentQuotaValue(*quota.Apps.TotalAppInstances)},
		{displayer.ui.TranslateText("route ports:"), displayer.presentQuotaValue(*quota.Routes.TotalReservedPorts)},
	}

	displayer.ui.DisplayKeyValueTable("", quotaTable, 3)
}

func (displayer QuotaDisplayer) presentBooleanValue(limit bool) string {
	if limit {
		return "allowed"
	} else {
		return "disallowed"
	}
}

func (displayer QuotaDisplayer) presentQuotaValue(limit types.NullInt) string {
	if !limit.IsSet {
		return "unlimited"
	} else {
		return strconv.Itoa(limit.Value)
	}
}

func (displayer QuotaDisplayer) presentQuotaMemoryValue(limit types.NullInt) string {
	if !limit.IsSet {
		return "unlimited"
	} else {
		return addMemoryUnits(float64(limit.Value) * MEGABYTE)
	}
}

func addMemoryUnits(bytes float64) string {
	unit := ""
	value := bytes

	switch {
	case bytes >= TERABYTE:
		unit = "T"
		value /= TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value /= GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value /= MEGABYTE
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}
