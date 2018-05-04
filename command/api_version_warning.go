package command

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
)

func WarnAPIVersionCheck(config Config, ui UI) error {
	err := minimumCLIVersionCheck(config.BinaryVersion(), config.MinCLIVersion(), config.APIVersion())

	if e, ok := err.(translatableerror.MinimumCLIVersionNotMetError); ok {
		ui.DisplayWarning(e.Error(), map[string]interface{}{
			"APIVersion":    e.APIVersion,
			"MinCLIVersion": e.MinCLIVersion,
			"BinaryVersion": e.BinaryVersion,
		})
		ui.DisplayNewline()
		return nil
	}

	// Only error if there was an issue in parsing versions
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
