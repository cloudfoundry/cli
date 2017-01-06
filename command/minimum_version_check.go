package command

import "github.com/blang/semver"

func MinimumAPIVersionCheck(current string, minimum string) error {
	if current == "BUILT_FROM_SOURCE" || minimum == "" {
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
		return MinimumAPIVersionNotMetError{
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
