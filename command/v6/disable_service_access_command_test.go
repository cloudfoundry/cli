package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("disable-service-access Command", func() {
	var (
		cmd             DisableServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeDisableServiceAccessActor
		binaryName      string
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeDisableServiceAccessActor)
		extraArgs = nil

		cmd = DisableServiceAccessCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			RequiredArgs: flag.Service{
				Service: "some-service",
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	When("the user provides arguments", func() {
		BeforeEach(func() {
			extraArgs = []string{"some-extra-arg"}
		})

		It("fails with a TooManyArgumentsError", func() {
			Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
				ExtraArgument: "some-extra-arg",
			}))
		})
	})

	When("a cloud controller API endpoint is set", func() {
		BeforeEach(func() {
			fakeConfig.TargetReturns("some-url")
		})

		When("checking target fails", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
			})

			It("returns a not logged in error", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
			})
		})

		When("the user is logged in", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "admin"},
					nil)
				fakeSharedActor.IsLoggedInReturns(true)
			})

			When("no flags are passed", func() {
				When("disabling access to the service succeeds", func() {
					BeforeEach(func() {
						fakeActor.DisableServiceForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, nil)
					})

					It("passes on the service", func() {
						Expect(fakeActor.DisableServiceForAllOrgsCallCount()).To(Equal(1))
						service := fakeActor.DisableServiceForAllOrgsArgsForCall(0)
						Expect(service).To(Equal("some-service"))
					})

					It("displays informative success message", func() {
						Expect(testUI.Out).To(Say("Disabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})

				When("disabling access to the service fails", func() {
					BeforeEach(func() {
						fakeActor.DisableServiceForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Disabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
						Expect(executeErr).To(MatchError("explode"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})
			})

			When("the -p flag is passed", func() {
				BeforeEach(func() {
					cmd.ServicePlan = "some-plan"
				})

				When("disabling access to the plan succeeds", func() {
					BeforeEach(func() {
						fakeActor.DisablePlanForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, nil)
					})

					It("passes on the service plan", func() {
						Expect(fakeActor.DisablePlanForAllOrgsCallCount()).To(Equal(1))
						service, plan := fakeActor.DisablePlanForAllOrgsArgsForCall(0)
						Expect(service).To(Equal("some-service"))
						Expect(plan).To(Equal("some-plan"))
					})

					It("displays an informative success message", func() {
						Expect(testUI.Out).To(Say("Disabling access of plan some-plan for service some-service as admin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})

				When("disabling access to the plan fails", func() {
					BeforeEach(func() {
						fakeActor.DisablePlanForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Disabling access of plan some-plan for service some-service as admin\\.\\.\\."))
						Expect(executeErr).To(MatchError("explode"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})
			})

			When("the -o flag is passed", func() {
				BeforeEach(func() {
					cmd.Organization = "some-org"
				})

				When("disabling access to the org succeeds", func() {
					BeforeEach(func() {
						fakeActor.DisableServiceForOrgReturns(v2action.Warnings{"warning", "second-warning"}, nil)
					})

					It("passes on the organization name", func() {
						Expect(fakeActor.DisableServiceForOrgCallCount()).To(Equal(1))
						service, org := fakeActor.DisableServiceForOrgArgsForCall(0)
						Expect(service).To(Equal("some-service"))
						Expect(org).To(Equal("some-org"))
					})

					It("displays an informative success message", func() {
						Expect(testUI.Out).To(Say("Disabling access to all plans of service some-service for the org some-org as admin..."))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})

				When("disabling access to the org fails", func() {
					BeforeEach(func() {
						fakeActor.DisableServiceForOrgReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Disabling access to all plans of service some-service for the org some-org as admin..."))
						Expect(executeErr).To(MatchError("explode"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})
			})

			When("the -p and -o flags are passed", func() {
				BeforeEach(func() {
					cmd.Organization = "some-org"
					cmd.ServicePlan = "some-plan"
				})

				When("disabling access to the plan in the org succeeds", func() {
					BeforeEach(func() {
						fakeActor.DisablePlanForOrgReturns(v2action.Warnings{"warning", "second-warning"}, nil)
					})

					It("passes on the plan and organization name", func() {
						Expect(fakeActor.DisablePlanForOrgCallCount()).To(Equal(1))
						service, plan, org := fakeActor.DisablePlanForOrgArgsForCall(0)

						Expect(service).To(Equal("some-service"))
						Expect(plan).To(Equal("some-plan"))
						Expect(org).To(Equal("some-org"))
					})

					It("displays an informative success message", func() {
						Expect(testUI.Out).To(Say("Disabling access to plan some-plan of service some-service for org some-org as admin..."))
						Expect(testUI.Out).To(Say("OK"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})

				When("disabling access to the plan in the org fails", func() {
					BeforeEach(func() {
						fakeActor.DisablePlanForOrgReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Disabling access to plan some-plan of service some-service for org some-org as admin..."))
						Expect(executeErr).To(MatchError("explode"))
					})

					It("returns the warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
						Expect(testUI.Err).To(Say("second-warning"))
					})
				})
			})
		})
	})
})
