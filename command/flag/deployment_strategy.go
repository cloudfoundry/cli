package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"github.com/jessevdk/go-flags"
)

type DeploymentStrategy struct {
	Name constant.DeploymentStrategy
}

func (DeploymentStrategy) Complete(prefix string) []flags.Completion {
	return completions([]string{string(constant.DeploymentStrategyRolling), string(constant.DeploymentStrategyCanary)}, prefix, false)
}

func (h *DeploymentStrategy) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)

	switch valLower {

	case string(constant.DeploymentStrategyDefault):
		// Do nothing, leave the default value

	case string(constant.DeploymentStrategyRolling),
		string(constant.DeploymentStrategyCanary):
		h.Name = constant.DeploymentStrategy(valLower)

	default:
		return &flags.Error{
			Type:    flags.ErrInvalidChoice,
			Message: `STRATEGY must be "canary", "rolling" or not set`,
		}
	}

	return nil
}
