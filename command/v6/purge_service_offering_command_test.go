package v6_test

import (
	"fmt"

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

var _ = Describe("purge-service-offering command", func() {
	var (
		cmd                   PurgeServiceOfferingCommand
		testUI                *ui.UI
		fakeConfig            *commandfakes.FakeConfig
		fakeSharedActor       *commandfakes.FakeSharedActor
		fakePurgeServiceActor *v6fakes.FakePurgeServiceOfferingActor
		input                 *Buffer
		binaryName            string
		executeErr            error
		extraArgs             []string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakePurgeServiceActor = new(v6fakes.FakePurgeServiceOfferingActor)
		extraArgs = nil

		cmd = PurgeServiceOfferingCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakePurgeServiceActor,
			RequiredArgs: flag.Service{
				Service: "some-service",
			},
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	When("the user provides extra arguments", func() {
		BeforeEach(func() {
			extraArgs = []string{"some-extra-arg"}
		})

		It("fails with a TooManyArgumentsError", func() {
			Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
				ExtraArgument: "some-extra-arg",
			}))
		})
	})

	When("cloud controller API endpoint is set", func() {
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
				It("prints the warning text", func() {
					Expect(testUI.Out).To(Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."))
					Expect(testUI.Out).To(Say("Really purge service offering some-service from Cloud Foundry?"))
				})

				When("the user chooses the default", func() {
					BeforeEach(func() {
						input.Write([]byte("\n"))
					})

					It("does not purge the service offering", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Purge service offering cancelled"))

						Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(0))
					})
				})

				When("the user inputs no", func() {
					BeforeEach(func() {
						input.Write([]byte("n\n"))
					})

					It("does not purge the service offering", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Purge service offering cancelled"))

						Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(0))
					})
				})

				When("the user inputs yes", func() {
					BeforeEach(func() {
						input.Write([]byte("y\n"))
						fakePurgeServiceActor.GetServiceByNameAndBrokerNameReturns(v2action.Service{
							Label: "some-service",
							GUID:  "some-service-guid",
						}, v2action.Warnings{"get-service-warning"}, nil)
						fakePurgeServiceActor.PurgeServiceOfferingReturns(v2action.Warnings{"warning-1"}, nil)
					})

					It("purges the service offering and prints all warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(1))

						service := fakePurgeServiceActor.PurgeServiceOfferingArgsForCall(0)
						Expect(service.Label).To(Equal("some-service"))
						Expect(service.GUID).To(Equal("some-service-guid"))

						Expect(testUI.Err).To(Say("get-service-warning"))
						Expect(testUI.Err).To(Say("warning-1"))
						Expect(testUI.Out).To(Say("OK"))
					})

					When("purge service offering fails", func() {
						When("the service offering does not exist", func() {
							BeforeEach(func() {
								fakePurgeServiceActor.GetServiceByNameAndBrokerNameReturns(v2action.Service{}, v2action.Warnings{"get-service-warning"}, actionerror.ServiceNotFoundError{Name: "some-service"})
							})

							It("returns a PurgeServiceOfferingNotFound error and prints the warnings, and OK", func() {
								Expect(fakePurgeServiceActor.GetServiceByNameAndBrokerNameCallCount()).To(Equal(1))

								serviceName, brokerName := fakePurgeServiceActor.GetServiceByNameAndBrokerNameArgsForCall(0)
								Expect(serviceName).To(Equal("some-service"))
								Expect(brokerName).To(Equal(""))

								Eventually(testUI.Out).Should(Say(`Service offering 'some-service' not found`))
								Eventually(testUI.Out).Should(Say(`OK`))
								Expect(testUI.Err).To(Say("get-service-warning"))
							})
						})

						When("an unknown error occurs", func() {
							BeforeEach(func() {
								fakePurgeServiceActor.PurgeServiceOfferingReturns(v2action.Warnings{"warning-1"}, fmt.Errorf("it broke!"))
							})

							It("returns the error and prints the warnings", func() {
								Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(1))

								Expect(executeErr).To(MatchError(fmt.Errorf("it broke!")))
								Expect(testUI.Err).To(Say("get-service-warning"))
								Expect(testUI.Err).To(Say("warning-1"))
							})
						})
					})
				})

				When("the user input is invalid", func() {
					BeforeEach(func() {
						input.Write([]byte("e\n\n"))
					})

					It("asks the user again", func() {
						Expect(executeErr).NotTo(HaveOccurred())

						Expect(testUI.Out).To(Say(`Really purge service offering some-service from Cloud Foundry\? \[yN\]:`))
						Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
						Expect(testUI.Out).To(Say(`Really purge service offering some-service from Cloud Foundry\? \[yN\]:`))

						Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(0))
					})
				})
			})

			When("the -f flag is passed", func() {
				BeforeEach(func() {
					cmd.Force = true

					fakePurgeServiceActor.GetServiceByNameAndBrokerNameReturns(v2action.Service{
						Label: "some-service",
						GUID:  "some-service-guid",
					}, v2action.Warnings{"get-service-warning"}, nil)
					fakePurgeServiceActor.PurgeServiceOfferingReturns(v2action.Warnings{"warning-1"}, nil)
				})

				It("purges the service offering without asking for confirmation", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(testUI.Out).NotTo(Say(`Really purge service offering some-service from Cloud Foundry\? \[yN\]:`))

					Expect(testUI.Out).To(Say(`Purging service some-service\.\.\.`))

					Expect(fakePurgeServiceActor.PurgeServiceOfferingCallCount()).To(Equal(1))
					service := fakePurgeServiceActor.PurgeServiceOfferingArgsForCall(0)
					Expect(service.Label).To(Equal("some-service"))
					Expect(service.GUID).To(Equal("some-service-guid"))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Out).To(Say("OK"))
				})
			})

			When("the -p flag is passed", func() {
				BeforeEach(func() {
					cmd.Provider = "dave"
					input.Write([]byte("y\n"))
				})

				It("returns an error that this flag is no longer supported", func() {
					Expect(executeErr).To(MatchError(translatableerror.FlagNoLongerSupportedError{Flag: "-p"}))
				})
			})
		})
	})
})
