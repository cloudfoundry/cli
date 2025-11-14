package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type SpaceRole struct {
	Role string
}

func (SpaceRole) Complete(prefix string) []flags.Completion {
	return completions([]string{"SpaceManager", "SpaceDeveloper", "SpaceAuditor", "SpaceSupporter"}, prefix, false)
}

func (s *SpaceRole) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "spaceauditor":
		s.Role = "SpaceAuditor"
	case "spacedeveloper":
		s.Role = "SpaceDeveloper"
	case "spacemanager":
		s.Role = "SpaceManager"
	case "spacesupporter":
		s.Role = "SpaceSupporter"
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `ROLE must be "SpaceManager", "SpaceDeveloper", "SpaceAuditor" or "SpaceSupporter"`,
		}
	}

	return nil
}
