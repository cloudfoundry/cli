package flag

import (
	"fmt"
	"strings"

	"github.com/docker/distribution/reference"
	flags "github.com/jessevdk/go-flags"
)

type DockerImage struct {
	Path string
}

func (d *DockerImage) UnmarshalFlag(val string) error {
	if strings.HasPrefix(val, "-") {
		return &flags.Error{
			Type:    flags.ErrExpectedArgument,
			Message: fmt.Sprintf("expected argument for flag --docker-image, but got option %s", val),
		}
	}
	_, err := reference.Parse(val)
	if err != nil {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: fmt.Sprintf("invalid docker reference: %s", err.Error()),
		}
	}
	d.Path = val
	return nil
}
