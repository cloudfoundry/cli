package net

import (
	"github.com/cloudfoundry/cli/cf/terminal"
)

type warningsCollector struct {
	ui                terminal.UI
	warning_producers []WarningProducer
}

type WarningProducer interface {
	Warnings() []string
}

func NewWarningsCollector(ui terminal.UI, warning_producers ...WarningProducer) (warnings_collector warningsCollector) {
	warnings_collector.ui = ui
	warnings_collector.warning_producers = warning_producers
	return
}

func (warnings_collector warningsCollector) PrintWarnings() {
	for _, warning_producer := range warnings_collector.warning_producers {
		for _, warning := range warning_producer.Warnings() {
			warnings_collector.ui.Warn(warning)
		}
	}
}
