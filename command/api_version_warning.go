package command

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

type APIVersionTooHighError struct{}

func (a APIVersionTooHighError) Error() string {
	return ""
}

func WarnIfCLIVersionBelowAPIDefinedMinimum(config Config, apiVersion string, ui UI) error {
	minVer := config.MinCLIVersion()
	currentVer := config.BinaryVersion()

	isOutdated, err := CheckVersionOutdated(currentVer, minVer)
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
	isOutdated, err := CheckVersionOutdated(apiVersion, ccversion.MinSupportedV2ClientVersion)
	if err != nil {
		return err
	}

	if isOutdated {
		ui.DisplayWarning("Your CF API version ({{.APIVersion}}) is no longer supported. "+
			"Upgrade to a newer version of the API (minimum version {{.MinSupportedVersion}}). Please refer to "+
			"https://github.com/cloudfoundry/cli/wiki/Versioning-Policy#cf-cli-minimum-supported-version",
			map[string]interface{}{
				"APIVersion":          apiVersion,
				"MinSupportedVersion": ccversion.MinSupportedV2ClientVersion,
			})
	}

	return nil
}

func FailIfAPIVersionAboveMaxServiceProviderVersion(apiVersion string) error {
	isTooNew, err := checkVersionNewerThan(apiVersion, ccversion.MaxVersionServiceProviderV2)
	if err != nil {
		return err
	}

	if isTooNew {
		return APIVersionTooHighError{}
	}

	return nil
}

func checkVersionNewerThan(current, maximum string) (bool, error) {
	currentSemver, err := semver.Make(current)
	if err != nil {
		return false, err
	}

	maximumSemver, err := semver.Make(maximum)
	if err != nil {
		return false, err
	}

	if currentSemver.Compare(maximumSemver) == 1 {
		return true, nil
	}

	return false, nil
}

func CheckVersionOutdated(current string, minimum string) (bool, error) {
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
