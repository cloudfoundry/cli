package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("service-brokers Command", func() {
	var (
		cmd             *v7.ServiceBrokersCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeServiceBrokersActor
		input           *Buffer
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeServiceBrokersActor)

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		cmd = &v7.ServiceBrokersCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeFalse())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the user is logged in and a space is targetted", func() {
		BeforeEach(func() {
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: "some-org",
				GUID: "some-org-guid",
			})

			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: "some-space",
				GUID: "some-space-guid",
			})

			fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
		})

		It("displays a message with the username", func() {
			Expect(testUI.Out).To(Say("Getting service brokers as %s...", "steve"))
		})

		It("calls the GetServiceBrokersActor", func() {
			Expect(fakeActor.GetServiceBrokersCallCount()).To(Equal(1))
		})

		When("there are no service brokers", func() {
			BeforeEach(func() {
				fakeActor.GetServiceBrokersReturns([]v7action.ServiceBroker{}, v7action.Warnings{"service-broker-warnings"}, nil)
			})

			It("says there are no service brokers", func() {
				Expect(testUI.Out).To(Say("No service brokers found"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
				Expect(testUI.Out).NotTo(Say("name\\s+url"), "printing table header when table is empty")
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("there is one service broker", func() {
			BeforeEach(func() {
				serviceBrokers := []v7action.ServiceBroker{
					{Name: "foo", URL: "http://foo.url", GUID: "guid-foo"},
				}
				fakeActor.GetServiceBrokersReturns(serviceBrokers, v7action.Warnings{"service-broker-warnings"}, nil)
			})

			It("prints a table header and the broker details", func() {
				Expect(testUI.Out).To(Say("name\\s+url"))
				Expect(testUI.Out).To(Say("foo\\s+http://foo.url"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("there are many service brokers", func() {
			BeforeEach(func() {
				serviceBrokers := []v7action.ServiceBroker{
					{Name: "foo", URL: "http://foo.url", GUID: "guid-foo"},
					{Name: "bar", URL: "https://bar.com", GUID: "guid-bar"},
				}
				fakeActor.GetServiceBrokersReturns(serviceBrokers, v7action.Warnings{"service-broker-warnings"}, nil)
			})

			It("prints a table header and the broker details", func() {
				Expect(testUI.Out).To(Say("name\\s+url"))
				Expect(testUI.Out).To(Say("foo\\s+http://foo.url"))
				Expect(testUI.Out).To(Say("bar\\s+https://bar.com"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("calling the GetServiceBrokersActor returns an error", func() {
			BeforeEach(func() {
				fakeActor.GetServiceBrokersReturns([]v7action.ServiceBroker{}, v7action.Warnings{"service-broker-warnings"}, errors.New("fake service-brokers error"))
			})

			It("prints the error and warnings", func() {
				Expect(executeErr).To(MatchError("fake service-brokers error"))
				Expect(testUI.Err).To(Say("service-broker-warnings"))
			})
		})
	})
})
