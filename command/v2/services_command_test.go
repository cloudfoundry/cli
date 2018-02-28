package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("services Command", func() {
	var (
		cmd             ServicesCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeServicesActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeServicesActor)

		cmd = ServicesCommand{
			UI:          testUI,
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

	Context("when an error is encountered checking if the environment is setup correctly", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrgArg).To(BeTrue())
			Expect(checkTargetedSpaceArg).To(BeTrue())
		})
	})

	Context("when the user is logged in and an org and space are targeted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
			})
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				GUID: "some-space-guid",
				Name: "some-space",
			})
		})

		Context("when getting the current user fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some-error that happened")
				fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})
		})

		Context("when getting the current user succeeds", func() {
			var (
				fakeUser configv3.User
			)

			BeforeEach(func() {
				fakeUser = configv3.User{Name: "some-user"}
				fakeConfig.CurrentUserReturns(fakeUser, nil)
				fakeConfig.TargetedSpaceReturns(configv3.Space{
					GUID: "some-space-guid",
					Name: "some-space",
				})
				fakeConfig.TargetedOrganizationReturns(configv3.Organization{
					Name: "some-org",
				})
			})

			Context("when there are no services", func() {
				BeforeEach(func() {
					fakeActor.GetServiceInstancesSummaryBySpaceReturns(
						nil,
						v2action.Warnings{"get-summary-warnings"},
						nil,
					)
				})

				It("displays that there are no services", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Getting services in org %s / space %s as %s...", "some-org",
						"some-space", fakeUser.Name))

					out := testUI.Out.(*Buffer).Contents()
					Expect(out).To(MatchRegexp("No services found"))
					Expect(out).ToNot(MatchRegexp("name\\s+service\\s+plan\\s+bound apps\\s+last operation"))
					Expect(testUI.Err).To(Say("get-summary-warnings"))
				})
			})

			Context("when there are services", func() {
				BeforeEach(func() {
					fakeActor.GetServiceInstancesSummaryBySpaceReturns(
						[]v2action.ServiceInstanceSummary{
							{
								ServiceInstance: v2action.ServiceInstance{
									Name: "instance-1",
									LastOperation: ccv2.LastOperation{
										Type:  "some-type",
										State: "some-state",
									},
									Type: constant.ServiceInstanceTypeManagedService,
								},
								ServicePlan:       v2action.ServicePlan{Name: "some-plan"},
								Service:           v2action.Service{Label: "some-service-1"},
								BoundApplications: []string{"app-1", "app-2"},
							},
							{
								ServiceInstance: v2action.ServiceInstance{
									Name: "instance-2",
									Type: constant.ServiceInstanceTypeManagedService,
								},
								Service: v2action.Service{Label: "some-service-2"},
							},
							{
								ServiceInstance: v2action.ServiceInstance{
									Name: "instance-3",
									Type: constant.ServiceInstanceTypeUserProvidedService,
								},
							},
						},
						v2action.Warnings{"get-summary-warnings"},
						nil,
					)
				})

				It("displays all the services in the org & space & warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Out).To(Say("Getting services in org %s / space %s as %s...", "some-org",
						"some-space", fakeUser.Name))
					Expect(testUI.Out).To(Say("name\\s+service\\s+plan\\s+bound apps\\s+last operation"))
					Expect(testUI.Out).To(Say("instance-1\\s+some-service-1\\s+some-plan\\s+app-1, app-2\\s+some-type some-state"))
					Expect(testUI.Out).To(Say("instance-2\\s+some-service-2\\s+"))
					Expect(testUI.Out).To(Say("instance-3\\s+user-provided\\s+"))
					Expect(testUI.Err).To(Say("get-summary-warnings"))
				})
			})
		})
	})
})
