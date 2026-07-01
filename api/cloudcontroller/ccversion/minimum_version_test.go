package ccversion_test

import (
	"testing"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
)

// TestMinVersionRoutePoliciesIsNotPlaceholder will keep failing until
// MinVersionRoutePolicies is updated from its placeholder value ("3.999.0")
// to the real CAPI version that introduces /v3/route_policies and the
// enforce_route_policies / route_policies_scope domain fields.
//
// To fix: coordinate with the CAPI team, confirm the released version, and
// replace ccversion.MinVersionRoutePolicies with the real value.
func TestMinVersionRoutePoliciesIsNotPlaceholder(t *testing.T) {
	const placeholder = "3.999.0"
	if ccversion.MinVersionRoutePolicies == placeholder {
		t.Fatalf(
			"ccversion.MinVersionRoutePolicies is still the placeholder %q.\n"+
				"Update it with the real CAPI version that supports route policies once known.",
			placeholder,
		)
	}
}
