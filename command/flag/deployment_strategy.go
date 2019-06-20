package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/jessevdk/go-flags"
)

type DeploymentStrategy struct {
	Name constant.DeploymentStrategy
}

func (DeploymentStrategy) Complete(prefix string) []flags.Completion {
	return completions([]string{string(constant.DeploymentStrategyRolling)}, prefix, false)
}

func (h *DeploymentStrategy) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)

	switch valLower {

	case string(constant.DeploymentStrategyDefault):
	case string(constant.DeploymentStrategyRolling):
		h.Name = constant.DeploymentStrategy(valLower)

	default:
		return &flags.Error{
			Type:    flags.ErrInvalidChoice,
			Message: `STRATEGY must be "rolling" or not set`,
		}
	}

	return nil
}
