package flag

import (
	"fmt"
	"strings"
)

type Droplet struct {
	Path string
}

func (d *Droplet) UnmarshalFlag(val string) error {
	if !strings.HasPrefix(val, "/") {
		d.Path = fmt.Sprintf("/%s", val)
	} else {
		d.Path = val
	}
	return nil
}
