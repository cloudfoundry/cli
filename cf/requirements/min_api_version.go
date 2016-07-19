package requirements

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"github.com/blang/semver"

	. "code.cloudfoundry.org/cli/cf/i18n"
)

type MinAPIVersionRequirement struct {
	config          coreconfig.Reader
	feature         string
	requiredVersion semver.Version
}

func NewMinAPIVersionRequirement(
	config coreconfig.Reader,
	feature string,
	requiredVersion semver.Version,
) MinAPIVersionRequirement {
	return MinAPIVersionRequirement{
		config:          config,
		feature:         feature,
		requiredVersion: requiredVersion,
	}
}

func (r MinAPIVersionRequirement) Execute() error {
	if r.config.APIVersion() == "" {
		return errors.New(T("Unable to determine CC API Version. Please log in again."))
	}

	apiVersion, err := semver.Make(r.config.APIVersion())
	if err != nil {
		return errors.New(T("Unable to parse CC API Version '{{.APIVersion}}'", map[string]interface{}{
			"APIVersion": r.config.APIVersion(),
		}))
	}

	if apiVersion.LT(r.requiredVersion) {
		return errors.New(T(`{{.Feature}} requires CF API version {{.RequiredVersion}}+. Your target is {{.APIVersion}}.`,
			map[string]interface{}{
				"APIVersion":      r.config.APIVersion(),
				"Feature":         r.feature,
				"RequiredVersion": r.requiredVersion.String(),
			}))
	}

	return nil
}
