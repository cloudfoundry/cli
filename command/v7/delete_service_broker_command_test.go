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

var _ = Describe("delete-service-broker Command", func() {
	var (
		cmd               DeleteServiceBrokerCommand
		testUI            *ui.UI
		fakeConfig        *commandfakes.FakeConfig
		fakeSharedActor   *commandfakes.FakeSharedActor
		fakeActor         *v7fakes.FakeActor
		input             *Buffer
		binaryName        string
		executeErr        error
		serviceBrokerName string
		serviceBrokerGUID string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		serviceBrokerName = "some-service-broker"
		serviceBrokerGUID = "service-broker-guid"

		cmd = DeleteServiceBrokerCommand{
			BaseCommand: BaseCommand{UI: testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, serviceBrokerName)

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{
				BinaryName: binaryName,
			})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the -f flag is NOT provided", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		When("the user inputs yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).ToNot(HaveOccurred())

				fakeActor.GetServiceBrokerByNameReturns(v7action.ServiceBroker{Name: serviceBrokerName, GUID: serviceBrokerGUID}, v7action.Warnings{"get-broker-by-name-warning"}, nil)
				fakeActor.DeleteServiceBrokerReturns(v7action.Warnings{"delete-broker-warning"}, nil)
			})

			It("delegates to the Actor", func() {
				actualServiceBrokerGUID := fakeActor.DeleteServiceBrokerArgsForCall(0)
				Expect(actualServiceBrokerGUID).To(Equal(serviceBrokerGUID))
			})

			It("deletes the service broker", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(`Deleting service broker %s...`, serviceBrokerName))
				Expect(testUI.Err).To(Say("get-broker-by-name-warning"))
				Expect(testUI.Err).To(Say("delete-broker-warning"))
				Expect(testUI.Out).To(Say("OK"))
				Expect(testUI.Out).NotTo(Say("ServiceBroker 'service-broker' does not exist"))
			})
		})

		When("the user inputs no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("'%s' has not been deleted.", serviceBrokerName))
				Expect(fakeActor.DeleteServiceBrokerCallCount()).To(Equal(0))
			})
		})

		When("the user chooses the default", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("cancels the delete", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("'%s' has not been deleted.", serviceBrokerName))
				Expect(fakeActor.DeleteServiceBrokerCallCount()).To(Equal(0))
			})
		})

		When("the user input is invalid", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("e\n\n"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("asks the user again", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`Really delete the service broker some\-service\-broker\? \[yN\]`))
				Expect(testUI.Out).To(Say(`invalid input \(not y, n, yes, or no\)`))
				Expect(testUI.Out).To(Say(`Really delete the service broker some\-service\-broker\? \[yN\]`))
				Expect(fakeActor.DeleteServiceBrokerCallCount()).To(Equal(0))
			})
		})
	})

	When("the -f flag is provided", func() {
		BeforeEach(func() {
			cmd.Force = true
		})

		When("deleting the service broker errors", func() {
			Context("generic error", func() {
				BeforeEach(func() {
					fakeActor.DeleteServiceBrokerReturns(v7action.Warnings{"some-warning"}, errors.New("some-error"))
				})

				It("displays all warnings, and returns the error", func() {
					Expect(testUI.Err).To(Say("some-warning"))
					Expect(testUI.Out).To(Say(`Deleting service broker some\-service\-broker\.\.\.`))
					Expect(testUI.Out).ToNot(Say("OK"))
					Expect(executeErr).To(MatchError("some-error"))
				})
			})
		})

		When("the service broker doesn't exist", func() {
			BeforeEach(func() {
				fakeActor.GetServiceBrokerByNameReturns(
					v7action.ServiceBroker{},
					v7action.Warnings{"some-warning"},
					actionerror.ServiceBrokerNotFoundError{
						Name: serviceBrokerName},
				)
			})

			It("displays all warnings, that the domain wasn't found, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting service broker %s...`, serviceBrokerName))
				Expect(testUI.Out).To(Say(`Service broker '%s' does not exist.`, serviceBrokerName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})

		When("the service broker exists", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceBrokerReturns(v7action.Warnings{"some-warning"}, nil)
			})

			It("displays all warnings, and does not error", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say(`Deleting service broker %s...`, serviceBrokerName))
				Expect(testUI.Out).To(Say("OK"))
			})
		})
	})
})
