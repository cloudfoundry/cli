package requirements

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type MinAPIVersionRequirement struct {
	ui              terminal.UI
	config          core_config.Reader
	commandName     string
	requiredVersion semver.Version
}

func NewMinAPIVersionRequirement(
	ui terminal.UI,
	config core_config.Reader,
	commandName string,
	requiredVersion semver.Version,
) MinAPIVersionRequirement {
	return MinAPIVersionRequirement{
		ui:              ui,
		config:          config,
		commandName:     commandName,
		requiredVersion: requiredVersion,
	}
}

func (r MinAPIVersionRequirement) Execute() bool {
	if r.config.ApiVersion() == "" {
		r.ui.Failed(T("Unable to determine CC API Version. Please log in again."))
	}

	apiVersion, err := semver.Make(r.config.ApiVersion())
	if err != nil {
		r.ui.Failed(T("Unable to parse CC API Version '{{.APIVersion}}'", map[string]interface{}{
			"APIVersion": r.config.ApiVersion(),
		}))
	}

	if apiVersion.LT(r.requiredVersion) {
		r.ui.Failed(T(`Current CF CLI version {{.Version}}
	Current CF API version {{.ApiVersion}}
	To use the {{.CommandName}} feature, you need to upgrade the CF API to at least {{.RequiredVersion}}`,
			map[string]interface{}{
				"Version":         cf.Version,
				"ApiVersion":      r.config.ApiVersion(),
				"CommandName":     r.commandName,
				"RequiredVersion": r.requiredVersion.String(),
			}))
	}

	return true
}
