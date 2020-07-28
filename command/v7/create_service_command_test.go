package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service Command", func() {
	var (
		cmd             v7.CreateServiceCommand
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

		cmd = v7.CreateServiceCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, requestedOfferingName, requestedPlanName, requestedServiceInstanceName)
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

		})

		It("prints a message and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say("Creating service instance %s in org %s / space %s as %s...", requestedServiceInstanceName, fakeOrgName, fakeSpaceName, fakeUserName),
				Say("OK"),
			))
		})

		It("Calls the actor with the right arguments", func() {
			Expect(fakeActor.CreateManagedServiceInstanceCallCount()).To(Equal(1))
			params := fakeActor.CreateManagedServiceInstanceArgsForCall(0)
			Expect(params.ServicePlanName).To(Equal(requestedPlanName))
			Expect(params.ServiceOfferingName).To(Equal(requestedOfferingName))
			Expect(params.ServiceBrokerName).To(BeEmpty())
			Expect(params.ServiceInstanceName).To(Equal(requestedServiceInstanceName))
			Expect(params.SpaceGUID).To(Equal(fakeSpaceGUID))
		})

		When("requesting from a specific broker", func() {
			var requestedBrokerName string

			BeforeEach(func() {
				requestedBrokerName = "aCoolBroker"
				setFlag(&cmd, "-b", requestedBrokerName)
			})

			It("passes the right parameters to the actor", func() {
				Expect(executeErr).To(Not(HaveOccurred()))

				Expect(fakeActor.CreateManagedServiceInstanceCallCount()).To(Equal(1))
				params := fakeActor.CreateManagedServiceInstanceArgsForCall(0)
				Expect(params.ServicePlanName).To(Equal(requestedPlanName))
				Expect(params.ServiceOfferingName).To(Equal(requestedOfferingName))
				Expect(params.ServiceBrokerName).To(Equal(requestedBrokerName))
				Expect(params.ServiceInstanceName).To(Equal(requestedServiceInstanceName))
				Expect(params.SpaceGUID).To(Equal(fakeSpaceGUID))
			})
		})

		When("there are user provided tags", func() {
			var requestedTags types.OptionalStringSlice

			BeforeEach(func() {
				requestedTags = types.NewOptionalStringSlice("tag-1", "tag-2")
				setFlag(&cmd, "-t", requestedTags)
			})

			It("passes the right parameters to the actor", func() {
				Expect(executeErr).To(Not(HaveOccurred()))

				Expect(fakeActor.CreateManagedServiceInstanceCallCount()).To(Equal(1))
				params := fakeActor.CreateManagedServiceInstanceArgsForCall(0)
				Expect(params.ServicePlanName).To(Equal(requestedPlanName))
				Expect(params.ServiceOfferingName).To(Equal(requestedOfferingName))
				Expect(params.ServiceInstanceName).To(Equal(requestedServiceInstanceName))
				Expect(params.SpaceGUID).To(Equal(fakeSpaceGUID))
				Expect(params.Tags).To(Equal(requestedTags))

			})
		})

		When("there are parameters", func() {
			var requestedParams map[string]interface{}

			BeforeEach(func() {
				requestedParams = map[string]interface{}{"param-1": "value-1", "param-2": "value-2"}
				setFlag(&cmd, "-c", types.NewOptionalObject(requestedParams))
			})

			It("passes the right parameters to the actor", func() {
				Expect(executeErr).To(Not(HaveOccurred()))

				Expect(fakeActor.CreateManagedServiceInstanceCallCount()).To(Equal(1))
				params := fakeActor.CreateManagedServiceInstanceArgsForCall(0)
				Expect(params.ServicePlanName).To(Equal(requestedPlanName))
				Expect(params.ServiceOfferingName).To(Equal(requestedOfferingName))
				Expect(params.ServiceInstanceName).To(Equal(requestedServiceInstanceName))
				Expect(params.SpaceGUID).To(Equal(fakeSpaceGUID))
				Expect(params.Parameters).To(Equal(types.NewOptionalObject(requestedParams)))

			})
		})

		When("Creation is successful", func() {
			BeforeEach(func() {
				fakeActor.CreateManagedServiceInstanceReturns(
					v7action.Warnings{"be warned", "take care"},
					nil,
				)
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Err).To(SatisfyAll(
					Say("be warned"),
					Say("take care"),
				))
				Expect(testUI.Out).To(Say("Create in progress. Use 'cf services' or 'cf service %s' to check operation status.", requestedServiceInstanceName))

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
				fakeActor.CreateManagedServiceInstanceReturns(
					v7action.Warnings{"warning one", "warning two"},
					expectedError,
				)
			})

			It("returns the error", func() {
				Expect(executeErr).To(HaveOccurred())

				Expect(executeErr).To(MatchError(expectedError))
			})

			It("prints warnings", func() {
				Expect(testUI.Out).NotTo(Say("OK"))

				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})

			When("the service instance name is taken", func() {
				BeforeEach(func() {
					fakeActor.CreateManagedServiceInstanceReturns([]string{"a-warning", "another-warning"}, ccerror.ServiceInstanceNameTakenError{})
				})

				It("succeeds, displaying warnings, 'OK' and an informative message", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Err).To(Say("a-warning"))
					Expect(testUI.Err).To(Say("another-warning"))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Service %s already exists", requestedServiceInstanceName))
				})
			})
		})
	})
})
