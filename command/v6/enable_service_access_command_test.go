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

var _ = Describe("enable-service-access command", func() {
	var (
		cmd             EnableServiceAccessCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeEnableServiceAccessActor
		binaryName      string
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeEnableServiceAccessActor)
		extraArgs = nil

		cmd = EnableServiceAccessCommand{
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
			})

			When("no flags are passed", func() {
				BeforeEach(func() {
					fakeActor.EnableServiceForAllOrgsReturns(nil, nil)
				})

				It("passes on the service", func() {
					Expect(fakeActor.EnableServiceForAllOrgsCallCount()).To(Equal(1))
					service := fakeActor.EnableServiceForAllOrgsArgsForCall(0)
					Expect(service).To(Equal("some-service"))
				})

				It("displays informative success message", func() {
					Expect(testUI.Out).To(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for the org some-org as admin..."))
					Expect(testUI.Out).NotTo(Say("Enabling access to plan some-plan of service some-service for org some-org as admin..."))
					Expect(testUI.Out).To(Say("OK"))
				})

				When("enabling access to the service fails", func() {
					BeforeEach(func() {
						fakeActor.EnableServiceForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
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
					fakeActor.EnablePlanForAllOrgsReturns(nil, nil)
				})

				It("passes on the service plan", func() {
					Expect(fakeActor.EnablePlanForAllOrgsCallCount()).To(Equal(1))
					service, plan := fakeActor.EnablePlanForAllOrgsArgsForCall(0)
					Expect(service).To(Equal("some-service"))
					Expect(plan).To(Equal("some-plan"))
				})

				It("displays an informative success message", func() {
					Expect(testUI.Out).To(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for the org some-org as admin..."))
					Expect(testUI.Out).NotTo(Say("Enabling access to plan some-plan of service some-service for org some-org as admin..."))
					Expect(testUI.Out).To(Say("OK"))
				})

				When("enabling access to the plan fails", func() {
					BeforeEach(func() {
						fakeActor.EnablePlanForAllOrgsReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
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
					fakeActor.EnableServiceForOrgReturns(nil, nil)
				})

				It("passes on the organization name", func() {
					Expect(fakeActor.EnableServiceForOrgCallCount()).To(Equal(1))
					service, org := fakeActor.EnableServiceForOrgArgsForCall(0)
					Expect(service).To(Equal("some-service"))
					Expect(org).To(Equal("some-org"))
				})

				It("displays an informative success message", func() {
					Expect(testUI.Out).To(Say("Enabling access to all plans of service some-service for the org some-org as admin..."))
					Expect(testUI.Out).NotTo(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to plan some-plan of service some-service for org some-org as admin..."))
					Expect(testUI.Out).To(Say("OK"))
				})

				When("enabling access to the org fails", func() {
					BeforeEach(func() {
						fakeActor.EnableServiceForOrgReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Enabling access to all plans of service some-service for the org some-org as admin..."))
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
					fakeActor.EnablePlanForOrgReturns(nil, nil)
				})

				It("passes on the plan and organization name", func() {
					Expect(fakeActor.EnablePlanForOrgCallCount()).To(Equal(1))
					service, plan, org := fakeActor.EnablePlanForOrgArgsForCall(0)

					Expect(service).To(Equal("some-service"))
					Expect(plan).To(Equal("some-plan"))
					Expect(org).To(Equal("some-org"))
				})

				It("displays an informative success message", func() {
					Expect(testUI.Out).To(Say("Enabling access to plan some-plan of service some-service for org some-org as admin..."))
					Expect(testUI.Out).NotTo(Say("Enabling access of plan some-plan for service some-service as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for all orgs as admin\\.\\.\\."))
					Expect(testUI.Out).NotTo(Say("Enabling access to all plans of service some-service for the org some-org as admin..."))
					Expect(testUI.Out).To(Say("OK"))
				})

				When("enabling access to the plan in the org fails", func() {
					BeforeEach(func() {
						fakeActor.EnablePlanForOrgReturns(v2action.Warnings{"warning", "second-warning"}, errors.New("explode"))
					})

					It("returns the error", func() {
						Expect(testUI.Out).To(Say("Enabling access to plan some-plan of service some-service for org some-org as admin..."))
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
