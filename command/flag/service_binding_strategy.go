package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/v9/resources"
	flags "github.com/jessevdk/go-flags"
)

type ServiceBindingStrategy struct {
	Strategy resources.BindingStrategyType
}

func (ServiceBindingStrategy) Complete(prefix string) []flags.Completion {
	return completions([]string{"single", "multiple"}, prefix, false)
}

func (h *ServiceBindingStrategy) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case "single", "multiple":
		h.Strategy = resources.BindingStrategyType(valLower)
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `STRATEGY must be "single" or "multiple"`,
		}
	}
	return nil
}
