package net

import (
	"os"
	"strings"

	"github.com/cloudfoundry/cli/cf/terminal"
)

type WarningsCollector struct {
	ui                terminal.UI
	warning_producers []WarningProducer
}

type WarningProducer interface {
	Warnings() []string
}

func NewWarningsCollector(ui terminal.UI, warning_producers ...WarningProducer) (warnings_collector WarningsCollector) {
	warnings_collector.ui = ui
	warnings_collector.warning_producers = warning_producers
	return
}

func (warnings_collector WarningsCollector) PrintWarnings() {
	if os.Getenv("CF_RAISE_ERROR_ON_WARNINGS") != "" {
		warnings := []string{}
		for _, warning_producer := range warnings_collector.warning_producers {
			for _, warning := range warning_producer.Warnings() {
				warnings = append(warnings, warning)
			}
		}

		if len(warnings) > 0 {
			panic(strings.Join(warnings, "\n"))
		}
	}
}
