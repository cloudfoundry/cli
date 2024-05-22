package flag

import (
	"fmt"

	"github.com/docker/distribution/reference"
	flags "github.com/jessevdk/go-flags"
)

type DockerImage struct {
	Path string
}

func (d *DockerImage) UnmarshalFlag(val string) error {
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
