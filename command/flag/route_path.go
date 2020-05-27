package flag

import (
	"fmt"
	"strings"
)

type V7RoutePath struct {
	Path string
}

func (h *V7RoutePath) UnmarshalFlag(val string) error {
	if val != "" && !strings.HasPrefix(val, "/") {
		h.Path = fmt.Sprintf("/%s", val)
	} else {
		h.Path = val
	}

	return nil
}
