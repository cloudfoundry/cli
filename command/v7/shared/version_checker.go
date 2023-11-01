package shared

import (
	"fmt"

	"github.com/blang/semver/v4"
)

const minimumCCAPIVersionForV7 = "3.85.0"

func CheckCCAPIVersion(currentAPIVersion string) (string, error) {
	currentSemver, err := semver.Make(currentAPIVersion)
	if err != nil {
		return "", err
	}

	minimumSemver, err := semver.Make(minimumCCAPIVersionForV7)
	if err != nil {
		return "", err
	}

	if currentSemver.LT(minimumSemver) {
		return fmt.Sprintf("\nWarning: Your targeted API's version (%s) is less than the minimum supported API version (%s). Some commands may not function correctly.", currentAPIVersion, minimumCCAPIVersionForV7), nil
	}

	return "", nil
}
