package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-shared-domain Command", func() {
	var (
		cmd             CreateSharedDomainCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor

		executeErr error

		binaryName string
		domainName string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		domainName = "example.com"

		cmd = CreateSharedDomainCommand{
			RequiredArgs: flag.Domain{
				Domain: domainName,
			},
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the environment is not set up correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{Name: "the-user"}, nil)
		})

		It("should print text indicating it is creating a domain", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say(`Creating shared domain %s as the-user\.\.\.`, domainName))
		})

		When("creating the domain errors", func() {
			BeforeEach(func() {
				fakeActor.CreateSharedDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("err-create-domain"))
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("err-create-domain"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("the provided router group does not exist", func() {
			BeforeEach(func() {
				fakeActor.CreateSharedDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, errors.New("bad-router-group"))
				cmd.RouterGroup = "bogus"
			})

			It("returns an error and displays warnings", func() {
				Expect(executeErr).To(MatchError("bad-router-group"))
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
			})
		})

		When("creating the domain is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateSharedDomainReturns(v7action.Warnings{"warnings-1", "warnings-2"}, nil)
				cmd.Internal = true
				cmd.RouterGroup = "router-group"
			})

			It("prints all warnings, ok and then a tip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("warnings-1"))
				Expect(testUI.Err).To(Say("warnings-2"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).To(Say("TIP: Domain '%s' is shared with all orgs. Run 'cf domains' to view available domains.", domainName))
			})

			It("creates the domain", func() {
				Expect(fakeActor.CreateSharedDomainCallCount()).To(Equal(1))
				expectedDomainName, expectedInternal, expectedRouterGroup, enforceRules, scope := fakeActor.CreateSharedDomainArgsForCall(0)
				Expect(expectedDomainName).To(Equal(domainName))
				Expect(expectedInternal).To(BeTrue())
				Expect(expectedRouterGroup).To(Equal("router-group"))
				Expect(enforceRules).To(BeFalse())
				Expect(scope).To(BeEmpty())
			})
		})

		When("--scope is specified without --enforce-route-policies", func() {
			BeforeEach(func() {
				cmd.Scope = "org"
				cmd.EnforceRoutePolicies = false
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("--scope can only be used with --enforce-route-policies"))
			})
		})

		When("--scope has an invalid value", func() {
			BeforeEach(func() {
				cmd.EnforceRoutePolicies = true
				cmd.Scope = "invalid"
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("--scope must be one of: any, org, space"))
			})
		})

		When("--enforce-route-policies is specified", func() {
			BeforeEach(func() {
				cmd.EnforceRoutePolicies = true
				fakeConfig.APIVersionReturns(ccversion.MinVersionRoutePolicies)
				fakeActor.CreateSharedDomainReturns(v7action.Warnings{}, nil)
			})

			When("the API version is too old", func() {
				BeforeEach(func() {
					fakeConfig.APIVersionReturns("0.0.0")
				})

				It("returns a version error", func() {
					Expect(executeErr).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
						Command:        "--enforce-route-policies",
						CurrentVersion: "0.0.0",
						MinimumVersion: ccversion.MinVersionRoutePolicies,
					}))
				})
			})

			When("the API version is sufficient", func() {
				It("passes enforce=true and empty scope to the actor", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(fakeActor.CreateSharedDomainCallCount()).To(Equal(1))
					_, _, _, enforceRules, scope := fakeActor.CreateSharedDomainArgsForCall(0)
					Expect(enforceRules).To(BeTrue())
					Expect(scope).To(BeEmpty())
				})

				It("prints the identity-aware TIP", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).To(Say("TIP: Domain '%s' is a shared identity-aware domain", domainName))
				})

				When("--scope is also specified", func() {
					BeforeEach(func() {
						cmd.Scope = "org"
					})

					It("passes the scope to the actor", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(fakeActor.CreateSharedDomainCallCount()).To(Equal(1))
						_, _, _, enforceRules, scope := fakeActor.CreateSharedDomainArgsForCall(0)
						Expect(enforceRules).To(BeTrue())
						Expect(scope).To(Equal("org"))
					})
				})
			})
		})
	})
})
