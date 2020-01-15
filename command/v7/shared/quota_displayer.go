package shared

import (
	"fmt"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
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

func (displayer QuotaDisplayer) DisplayQuotasTable(orgQuotas []v7action.OrganizationQuota) {
	var keyValueTable = [][]string{
		{"name", "total memory", "instance memory", "routes", "service instances", "paid service plans", "app instances", "route ports"},
	}

	for _, orgQuota := range orgQuotas {
		paidServicesOutput := "disallowed"
		if orgQuota.Services.PaidServicePlans {
			paidServicesOutput = "allowed"
		}

		keyValueTable = append(keyValueTable, []string{
			orgQuota.Name,
			displayer.presentQuotaMemoryValue(orgQuota.Apps.TotalMemory),
			displayer.presentQuotaMemoryValue(orgQuota.Apps.InstanceMemory),
			displayer.presentQuotaValue(orgQuota.Routes.TotalRoutes),
			displayer.presentQuotaValue(orgQuota.Services.TotalServiceInstances),
			paidServicesOutput,
			displayer.presentQuotaValue(orgQuota.Apps.TotalAppInstances),
			displayer.presentQuotaValue(orgQuota.Routes.TotalRoutePorts),
		})
	}

	displayer.ui.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
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
