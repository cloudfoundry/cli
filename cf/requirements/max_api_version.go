package requirements

import (
	"errors"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type MaxAPIVersionRequirement struct {
	config         core_config.Reader
	feature        string
	maximumVersion semver.Version
}

func NewMaxAPIVersionRequirement(
	config core_config.Reader,
	feature string,
	maximumVersion semver.Version,
) MaxAPIVersionRequirement {
	return MaxAPIVersionRequirement{
		config:         config,
		feature:        feature,
		maximumVersion: maximumVersion,
	}
}

func (r MaxAPIVersionRequirement) Execute() error {
	if r.config.ApiVersion() == "" {
		return errors.New(T("Unable to determine CC API Version. Please log in again."))
	}

	apiVersion, err := semver.Make(r.config.ApiVersion())
	if err != nil {
		return errors.New(T("Unable to parse CC API Version '{{.APIVersion}}'", map[string]interface{}{
			"APIVersion": r.config.ApiVersion(),
		}))
	}

	if apiVersion.GT(r.maximumVersion) {
		return errors.New(T(`{{.Feature}} only works up to CF API version {{.MaximumVersion}}. Your target is {{.ApiVersion}}.`,
			map[string]interface{}{
				"ApiVersion":     r.config.ApiVersion(),
				"Feature":        r.feature,
				"MaximumVersion": r.maximumVersion.String(),
			}))
	}

	return nil
}
