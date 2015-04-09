package utils

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"
)

func NotifyUpdateIfNeeded(ui terminal.UI, config core_config.Reader) {
	if !config.IsMinCliVersion(cf.Version) {
		ui.Say("")
		ui.Say(fmt.Sprintf("Cloud Foundry API version %s requires CLI version %s.  You are currently on version %s.  To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", config.ApiVersion(), config.MinRecommendedCliVersion(), cf.Version))
	}
}
