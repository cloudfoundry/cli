package requirements

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/terminal"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type MaxAPIVersionRequirement struct {
	ui             terminal.UI
	config         core_config.Reader
	feature        string
	maximumVersion semver.Version
}

func NewMaxAPIVersionRequirement(
	ui terminal.UI,
	config core_config.Reader,
	feature string,
	maximumVersion semver.Version,
) MaxAPIVersionRequirement {
	return MaxAPIVersionRequirement{
		ui:             ui,
		config:         config,
		feature:        feature,
		maximumVersion: maximumVersion,
	}
}

func (r MaxAPIVersionRequirement) Execute() bool {
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

	if apiVersion.GT(r.maximumVersion) {
		r.ui.Failed(T(`{{.Feature}} only works up to CF API version {{.MaximumVersion}}. Your target is {{.ApiVersion}}.`,
			map[string]interface{}{
				"ApiVersion":     r.config.ApiVersion(),
				"Feature":        r.feature,
				"MaximumVersion": r.maximumVersion.String(),
			}))

		return false
	}

	return true
}
