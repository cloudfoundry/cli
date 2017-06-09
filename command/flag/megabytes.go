package flag

import (
	"strings"

	"github.com/cloudfoundry/bytefmt"
	flags "github.com/jessevdk/go-flags"
)

type Megabytes struct {
	Size uint64
}

func (m *Megabytes) UnmarshalFlag(val string) error {
	size, err := bytefmt.ToMegabytes(val)

	if err != nil ||
		!strings.ContainsAny(strings.ToLower(val), "mg") ||
		strings.Contains(strings.ToLower(val), ".") {
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB`,
		}
	}

	m.Size = size
	return nil
}
