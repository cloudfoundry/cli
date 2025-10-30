package flag

import (
	"strings"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"

	flags "github.com/jessevdk/go-flags"
)

type HealthCheckType struct {
	Type constant.HealthCheckType
}

func (HealthCheckType) Complete(prefix string) []flags.Completion {
	return completions([]string{string(constant.HTTP), string(constant.Port), string(constant.Process)}, prefix, false)
}

func (h *HealthCheckType) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case string(constant.HTTP), string(constant.Port), string(constant.Process):
		h.Type = constant.HealthCheckType(valLower)
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
		}
	}
	return nil
}
