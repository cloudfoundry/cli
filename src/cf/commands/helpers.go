package commands

import (
	term "cf/terminal"
	"fmt"
	"strings"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func byteSize(bytes int) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimRight(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

func coloredState(state string) (colored string) {
	switch state {
	case "started", "running":
		colored = term.SuccessColor("running")
	case "stopped":
		colored = term.StoppedColor("stopped")
	case "flapping":
		colored = term.WarningColor("flapping")
	case "starting":
		colored = term.AdvisoryColor("starting")
	default:
		colored = term.FailureColor(state)
	}

	return
}
