package flag

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type HealthCheckType struct {
	Type constant.HealthCheckType
}

func (HealthCheckType) Complete(prefix string) []flags.Completion {
	return completions([]string{"http", "port", "process"}, prefix, false)
}

func (h *HealthCheckType) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case "port", "process", "http":
		h.Type = constant.HealthCheckType(valLower)
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
		}
	}
	return nil
}
