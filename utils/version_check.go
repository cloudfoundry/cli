package utils

import (
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

func NotifyUpdateIfNeeded(ui terminal.UI, config core_config.Reader) {
	if !config.IsMinCliVersion(cf.Version) {
		ui.Say("")
		ui.Say(T("Cloud Foundry API version {{.ApiVer}} requires CLI version {{.CliMin}}.  You are currently on version {{.CliVer}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"ApiVer": config.ApiVersion(),
				"CliMin": config.MinCliVersion(),
				"CliVer": cf.Version,
			}))
	}
}
