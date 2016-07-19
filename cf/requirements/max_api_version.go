package requirements

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"github.com/blang/semver"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type MaxAPIVersionRequirement struct {
	config         coreconfig.Reader
	feature        string
	maximumVersion semver.Version
}

func NewMaxAPIVersionRequirement(
	config coreconfig.Reader,
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
	if r.config.APIVersion() == "" {
		return errors.New(T("Unable to determine CC API Version. Please log in again."))
	}

	apiVersion, err := semver.Make(r.config.APIVersion())
	if err != nil {
		return errors.New(T("Unable to parse CC API Version '{{.APIVersion}}'", map[string]interface{}{
			"APIVersion": r.config.APIVersion(),
		}))
	}

	if apiVersion.GT(r.maximumVersion) {
		return errors.New(T(`{{.Feature}} only works up to CF API version {{.MaximumVersion}}. Your target is {{.APIVersion}}.`,
			map[string]interface{}{
				"APIVersion":     r.config.APIVersion(),
				"Feature":        r.feature,
				"MaximumVersion": r.maximumVersion.String(),
			}))
	}

	return nil
}
