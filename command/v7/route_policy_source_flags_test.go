package v7_test

// Tests for RoutePolicySourceFlags validation and resolution, exercised
// through AddRoutePolicyCommand.Execute() using the standard v7fakes.FakeActor
// pattern instead of an inline stub actor.

import (
	"errors"
	"fmt"

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

var _ = Describe("RoutePolicySourceFlags", func() {
	var (
		cmd             AddRoutePolicyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		fakeConfig.APIVersionReturns("3.999.0")
		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: "targeted-space-guid", Name: "targeted-space"})
		fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: "targeted-org-guid", Name: "targeted-org"})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: "some-user"}, nil)
		fakeActor.AddRoutePolicyReturns(nil, nil)

		cmd = AddRoutePolicyCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.RoutePolicyArgs{Domain: "apps.example.com"},
			Hostname:     "myapp",
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("validateSourceFlags", func() {
		It("returns RequiredArgumentError when no source flag is given", func() {
			Expect(executeErr).To(MatchError(translatableerror.RequiredArgumentError{
				ArgumentName: "one of: --source-app, --source-space, --source-org, --source-any, or --source",
			}))
		})

		When("--source-org is given with --source-app but without --source-space", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app", SourceOrg: "my-org"}
			})
			It("returns RequiredFlagsError", func() {
				Expect(executeErr).To(MatchError(translatableerror.RequiredFlagsError{
					Arg1: "--source-org",
					Arg2: "--source-space",
				}))
			})
		})

		When("two primary flags are given (--source-app + --source-any)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app", SourceAny: true}
			})
			It("returns ArgumentCombinationError", func() {
				Expect(executeErr).To(BeAssignableToTypeOf(translatableerror.ArgumentCombinationError{}))
				Expect(executeErr.(translatableerror.ArgumentCombinationError).Args).To(ConsistOf("--source-app", "--source-any"))
			})
		})

		When("two primary flags are given (--source + --source-any)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{Source: "cf:any", SourceAny: true}
			})
			It("returns ArgumentCombinationError", func() {
				Expect(executeErr).To(BeAssignableToTypeOf(translatableerror.ArgumentCombinationError{}))
				Expect(executeErr.(translatableerror.ArgumentCombinationError).Args).To(ConsistOf("--source", "--source-any"))
			})
		})
	})

	Describe("resolveSource", func() {
		Context("--source (raw passthrough)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{Source: "cf:app:some-guid"}
			})

			It("passes the raw value to AddRoutePolicy without calling any actor lookup", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.GetOrganizationByNameCallCount()).To(Equal(0))
				Expect(fakeActor.GetSpaceByNameAndOrganizationCallCount()).To(Equal(0))
				Expect(fakeActor.GetApplicationByNameAndSpaceCallCount()).To(Equal(0))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:app:some-guid"))
			})
		})

		Context("--source-any", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceAny: true}
			})

			It("passes cf:any to AddRoutePolicy", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:any"))
				Expect(testUI.Out).To(Say("scope: any"))
			})
		})

		Context("--source-app (app in targeted space)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app"}
				fakeActor.GetApplicationByNameAndSpaceReturns(
					resources.Application{GUID: "app-guid"},
					v7action.Warnings{"app-warning"},
					nil,
				)
			})

			It("resolves the app GUID using the targeted space and passes cf:app:<guid>", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("my-app"))
				Expect(spaceGUID).To(Equal("targeted-space-guid"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:app:app-guid"))
				Expect(testUI.Err).To(Say("app-warning"))
			})
		})

		Context("--source-app when the app is not found in the targeted space", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app"}
				fakeActor.GetApplicationByNameAndSpaceReturns(
					resources.Application{},
					nil,
					actionerror.ApplicationNotFoundError{Name: "my-app"},
				)
			})

			It("returns a friendly error with a TIP about --source-space / --source-org", func() {
				Expect(executeErr).To(MatchError(fmt.Sprintf(
					"App 'my-app' not found in space 'targeted-space' / org 'targeted-org'.\nTIP: If the app is in a different space or org, use --source-space and/or --source-org flags.",
				)))
			})
		})

		Context("--source-app with a non-NotFound actor error", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app"}
				fakeActor.GetApplicationByNameAndSpaceReturns(
					resources.Application{},
					v7action.Warnings{"w1"},
					errors.New("upstream error"),
				)
			})

			It("passes the error through and displays warnings", func() {
				Expect(executeErr).To(MatchError("upstream error"))
				Expect(testUI.Err).To(Say("w1"))
			})
		})

		Context("--source-app + --source-space (cross-space app lookup)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceApp: "my-app", SourceSpace: "other-space"}
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					resources.Space{GUID: "other-space-guid", Name: "other-space"},
					v7action.Warnings{"space-warning"},
					nil,
				)
				fakeActor.GetApplicationByNameAndSpaceReturns(
					resources.Application{GUID: "app-guid"},
					v7action.Warnings{"app-warning"},
					nil,
				)
			})

			It("resolves space (in targeted org) then app, includes space in scope display", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("other-space"))
				Expect(orgGUID).To(Equal("targeted-org-guid"))
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("my-app"))
				Expect(spaceGUID).To(Equal("other-space-guid"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:app:app-guid"))
				Expect(testUI.Out).To(Say("space: other-space"))
				Expect(testUI.Err).To(Say("space-warning"))
				Expect(testUI.Err).To(Say("app-warning"))
			})
		})

		Context("--source-app + --source-space + --source-org (cross-org/space app lookup)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{
					SourceApp:   "my-app",
					SourceSpace: "other-space",
					SourceOrg:   "other-org",
				}
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{GUID: "other-org-guid"},
					v7action.Warnings{"org-warning"},
					nil,
				)
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					resources.Space{GUID: "other-space-guid", Name: "other-space"},
					v7action.Warnings{"space-warning"},
					nil,
				)
				fakeActor.GetApplicationByNameAndSpaceReturns(
					resources.Application{GUID: "app-guid"},
					v7action.Warnings{"app-warning"},
					nil,
				)
			})

			It("resolves org, space, then app; includes both in scope display", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.GetOrganizationByNameArgsForCall(0)).To(Equal("other-org"))
				spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("other-space"))
				Expect(orgGUID).To(Equal("other-org-guid"))
				appName, spaceGUID := fakeActor.GetApplicationByNameAndSpaceArgsForCall(0)
				Expect(appName).To(Equal("my-app"))
				Expect(spaceGUID).To(Equal("other-space-guid"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:app:app-guid"))
				Expect(testUI.Out).To(Say("space: other-space, org: other-org"))
				Expect(testUI.Err).To(Say("org-warning"))
				Expect(testUI.Err).To(Say("space-warning"))
				Expect(testUI.Err).To(Say("app-warning"))
			})
		})

		Context("--source-space (standalone space policy in targeted org)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceSpace: "my-space"}
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					resources.Space{GUID: "space-guid"},
					v7action.Warnings{"space-warning"},
					nil,
				)
			})

			It("resolves space in the targeted org and passes cf:space:<guid>", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("my-space"))
				Expect(orgGUID).To(Equal("targeted-org-guid"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:space:space-guid"))
				Expect(testUI.Err).To(Say("space-warning"))
			})
		})

		Context("--source-space + --source-org (space in a different org)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceSpace: "my-space", SourceOrg: "other-org"}
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{GUID: "other-org-guid"},
					v7action.Warnings{"org-warning"},
					nil,
				)
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					resources.Space{GUID: "space-guid"},
					v7action.Warnings{"space-warning"},
					nil,
				)
			})

			It("resolves org then space; includes org in scope display", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.GetOrganizationByNameArgsForCall(0)).To(Equal("other-org"))
				spaceName, orgGUID := fakeActor.GetSpaceByNameAndOrganizationArgsForCall(0)
				Expect(spaceName).To(Equal("my-space"))
				Expect(orgGUID).To(Equal("other-org-guid"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:space:space-guid"))
				Expect(testUI.Out).To(Say("org: other-org"))
				Expect(testUI.Err).To(Say("org-warning"))
				Expect(testUI.Err).To(Say("space-warning"))
			})
		})

		Context("--source-org (standalone org policy)", func() {
			BeforeEach(func() {
				cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceOrg: "my-org"}
				fakeActor.GetOrganizationByNameReturns(
					resources.Organization{GUID: "org-guid"},
					v7action.Warnings{"org-warning"},
					nil,
				)
			})

			It("resolves the org and passes cf:org:<guid>", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(fakeActor.GetOrganizationByNameArgsForCall(0)).To(Equal("my-org"))
				_, sourceArg, _, _ := fakeActor.AddRoutePolicyArgsForCall(0)
				Expect(sourceArg).To(Equal("cf:org:org-guid"))
				Expect(testUI.Err).To(Say("org-warning"))
			})
		})

		Context("error propagation", func() {
			When("GetOrganizationByName fails (--source-org)", func() {
				BeforeEach(func() {
					cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceOrg: "bad-org"}
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"warn"},
						errors.New("org-lookup-error"),
					)
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(MatchError("org-lookup-error"))
					Expect(testUI.Err).To(Say("warn"))
				})
			})

			When("GetOrganizationByName fails (--source-space + --source-org)", func() {
				BeforeEach(func() {
					cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceSpace: "my-space", SourceOrg: "bad-org"}
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"warn"},
						errors.New("org-lookup-error"),
					)
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(MatchError("org-lookup-error"))
					Expect(testUI.Err).To(Say("warn"))
				})
			})

			When("GetSpaceByNameAndOrganization fails (--source-space)", func() {
				BeforeEach(func() {
					cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{SourceSpace: "bad-space"}
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{},
						v7action.Warnings{"warn"},
						errors.New("space-lookup-error"),
					)
				})

				It("returns the error and displays warnings", func() {
					Expect(executeErr).To(MatchError("space-lookup-error"))
					Expect(testUI.Err).To(Say("warn"))
				})
			})

			When("GetOrganizationByName fails (--source-app + --source-space + --source-org)", func() {
				BeforeEach(func() {
					cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{
						SourceApp:   "my-app",
						SourceSpace: "my-space",
						SourceOrg:   "bad-org",
					}
					fakeActor.GetOrganizationByNameReturns(
						resources.Organization{},
						v7action.Warnings{"warn"},
						errors.New("org-lookup-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("org-lookup-error"))
				})
			})

			When("GetSpaceByNameAndOrganization fails (--source-app + --source-space)", func() {
				BeforeEach(func() {
					cmd.RoutePolicySourceFlags = RoutePolicySourceFlags{
						SourceApp:   "my-app",
						SourceSpace: "bad-space",
					}
					fakeActor.GetSpaceByNameAndOrganizationReturns(
						resources.Space{},
						v7action.Warnings{"warn"},
						errors.New("space-lookup-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("space-lookup-error"))
				})
			})
		})
	})
})
