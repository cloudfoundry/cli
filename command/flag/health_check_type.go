package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type HealthCheckType struct {
	Type string
}

func (m *HealthCheckType) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case "port", "process", "http":
		m.Type = valLower
		return nil
	case "none": // Deprecated
		m.Type = "process"
		return nil
	}

	return &flags.Error{
		Type:    flags.ErrRequired,
		Message: `HEALTH_CHECK_TYPE must be "port", "process", or "http"`,
	}
}
