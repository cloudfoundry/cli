package command

import (
	"code.cloudfoundry.org/cli/command/v2/constant"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

func WarnCLIVersionCheck(config Config, ui UI) error {
	minVer := config.MinCLIVersion()
	currentVer := config.BinaryVersion()

	isOutdated, err := checkVersionOutdated(currentVer, minVer)
	if err != nil {
		return err
	}

	if isOutdated {
		ui.DisplayWarning("Cloud Foundry API version {{.APIVersion}} requires CLI version {{.MinCLIVersion}}. You are currently on version {{.BinaryVersion}}. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads",
			map[string]interface{}{
				"APIVersion":    config.APIVersion(),
				"MinCLIVersion": minVer,
				"BinaryVersion": currentVer,
			})
		ui.DisplayNewline()
	}

	return nil
}

func WarnAPIVersionCheck(apiVersion string, ui UI) error {
	isOutdated, err := checkVersionOutdated(apiVersion, constant.MinimumAPIVersion)
	if err != nil {
		return err
	}

	if isOutdated {
		ui.DisplayWarning("Your API version is no longer supported. Upgrade to a newer version of the API.")
	}

	return nil
}

func checkVersionOutdated(current string, minimum string) (bool, error) {
	if current == version.DefaultVersion || minimum == "" {
		return false, nil
	}

	currentSemvar, err := semver.Make(current)
	if err != nil {
		return false, err
	}

	minimumSemvar, err := semver.Make(minimum)
	if err != nil {
		return false, err
	}

	if currentSemvar.Compare(minimumSemvar) == -1 {
		return true, nil
	}

	return false, nil
}
