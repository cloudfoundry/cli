package requirements

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type MinAPIVersionRequirement struct {
	ui              terminal.UI
	config          core_config.Reader
	feature         string
	requiredVersion semver.Version
}

func NewMinAPIVersionRequirement(
	ui terminal.UI,
	config core_config.Reader,
	feature string,
	requiredVersion semver.Version,
) MinAPIVersionRequirement {
	return MinAPIVersionRequirement{
		ui:              ui,
		config:          config,
		feature:         feature,
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

		return false
	}

	if apiVersion.LT(r.requiredVersion) {
		r.ui.Failed(T(`{{.Feature}} requires CF API version {{.RequiredVersion}}+. Your target is {{.ApiVersion}}.`,
			map[string]interface{}{
				"ApiVersion":      r.config.ApiVersion(),
				"Feature":         r.feature,
				"RequiredVersion": r.requiredVersion.String(),
			}))

		return false
	}

	return true
}
