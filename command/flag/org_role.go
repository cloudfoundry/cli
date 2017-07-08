package flag

import (
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type OrgRole struct {
	Role string
}

func (OrgRole) Complete(prefix string) []flags.Completion {
	return completions([]string{"OrgManager", "BillingManager", "OrgAuditor"}, prefix, false)
}

func (o *OrgRole) UnmarshalFlag(val string) error {
	switch strings.ToLower(val) {
	case "orgauditor":
		o.Role = "OrgAuditor"
	case "billingmanager":
		o.Role = "BillingManager"
	case "orgmanager":
		o.Role = "OrgManager"
	default:
		return &flags.Error{
			Type:    flags.ErrRequired,
			Message: `ROLE must be "OrgManager", "BillingManager" and "OrgAuditor"`,
		}
	}

	return nil
}
