package command

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v2/constant"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

func WarnCLIVersionCheck(config Config, ui UI) error {
	err := minimumCLIVersionCheck(config.BinaryVersion(), config.MinCLIVersion(), config.APIVersion())
	if err != nil {
		if e, ok := err.(translatableerror.MinimumCLIVersionNotMetError); ok {
			ui.DisplayWarning(e.Error(), map[string]interface{}{
				"APIVersion":    e.APIVersion,
				"MinCLIVersion": e.MinCLIVersion,
				"BinaryVersion": e.BinaryVersion,
			})
			ui.DisplayNewline()
			return nil
		}
	}

	return err
}

func WarnAPIVersionCheck(apiVersion string, ui UI) error {
	err := MinimumAPIVersionCheck(apiVersion, constant.MinimumAPIVersion)
	if _, ok := err.(translatableerror.MinimumAPIVersionNotMetError); ok {
		ui.DisplayWarning("Your API version is no longer supported. Upgrade to a newer version of the API.")
		return nil
	}

	return err
}

func minimumCLIVersionCheck(current string, minimum string, apiVersion string) error {
	comparison, err := compareSemVer(current, minimum)
	if err != nil {
		return err
	}

	if comparison == -1 {
		return translatableerror.MinimumCLIVersionNotMetError{
			APIVersion:    apiVersion,
			BinaryVersion: current,
			MinCLIVersion: minimum,
		}
	}

	return nil
}

func compareSemVer(current string, minimum string) (int, error) {
	if current == version.DefaultVersion || minimum == "" {
		return 2, nil
	}

	currentSemvar, err := semver.Make(current)
	if err != nil {
		return 2, err
	}

	minimumSemvar, err := semver.Make(minimum)
	if err != nil {
		return 2, err
	}

	return currentSemvar.Compare(minimumSemvar), nil
}
