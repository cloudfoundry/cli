package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service Command", func() {
	var (
		cmd             *v7.CreateServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
		expectedError   error
	)

	const (
		fakeUserName                 = "fake-user-name"
		requestedServiceInstanceName = "service-instance-name"
		fakeOrgName                  = "fake-org-name"
		fakeSpaceName                = "fake-space-name"
		fakeSpaceGUID                = "fake-space-guid"
		requestedPlanName            = "coolPlan"
		requestedOfferingName        = "coolOffering"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = &v7.CreateServiceCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(cmd, requestedOfferingName, requestedPlanName, requestedServiceInstanceName)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		org, space := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(org).To(BeTrue())
		Expect(space).To(BeTrue())
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})

	Context("When logged in and targeting a space", func() {
		BeforeEach(func() {
			fakeConfig.TargetedSpaceReturns(configv3.Space{
				Name: fakeSpaceName,
				GUID: fakeSpaceGUID,
			})

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{
				Name: fakeOrgName,
			})

			fakeConfig.CurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

			fakeActor.GetServicePlanByNameOfferingAndBrokerReturns(
				resources.ServicePlan{},
				v7action.Warnings{"be warned", "take care"},
				nil,
			)
		})

		It("prints a message and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say("Creating service instance %s in org %s / space %s as %s...", requestedServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName),
				Say("OK"),
			))

			Expect(testUI.Err).To(SatisfyAll(
				Say("be warned"),
				Say("take care"),
			))
		})

		It("Calls the client with the right arguments", func() {
			Expect(fakeActor.GetServicePlanByNameOfferingAndBrokerCallCount()).To(Equal(1))
			planName, offeringName, brokerName := fakeActor.GetServicePlanByNameOfferingAndBrokerArgsForCall(0)
			Expect(planName).To(Equal(requestedPlanName))
			Expect(offeringName).To(Equal(requestedOfferingName))
			Expect(brokerName).To(BeEmpty())
		})

		When("requesting from a specific broker", func() {
			var requestedBrokerName string

			BeforeEach(func() {
				requestedBrokerName = "aCoolBroker"
				setFlag(cmd, "-b", requestedBrokerName)
			})

			It("passes the right parameters to the actor", func() {
				Expect(executeErr).To(Not(HaveOccurred()))

				Expect(fakeActor.GetServicePlanByNameOfferingAndBrokerCallCount()).To(Equal(1))
				planName, offeringName, brokerName := fakeActor.GetServicePlanByNameOfferingAndBrokerArgsForCall(0)
				Expect(planName).To(Equal(requestedPlanName))
				Expect(offeringName).To(Equal(requestedOfferingName))
				Expect(brokerName).To(Equal(requestedBrokerName))
			})
		})

		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: fakeUserName}, errors.New("boom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})

		When("client returns an error", func() {
			BeforeEach(func() {
				expectedError = actionerror.ServicePlanNotFoundError{PlanName: requestedPlanName}
				fakeActor.GetServicePlanByNameOfferingAndBrokerReturns(
					resources.ServicePlan{},
					v7action.Warnings{"warning one", "warning two"},
					expectedError,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(HaveOccurred())

				Expect(executeErr).To(MatchError(expectedError))
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Out).NotTo(Say("OK"))

				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})
	})
})
