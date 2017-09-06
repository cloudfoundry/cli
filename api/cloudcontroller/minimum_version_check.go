package cloudcontroller

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"github.com/blang/semver"
)

func MinimumAPIVersionCheck(current string, minimum string) error {
	if minimum == "" {
		return nil
	}

	currentSemvar, err := semver.Make(current)
	if err != nil {
		return err
	}

	minimumSemvar, err := semver.Make(minimum)
	if err != nil {
		return err
	}

	if currentSemvar.Compare(minimumSemvar) == -1 {
		return ccerror.MinimumAPIVersionNotMetError{
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
