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
	if valLower == "none" || valLower == "port" {
		m.Type = valLower
		return nil
	}

	return &flags.Error{
		Type:    flags.ErrRequired,
		Message: `HEALTH_CHECK_TYPE must be "port" or "none"`,
	}
}
