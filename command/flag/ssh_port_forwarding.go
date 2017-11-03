package flag

import (
	"fmt"
	"regexp"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

const DefaultLocalAddress = "localhost"

type SSHPortForwarding struct {
	LocalAddress  string
	RemoteAddress string
}

func (s *SSHPortForwarding) UnmarshalFlag(val string) error {
	splitHosts := strings.Split(val, ":")
	for _, piece := range splitHosts {
		if len(piece) == 0 {
			return &flags.Error{
				Type:    flags.ErrRequired,
				Message: fmt.Sprintf("Bad local forwarding specification '%s'", val),
			}
		}
	}

	re := regexp.MustCompile("^\\d+$")
	switch {
	case len(splitHosts) == 3 && re.MatchString(splitHosts[0]) && re.MatchString(splitHosts[2]):
		s.LocalAddress = fmt.Sprintf("%s:%s", DefaultLocalAddress, splitHosts[0])
		s.RemoteAddress = fmt.Sprintf("%s:%s", splitHosts[1], splitHosts[2])
	case len(splitHosts) == 4 && re.MatchString(splitHosts[1]) && re.MatchString(splitHosts[3]):
		s.LocalAddress = fmt.Sprintf("%s:%s", splitHosts[0], splitHosts[1])
		s.RemoteAddress = fmt.Sprintf("%s:%s", splitHosts[2], splitHosts[3])
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: fmt.Sprintf("Bad local forwarding specification '%s'", val),
		}
	}

	return nil
}
