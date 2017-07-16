package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type HealthCheckType struct {
	Type string
}

func (HealthCheckType) Complete(prefix string) []flags.Completion {
	return completions([]string{"http", "port", "process"}, prefix, false)
}

func (h *HealthCheckType) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case "port", "process", "http", "none":
		h.Type = valLower
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
		}
	}
	return nil
}
