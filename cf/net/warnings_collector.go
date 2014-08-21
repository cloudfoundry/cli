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
	warnings := []string{}
	for _, warning_producer := range warnings_collector.warning_producers {
		for _, warning := range warning_producer.Warnings() {
			warnings = append(warnings, warning)
		}
	}

	if os.Getenv("CF_RAISE_ERROR_ON_WARNINGS") != "" {
		if len(warnings) > 0 {
			panic(strings.Join(warnings, "\n"))
		}
	}

	warnings = warnings_collector.removeDuplicates(warnings)

	for _, warning := range warnings {
		warnings_collector.ui.Warn(warning)
	}
}

func (warnings_collector WarningsCollector) removeDuplicates(stringArray []string) []string {
	length := len(stringArray) - 1
	for i := 0; i < length; i++ {
		for j := i + 1; j <= length; j++ {
			if stringArray[i] == stringArray[j] {
				stringArray[j] = stringArray[length]
				stringArray = stringArray[0:length]
				length--
				j--
			}
		}
	}
	return stringArray
}
