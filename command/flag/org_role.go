package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type OrgRole struct {
	Role string
}

func (s OrgRole) Complete(prefix string) []flags.Completion {
	return completions([]string{"OrgManager", "BillingManager", "OrgAuditor"}, prefix)
}

func (s *OrgRole) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "orgauditor":
		s.Role = "OrgAuditor"
	case "billingmanager":
		s.Role = "BillingManager"
	case "orgmanager":
		s.Role = "OrgManager"
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`,
		}
	}

	return nil
}
