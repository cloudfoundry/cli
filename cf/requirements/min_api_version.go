package requirements

import (
	"errors"

	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"

	. "github.com/cloudfoundry/cli/cf/i18n"
)

type MinAPIVersionRequirement struct {
	config          core_config.Reader
	feature         string
	requiredVersion semver.Version
}

func NewMinAPIVersionRequirement(
	config core_config.Reader,
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
	if r.config.ApiVersion() == "" {
		return errors.New(T("Unable to determine CC API Version. Please log in again."))
	}

	apiVersion, err := semver.Make(r.config.ApiVersion())
	if err != nil {
		return errors.New(T("Unable to parse CC API Version '{{.APIVersion}}'", map[string]interface{}{
			"APIVersion": r.config.ApiVersion(),
		}))
	}

	if apiVersion.LT(r.requiredVersion) {
		return errors.New(T(`{{.Feature}} requires CF API version {{.RequiredVersion}}+. Your target is {{.ApiVersion}}.`,
			map[string]interface{}{
				"ApiVersion":      r.config.ApiVersion(),
				"Feature":         r.feature,
				"RequiredVersion": r.requiredVersion.String(),
			}))
	}

	return nil
}
