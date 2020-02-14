package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("routes Command", func() {
	var (
		cmd             RoutesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
		args            []string
		binaryName      string
	)

	const tableHeaders = `space\s+host\s+domain\s+path\s+apps`

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)
		args = nil

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = RoutesCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			Orglevel:    false,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(args)
	})

	When("the environment is not setup correctly", func() {
		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})

		When("when there is no org targeted", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))
				checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkTargetedOrg).To(BeTrue())
				Expect(checkTargetedSpace).To(BeTrue())
			})
		})
	})

	Context("When the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "banana"}, nil)
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				GUID: "some-org-guid",
				Name: "some-org",
			})

			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		When("getting routes succeeds", func() {
			var (
				routes []v7action.Route
			)

			BeforeEach(func() {
				routes = []v7action.Route{
					{DomainName: "domain3", GUID: "route-guid-3", SpaceName: "space-3", Host: "host-1"},
					{DomainName: "domain1", GUID: "route-guid-1", SpaceName: "space-1"},
					{DomainName: "domain2", GUID: "route-guid-2", SpaceName: "space-2", Host: "host-3", Path: "/path/2"},
				}

				fakeActor.GetRoutesBySpaceReturns(
					routes,
					v7action.Warnings{"actor-warning-1"},
					nil,
				)
			})

			It("delegates to the actor for summaries", func() {
				Expect(fakeActor.GetRouteSummariesCallCount()).To(Equal(1))

				Expect(fakeActor.GetRouteSummariesArgsForCall(0)).To(Equal(routes))
			})

			When("getting route summaries succeeds", func() {
				var (
					routeSummaries []v7action.RouteSummary
				)

				BeforeEach(func() {
					routeSummaries = []v7action.RouteSummary{
						{Route: v7action.Route{DomainName: "domain1", GUID: "route-guid-1", SpaceName: "space-1"}},
						{Route: v7action.Route{DomainName: "domain2", GUID: "route-guid-2", SpaceName: "space-2", Host: "host-3", Path: "/path/2"}},
						{Route: v7action.Route{DomainName: "domain3", GUID: "route-guid-3", SpaceName: "space-3", Host: "host-1"}, AppNames: []string{"app1", "app2"}},
					}

					fakeActor.GetRouteSummariesReturns(
						routeSummaries,
						v7action.Warnings{"actor-warning-2"},
						nil,
					)
				})

				It("prints routes in a table", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Err).To(Say("actor-warning-1"))
					Expect(testUI.Err).To(Say("actor-warning-2"))

					Expect(testUI.Out).To(Say(tableHeaders))
					Expect(testUI.Out).To(Say(`space-1\s+domain1\s+`))
					Expect(testUI.Out).To(Say(`space-2\s+host-3\s+domain2\s+\/path\/2`))
					Expect(testUI.Out).To(Say(`space-3\s+host-1\s+domain3\s+app1, app2`))
				})
			})

			When("getting route summaries fails", func() {
				BeforeEach(func() {
					fakeActor.GetRouteSummariesReturns(
						nil,
						v7action.Warnings{"actor-warning-2", "actor-warning-3"},
						errors.New("summaries-error"),
					)
				})

				It("prints warnings and returns error", func() {
					Expect(executeErr).To(MatchError("summaries-error"))

					Expect(testUI.Err).To(Say("actor-warning-1"))
					Expect(testUI.Err).To(Say("actor-warning-2"))
					Expect(testUI.Err).To(Say("actor-warning-3"))
				})
			})
		})

		When("getting space routes fails", func() {
			var expectedErr error

			BeforeEach(func() {
				warnings := v7action.Warnings{"warning-1", "warning-2"}
				expectedErr = errors.New("some-error")
				fakeActor.GetRoutesBySpaceReturns(nil, warnings, expectedErr)
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(Equal(expectedErr))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
				Expect(testUI.Out).ToNot(Say(tableHeaders))
			})
		})

		When("--org-level is passed and getting org routes fails", func() {
			var expectedErr error

			BeforeEach(func() {
				cmd.Orglevel = true
				warnings := v7action.Warnings{"warning-1", "warning-2"}
				expectedErr = errors.New("some-error")
				fakeActor.GetRoutesByOrgReturns(nil, warnings, expectedErr)
			})

			It("prints warnings and returns error", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
				Expect(testUI.Out).ToNot(Say(tableHeaders))
			})
		})

		When("--labels is passed in", func() {
			BeforeEach(func() {
				cmd.Labels = "some_label=fun"
			})
			It("passes the labels to the actor", func() {
				_, labels := fakeActor.GetRoutesBySpaceArgsForCall(0)
				Expect(labels).To(Equal("some_label=fun"))
			})
			When("--org-level is passed in", func() {
				BeforeEach(func() {
					cmd.Orglevel = true
				})
				It("passes the labels to the actor", func() {
					_, labels := fakeActor.GetRoutesByOrgArgsForCall(0)
					Expect(labels).To(Equal("some_label=fun"))
				})
			})
		})
	})
})
