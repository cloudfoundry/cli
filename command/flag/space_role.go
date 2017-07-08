package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type SpaceRole struct {
	Role string
}

func (SpaceRole) Complete(prefix string) []flags.Completion {
	return completions([]string{"SpaceManager", "SpaceDeveloper", "SpaceAuditor"}, prefix, false)
}

func (s *SpaceRole) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "spaceauditor":
		s.Role = "SpaceAuditor"
	case "spacedeveloper":
		s.Role = "SpaceDeveloper"
	case "spacemanager":
		s.Role = "SpaceManager"
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `ROLE must be "SpaceManager", "SpaceDeveloper" and "SpaceAuditor"`,
		}
	}

	return nil
}
