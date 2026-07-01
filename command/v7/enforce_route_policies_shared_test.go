package v7_test

import (
	ccversionPkg "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccversion"
	translatablerrorPkg "code.cloudfoundry.org/cli/v8/command/translatableerror"
	uiPkg "code.cloudfoundry.org/cli/v8/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

// EnforceRoutePoliciesBehavior carries the command-specific hooks needed to
// run the shared --enforce-route-policies / --scope test suite.  Callers
// populate it with closures that capture their local test variables so the
// shared specs remain free of command-specific knowledge.
type EnforceRoutePoliciesBehavior struct {
	// Setup — called inside BeforeEach blocks within the shared suite.
	SetEnforce       func(bool)
	SetScope         func(string)
	SetAPIVersion    func(string)
	SetActorSucceeds func() // configure the actor's create-domain call to return no error

	// Observation — called inside It blocks after JustBeforeEach has run Execute.
	ExecuteErr func() error
	UI         func() *uiPkg.UI
	DomainName func() string
	EnforceArg func() bool   // the enforceAccessRules arg forwarded to the actor
	ScopeArg   func() string // the scope arg forwarded to the actor

	// The domain-type-specific adjective in the identity-aware TIP message,
	// e.g. "shared" or "private".
	TIPAdjective string
}

// ItEnforcesRoutePolicies injects all shared When/It blocks that cover the
// --enforce-route-policies and --scope flag behaviour.  It must be called
// inside the "When the environment is setup correctly" context so that
// GetCurrentUser is already stubbed and JustBeforeEach runs Execute.
func ItEnforcesRoutePolicies(b *EnforceRoutePoliciesBehavior) {
	When("--scope is specified without --enforce-route-policies", func() {
		BeforeEach(func() {
			b.SetEnforce(false)
			b.SetScope("org")
		})

		It("returns an error", func() {
			Expect(b.ExecuteErr()).To(MatchError("--scope can only be used with --enforce-route-policies"))
		})
	})

	When("--scope has an invalid value", func() {
		BeforeEach(func() {
			b.SetEnforce(true)
			b.SetScope("invalid")
		})

		It("returns an error", func() {
			Expect(b.ExecuteErr()).To(MatchError("--scope must be one of: any, org, space"))
		})
	})

	When("--enforce-route-policies is specified", func() {
		BeforeEach(func() {
			b.SetEnforce(true)
			b.SetAPIVersion(ccversionPkg.MinVersionRoutePolicies)
			b.SetActorSucceeds()
		})

		When("the API version is too old", func() {
			BeforeEach(func() {
				b.SetAPIVersion("0.0.0")
			})

			It("returns a version error", func() {
				Expect(b.ExecuteErr()).To(MatchError(translatablerrorPkg.MinimumCFAPIVersionNotMetError{
					Command:        "--enforce-route-policies",
					CurrentVersion: "0.0.0",
					MinimumVersion: ccversionPkg.MinVersionRoutePolicies,
				}))
			})
		})

		When("the API version is sufficient", func() {
			It("passes enforce=true and empty scope to the actor", func() {
				Expect(b.ExecuteErr()).NotTo(HaveOccurred())
				Expect(b.EnforceArg()).To(BeTrue())
				Expect(b.ScopeArg()).To(BeEmpty())
			})

			It("prints the identity-aware TIP", func() {
				Expect(b.ExecuteErr()).NotTo(HaveOccurred())
				Expect(b.UI().Out).To(Say(
					"TIP: Domain '%s' is a %s identity-aware domain",
					b.DomainName(), b.TIPAdjective,
				))
			})

			When("--scope is also specified", func() {
				BeforeEach(func() {
					b.SetScope("org")
				})

				It("passes the scope to the actor", func() {
					Expect(b.ExecuteErr()).NotTo(HaveOccurred())
					Expect(b.EnforceArg()).To(BeTrue())
					Expect(b.ScopeArg()).To(Equal("org"))
				})
			})
		})
	})
}
