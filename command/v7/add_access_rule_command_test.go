package v7_test

import (
	"code.cloudfoundry.org/cli/v9/actor/actionerror"
	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command/commandfakes"
	"code.cloudfoundry.org/cli/v9/command/flag"
	"code.cloudfoundry.org/cli/v9/command/translatableerror"
	. "code.cloudfoundry.org/cli/v9/command/v7"
	"code.cloudfoundry.org/cli/v9/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v9/resources"
	"code.cloudfoundry.org/cli/v9/util/configv3"
	"code.cloudfoundry.org/cli/v9/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("add-access-rule Command", func() {
	var (
		cmd             AddAccessRuleCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = AddAccessRuleCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		// Setup default config returns
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			GUID: "org-guid",
			Name: "org-name",
		})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "space-guid",
			Name: "space-name",
		})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: "test-user"}, nil)

		args = []string{}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	Describe("validation", func() {
		Context("when no source flags are provided", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.AddAccessRuleArgs{
					RuleName: "test-rule",
					Domain:   "apps.internal",
				}
				cmd.Hostname = "backend"
			})

			It("returns a RequiredArgumentError", func() {
				Expect(executeErr).To(MatchError(translatableerror.RequiredArgumentError{
					ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --selector",
				}))
			})
		})

		Context("when multiple mutually exclusive source flags are provided", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.AddAccessRuleArgs{
					RuleName: "test-rule",
					Domain:   "apps.internal",
				}
				cmd.Hostname = "backend"
				cmd.SourceApp = "app-name"
				cmd.SourceAny = true
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--source-app", "--source-any"},
				}))
			})
		})

		Context("when --source-space and --source-any are both provided", func() {
			BeforeEach(func() {
				cmd.RequiredArgs = flag.AddAccessRuleArgs{
					RuleName: "test-rule",
					Domain:   "apps.internal",
				}
				cmd.Hostname = "backend"
				cmd.SourceSpace = "some-space"
				cmd.SourceAny = true
			})

			It("returns an ArgumentCombinationError", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--source-space", "--source-any"},
				}))
			})
		})
	})

	When("the user is logged in, an org is targeted, and a space is targeted", func() {
		BeforeEach(func() {
			cmd.RequiredArgs = flag.AddAccessRuleArgs{
				RuleName: "test-rule",
				Domain:   "apps.internal",
			}
			cmd.Hostname = "backend"
		})

		Describe("source resolution", func() {
			Context("when --source-app is provided (current space)", func() {
				BeforeEach(func() {
					cmd.SourceApp = "frontend-app"
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{GUID: "app-guid"},
						v7action.Warnings{"app-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("resolves the app and creates the access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify app lookup
					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("frontend-app"))
					Expect(spaceGUID).To(Equal("space-guid"))

					// Verify access rule creation with resolved selector
					Expect(fakeActor.AddAccessRuleCallCount()).To(Equal(1))
					ruleName, domain, selector, hostname, path := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(ruleName).To(Equal("test-rule"))
					Expect(domain).To(Equal("apps.internal"))
					Expect(selector).To(Equal("cf:app:app-guid"))
					Expect(hostname).To(Equal("backend"))
					Expect(path).To(BeEmpty())

					// Verify output
					Expect(testUI.Out).To(Say("Adding access rule test-rule"))
					Expect(testUI.Out).To(Say("scope: app, source: frontend-app"))
					Expect(testUI.Out).To(Say("selector: cf:app:app-guid"))
					Expect(testUI.Out).To(Say("OK"))
				})

				It("displays warnings", func() {
					Expect(testUI.Err).To(Say("app-warning"))
					Expect(testUI.Err).To(Say("add-warning"))
				})
			})

			Context("when --source-app is provided with --source-space (cross-space)", func() {
				BeforeEach(func() {
					cmd.SourceApp = "frontend-app"
					cmd.SourceSpace = "other-space"

					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{GUID: "other-space-guid"},
						v7action.Warnings{"space-warning"},
						nil,
					)
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{GUID: "app-guid"},
						v7action.Warnings{"app-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("resolves space then app and creates the access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify space lookup
					Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(1))
					spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
					Expect(spaceName).To(Equal("other-space"))
					Expect(orgGUID).To(Equal("org-guid"))

					// Verify app lookup in resolved space
					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
					appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
					Expect(appName).To(Equal("frontend-app"))
					Expect(spaceGUID).To(Equal("other-space-guid"))

					// Verify selector
					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:app:app-guid"))

					// Verify output shows cross-space info
					Expect(testUI.Out).To(Say("scope: app, source: frontend-app \\(space: other-space\\)"))
				})
			})

			Context("when --source-app is provided with --source-space and --source-org (cross-org)", func() {
				BeforeEach(func() {
					cmd.SourceApp = "frontend-app"
					cmd.SourceSpace = "other-space"
					cmd.SourceOrg = "other-org"

					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{GUID: "other-org-guid"},
						v7action.Warnings{"org-warning"},
						nil,
					)
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{GUID: "other-space-guid"},
						v7action.Warnings{"space-warning"},
						nil,
					)
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{GUID: "app-guid"},
						v7action.Warnings{"app-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("resolves org, space, then app and creates the access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify org lookup
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
					Expect(orgName).To(Equal("other-org"))

					// Verify space lookup with resolved org
					spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
					Expect(spaceName).To(Equal("other-space"))
					Expect(orgGUID).To(Equal("other-org-guid"))

					// Verify output shows cross-org info
					Expect(testUI.Out).To(Say("scope: app, source: frontend-app \\(space: other-space, org: other-org\\)"))
				})
			})

			Context("when --source-app is not found in current space", func() {
				BeforeEach(func() {
					cmd.SourceApp = "missing-app"
					fakeActor.GetApplicationByNameAndSpaceReturns(
						resources.Application{},
						v7action.Warnings{"app-warning"},
						actionerror.ApplicationNotFoundError{Name: "missing-app"},
					)
				})

				It("returns a helpful error message", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(ContainSubstring("App 'missing-app' not found in space 'space-name' / org 'org-name'"))
					Expect(executeErr.Error()).To(ContainSubstring("TIP: If the app is in a different space or org, use --source-space and/or --source-org flags"))
				})
			})

			Context("when --source-space is provided (without --source-app)", func() {
				BeforeEach(func() {
					cmd.SourceSpace = "monitoring-space"
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{GUID: "space-guid-123"},
						v7action.Warnings{"space-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("creates a space-level access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify selector is space-level
					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:space:space-guid-123"))

					Expect(testUI.Out).To(Say("scope: space, source: monitoring-space"))
				})
			})

			Context("when --source-space is provided with --source-org (cross-org space rule)", func() {
				BeforeEach(func() {
					cmd.SourceSpace = "prod-space"
					cmd.SourceOrg = "prod-org"

					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{GUID: "prod-org-guid"},
						v7action.Warnings{"org-warning"},
						nil,
					)
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{GUID: "prod-space-guid"},
						v7action.Warnings{"space-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("creates a space-level access rule for the specified org", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify org lookup
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(1))
					orgName := fakeActor.GetOrganizationByNameArgsForCall(0)
					Expect(orgName).To(Equal("prod-org"))

					// Verify space lookup with resolved org
					spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
					Expect(spaceName).To(Equal("prod-space"))
					Expect(orgGUID).To(Equal("prod-org-guid"))

					// Verify selector is space-level
					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:space:prod-space-guid"))

					Expect(testUI.Out).To(Say("scope: space, source: prod-space \\(org: prod-org\\)"))
				})
			})

			Context("when --source-org is provided (without --source-space or --source-app)", func() {
				BeforeEach(func() {
					cmd.SourceOrg = "platform-org"
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{GUID: "org-guid-456"},
						v7action.Warnings{"org-warning"},
						nil,
					)
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("creates an org-level access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					// Verify selector is org-level
					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:org:org-guid-456"))

					Expect(testUI.Out).To(Say("scope: org, source: platform-org"))
				})
			})

			Context("when --source-any is provided", func() {
				BeforeEach(func() {
					cmd.SourceAny = true
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("creates an 'any' access rule", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:any"))

					Expect(testUI.Out).To(Say("scope: any, source: any authenticated app"))
				})
			})

			Context("when --selector is provided (raw selector)", func() {
				BeforeEach(func() {
					cmd.Selector = "cf:app:raw-guid-123"
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("uses the raw selector without resolution", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					_, _, selector, _, _ := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(selector).To(Equal("cf:app:raw-guid-123"))

					Expect(testUI.Out).To(Say("selector: cf:app:raw-guid-123"))

					// Should not call any resolution methods
					Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))
					Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(0))
					Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				})
			})

			Context("when --path is provided", func() {
				BeforeEach(func() {
					cmd.SourceAny = true
					cmd.Path = "/metrics"
					fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, nil)
				})

				It("passes the path to the actor", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					_, _, _, _, path := fakeActor.AddAccessRuleArgsForCall(0)
					Expect(path).To(Equal("/metrics"))
				})
			})
		})

		Describe("error handling", func() {
		Context("when AddAccessRule fails", func() {
			BeforeEach(func() {
				cmd.SourceAny = true
				fakeActor.AddAccessRuleReturns(v7action.Warnings{"add-warning"}, actionerror.RouteNotFoundError{})
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{}))
				Expect(testUI.Err).To(Say("add-warning"))
			})
		})

			Context("when space lookup fails", func() {
				BeforeEach(func() {
					cmd.SourceSpace = "nonexistent-space"
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{},
						v7action.Warnings{"space-warning"},
						actionerror.SpaceNotFoundError{Name: "nonexistent-space"},
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(actionerror.SpaceNotFoundError{Name: "nonexistent-space"}))
					Expect(testUI.Err).To(Say("space-warning"))
				})
			})

			Context("when org lookup fails", func() {
				BeforeEach(func() {
					cmd.SourceOrg = "nonexistent-org"
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"org-warning"},
						actionerror.OrganizationNotFoundError{Name: "nonexistent-org"},
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "nonexistent-org"}))
					Expect(testUI.Err).To(Say("org-warning"))
				})
			})
		})
	})
})
