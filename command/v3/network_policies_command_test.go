package v3_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/cfnetworkingaction"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v3"
	"code.cloudfoundry.org/cli/command/v3/v3fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("network-policies Command", func() {
	var (
		cmd             NetworkPoliciesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v3fakes.FakeNetworkPoliciesActor
		binaryName      string
		executeErr      error
		srcApp          string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v3fakes.FakeNetworkPoliciesActor)

		srcApp = ""

		cmd = NetworkPoliciesCommand{
			UI:          testUI,
			SourceApp:   srcApp,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Context("when checking target fails", func() {
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

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(configv3.User{Name: "some-user"}, nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "some-space", GUID: "some-space-guid"})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "some-org"})
		})

		It("outputs flavor text", func() {
			Expect(testUI.Out).To(Say(`Listing network policies in org some-org / space some-space as some-user\.\.\.`))
		})

		Context("when fetching the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("some-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("some-error"))
			})
		})

		Context("when listing policies is successful", func() {
			BeforeEach(func() {
				fakeActor.NetworkPoliciesBySpaceReturns([]cfnetworkingaction.Policy{
					{
						SourceName:      "app1",
						DestinationName: "app2",
						Protocol:        "tcp",
						StartPort:       8080,
						EndPort:         8080,
					}, {
						SourceName:      "app2",
						DestinationName: "app1",
						Protocol:        "udp",
						StartPort:       1234,
						EndPort:         2345,
					},
				}, cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, nil)
			})

			It("lists the policies when no error occurs", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(fakeActor.NetworkPoliciesBySpaceCallCount()).To(Equal(1))
				passedSpaceGuid := fakeActor.NetworkPoliciesBySpaceArgsForCall(0)
				Expect(passedSpaceGuid).To(Equal("some-space-guid"))

				Expect(testUI.Out).To(Say(`Listing network policies in org some-org / space some-space as some-user\.\.\.`))
				Expect(testUI.Out).To(Say("\n\n"))
				Expect(testUI.Out).To(Say("source\\s+destination\\s+protocol\\s+ports"))
				Expect(testUI.Out).To(Say("app1\\s+app2\\s+tcp\\s+8080[^-]"))
				Expect(testUI.Out).To(Say("app2\\s+app1\\s+udp\\s+1234-2345"))

				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})

			Context("when a source app name is passed", func() {
				BeforeEach(func() {
					cmd.SourceApp = "some-app"
					fakeActor.NetworkPoliciesBySpaceAndAppNameReturns([]cfnetworkingaction.Policy{
						{
							SourceName:      "app1",
							DestinationName: "app2",
							Protocol:        "tcp",
							StartPort:       8080,
							EndPort:         8080,
						}, {
							SourceName:      "app2",
							DestinationName: "app1",
							Protocol:        "udp",
							StartPort:       1234,
							EndPort:         2345,
						},
					}, cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, nil)
				})

				It("lists the policies when no error occurs", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(fakeActor.NetworkPoliciesBySpaceAndAppNameCallCount()).To(Equal(1))
					passedSpaceGuid, passedSrcAppName := fakeActor.NetworkPoliciesBySpaceAndAppNameArgsForCall(0)
					Expect(passedSpaceGuid).To(Equal("some-space-guid"))
					Expect(passedSrcAppName).To(Equal("some-app"))

					Expect(testUI.Out).To(Say(`Listing network policies of app %s in org some-org / space some-space as some-user\.\.\.`, cmd.SourceApp))
					Expect(testUI.Out).To(Say("\n\n"))
					Expect(testUI.Out).To(Say("source\\s+destination\\s+protocol\\s+ports"))
					Expect(testUI.Out).To(Say("app1\\s+app2\\s+tcp\\s+8080[^-]"))
					Expect(testUI.Out).To(Say("app2\\s+app1\\s+udp\\s+1234-2345"))

					Expect(testUI.Err).To(Say("some-warning-1"))
					Expect(testUI.Err).To(Say("some-warning-2"))
				})
			})
		})

		Context("when listing the policies is not successful", func() {
			BeforeEach(func() {
				fakeActor.NetworkPoliciesBySpaceReturns([]cfnetworkingaction.Policy{}, cfnetworkingaction.Warnings{"some-warning-1", "some-warning-2"}, actionerror.ApplicationNotFoundError{Name: srcApp})
			})

			It("displays warnings and returns the error", func() {
				Expect(executeErr).To(MatchError(actionerror.ApplicationNotFoundError{Name: srcApp}))

				Expect(testUI.Out).To(Say(`Listing network policies in org some-org / space some-space as some-user\.\.\.`))
				Expect(testUI.Err).To(Say("some-warning-1"))
				Expect(testUI.Err).To(Say("some-warning-2"))
			})
		})
	})
})
