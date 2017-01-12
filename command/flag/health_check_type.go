package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type HealthCheckType struct {
	Port bool
	None bool
}

func (m *HealthCheckType) UnmarshalFlag(val string) error {
	if strings.EqualFold(val, "port") {
		m.Port = true
		return nil
	} else if strings.EqualFold(val, "none") {
		m.None = true
		return nil
	}
	return &flags.Error{
		Type:    flags.ErrRequired,
		Message: `HEALTH_CHECK_TYPE must be "port" or "none"`,
	}
}
