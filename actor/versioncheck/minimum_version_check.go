package versioncheck

import (
	"github.com/blang/semver"
)

func IsMinimumAPIVersionMet(current string, minimum string) (bool, error) {
	if minimum == "" {
		return true, nil
	}

	currentSemver, err := semver.Make(current)
	if err != nil {
		return false, err
	}

	minimumSemver, err := semver.Make(minimum)
	if err != nil {
		return false, err
	}

	if currentSemver.GTE(minimumSemver) {
		return true, nil
	}

	return false, nil
}
