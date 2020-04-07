package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("purge-service-offering command", func() {
	var (
		cmd             PurgeServiceOfferingCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor

		input      *Buffer
		executeErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = PurgeServiceOfferingCommand{
			RequiredArgs:  flag.Service{ServiceOffering: "fake-service-offering"},
			ServiceBroker: "fake-service-broker",
			Force:         true,
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				Actor:       fakeActor,
				SharedActor: fakeSharedActor,
			},
		}

		fakeActor.PurgeServiceOfferingByNameAndBrokerReturns(v7action.Warnings{"a warning"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the target", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		actualOrgRequired, actualSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(actualOrgRequired).To(BeFalse())
		Expect(actualSpaceRequired).To(BeFalse())
	})

	It("calls the actor with the right arguments", func() {
		Expect(fakeActor.PurgeServiceOfferingByNameAndBrokerCallCount()).To(Equal(1))
		actualOffering, actualBroker := fakeActor.PurgeServiceOfferingByNameAndBrokerArgsForCall(0)
		Expect(actualOffering).To(Equal("fake-service-offering"))
		Expect(actualBroker).To(Equal("fake-service-broker"))
	})

	It("prints messages and warnings", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(testUI.Out).To(Say(`Purging service offering fake-service-offering\.\.\.`))
		Expect(testUI.Out).To(Say("OK"))

		Expect(testUI.Err).To(Say("a warning"))
	})

	When("the -f (force) flag is not specified", func() {
		BeforeEach(func() {
			cmd.Force = false
		})

		It("prints a warning", func() {
			Expect(testUI.Out).To(Say(`WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\. All knowledge of the service will be removed from Cloud Foundry, including service instances and service bindings\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\.`))
			Expect(testUI.Out).To(Say("Really purge service offering fake-service-offering from broker fake-service-broker from Cloud Foundry?"))
		})

		When("the service broker name is not specified", func() {
			BeforeEach(func() {
				cmd.ServiceBroker = ""
			})

			It("prints a message that does not include the service broker name", func() {
				Expect(testUI.Out).To(Say("Really purge service offering fake-service-offering from Cloud Foundry?"))
			})
		})

		When("the user chooses the default", func() {
			BeforeEach(func() {
				input.Write([]byte("\n"))
			})

			It("does not purge the service offering", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`Purge service offering cancelled\.`))
				Expect(fakeActor.PurgeServiceOfferingByNameAndBrokerCallCount()).To(Equal(0))
			})
		})

		When("the user chooses `no`", func() {
			BeforeEach(func() {
				input.Write([]byte("n\n"))
			})

			It("does not purge the service offering", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`Purge service offering cancelled\.`))
				Expect(fakeActor.PurgeServiceOfferingByNameAndBrokerCallCount()).To(Equal(0))
			})
		})

		When("the user chooses `yes`", func() {
			BeforeEach(func() {
				input.Write([]byte("y\n"))
			})

			It("purges the service offering", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`Purging service offering fake-service-offering\.\.\.`))
				Expect(fakeActor.PurgeServiceOfferingByNameAndBrokerCallCount()).To(Equal(1))
			})
		})
	})

	When("the actor fails", func() {
		BeforeEach(func() {
			fakeActor.PurgeServiceOfferingByNameAndBrokerReturns(v7action.Warnings{"actor warning"}, errors.New("fake actor error"))
		})

		It("returns an error", func() {
			Expect(testUI.Err).To(Say("actor warning"))
			Expect(executeErr).To(MatchError("fake actor error"))
		})
	})

	When("checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("fake target error"))
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError("fake target error"))
		})
	})
})
