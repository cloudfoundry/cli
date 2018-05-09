package flag

import (
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type RoutePath struct {
	Path string
}

func (h *RoutePath) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --route-path, but got option %s", val),
		}
	}

	if !strings.HasPrefix(val, "/") {
		h.Path = fmt.Sprintf("/%s", val)
	} else {
		h.Path = val
	}
	return nil
}
