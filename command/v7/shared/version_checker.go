package shared

import (
	"fmt"

	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"github.com/blang/semver/v4"
)

func CheckCCAPIVersion(currentAPIVersion string) (string, error) {
	currentSemver, err := semver.Make(currentAPIVersion)
	if err != nil {
		return "", err
	}

	minimumSemver, err := semver.Make(ccversion.MinSupportedClientVersionV9)
	if err != nil {
		return "", err
	}

	if currentSemver.LT(minimumSemver) {
		return fmt.Sprintf("\nWarning: Your targeted API's version (%s) is less than the minimum supported API version (%s). Some commands may not function correctly.", currentAPIVersion, ccversion.MinSupportedClientVersionV9), nil
	}

	return "", nil
}
