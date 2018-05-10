package flag

import (
	"fmt"
	"strings"
)

type RoutePath struct {
	Path string
}

func (h *RoutePath) UnmarshalFlag(val string) error {
	if !strings.HasPrefix(val, "/") {
		h.Path = fmt.Sprintf("/%s", val)
	} else {
		h.Path = val
	}
	return nil
}
