package command

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

func WarnIfCLIVersionBelowAPIDefinedMinimum(config Config, apiVersion string, ui UI) error {
	minVer := config.MinCLIVersion()
	currentVer := config.BinaryVersion()

	isOutdated, err := checkVersionOutdated(currentVer, minVer)
	if err != nil {
		return err
	}

	if isOutdated {
		ui.DisplayWarning("Cloud Foundry API version {{.APIVersion}} requires CLI version {{.MinCLIVersion}}. You are currently on version {{.BinaryVersion}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"APIVersion":    apiVersion,
				"MinCLIVersion": minVer,
				"BinaryVersion": currentVer,
			})
		ui.DisplayNewline()
	}

	return nil
}

func WarnIfAPIVersionBelowSupportedMinimum(apiVersion string, ui UI) error {
	isOutdated, err := checkVersionOutdated(apiVersion, ccversion.MinV2ClientVersion)
	if err != nil {
		return err
	}

	if isOutdated {
		ui.DisplayWarning("Your API version is no longer supported. Upgrade to a newer version of the API. Please refer to https://github.com/cloudfoundry/cli/wiki/Versioning-Policy#cf-cli-minimum-supported-version")
	}

	return nil
}

func checkVersionOutdated(current string, minimum string) (bool, error) {
	if current == version.DefaultVersion || minimum == "" {
		return false, nil
	}

	currentSemver, err := semver.Make(current)
	if err != nil {
		return false, err
	}

	minimumSemver, err := semver.Make(minimum)
	if err != nil {
		return false, err
	}

	if currentSemver.Compare(minimumSemver) == -1 {
		return true, nil
	}

	return false, nil
}
