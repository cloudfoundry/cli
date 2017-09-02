package version

import (
	"code.cloudfoundry.org/cli/command/translatableerror"
	"github.com/blang/semver"
)

const (
	MinVersionLifecyleStagingV2         = "2.68.0"
	MinVersionHTTPEndpointHealthCheckV2 = "2.68.0"
	MinVersionProcessHealthCheckV2      = "2.47.0"

	MinVersionHTTPRoutePath                 = "2.36.0"
	MinVersionTCPRouting                    = "2.53.0"
	MinVersionNoHostInReservedRouteEndpoint = "2.55.0"

	MinVersionV3                 = "3.27.0"
	MinVersionRunTaskV3          = "3.0.0"
	MinVersionIsolationSegmentV3 = "3.11.0"
)

func MinimumAPIVersionCheck(current string, minimum string, customCommand ...string) error {
	if current == DefaultVersion || minimum == "" {
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

	var command string
	if len(customCommand) > 0 {
		command = customCommand[0]
	}

	if currentSemvar.Compare(minimumSemvar) == -1 {
		return translatableerror.MinimumAPIVersionNotMetError{
			Command:        command,
			CurrentVersion: current,
			MinimumVersion: minimum,
		}
	}

	return nil
}
