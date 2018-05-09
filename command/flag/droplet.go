package flag

import (
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type Droplet struct {
	Path string
}

func (d *Droplet) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --droplet, but got option %s", val),
		}
	}
	if !strings.HasPrefix(val, "/") {
		d.Path = fmt.Sprintf("/%s", val)
	} else {
		d.Path = val
	}
	return nil
}
