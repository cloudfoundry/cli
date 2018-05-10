package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type NetworkProtocol struct {
	Protocol string
}

func (NetworkProtocol) Complete(prefix string) []flags.Completion {
	return completions([]string{"tcp", "udp"}, prefix, false)
}

func (h *NetworkProtocol) UnmarshalFlag(val string) error {
	valLower := strings.ToLower(val)
	switch valLower {
	case "tcp", "udp":
		h.Protocol = valLower
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `PROTOCOL must be "tcp" or "udp"`,
		}
	}
	return nil
}
