package flag

import (
	"fmt"
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
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --protocol, but got option %s", val),
		}
	}
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
