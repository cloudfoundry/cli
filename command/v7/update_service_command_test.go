package v7_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("update-service command", func() {
	const (
		serviceInstanceName = "fake-service-instance-name"
		spaceName           = "fake-space-name"
		spaceGUID           = "fake-space-guid"
		orgName             = "fake-org-name"
		username            = "fake-username"
	)

	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		cmd             UpdateServiceCommand
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UpdateServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, flag.TrimmedString(serviceInstanceName))

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: username}, nil)

		fakeActor.UpdateManagedServiceInstanceReturns(
			false,
			v7action.Warnings{"actor warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	When("upgrade flag specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "--upgrade")
		})

		It("prints a message and returns an error", func() {
			Expect(executeErr).To(MatchError(
				fmt.Sprintf(
					`Upgrading is no longer supported via updates, please run "cf upgrade-service %s" instead.`,
					serviceInstanceName,
				),
			))
		})
	})

	When("no parameters specified", func() {
		It("prints a message and exits 0", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("No flags specified. No changes were made"))
		})
	})

	Describe("updates", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-t", flag.Tags{IsSet: true, Value: []string{"foo", "bar"}})
			setFlag(&cmd, "-c", flag.JSONOrFileWithValidation{
				IsSet: true,
				Value: map[string]interface{}{"baz": "quz"},
			})
			setFlag(&cmd, "-p", flag.OptionalString{IsSet: true, Value: "some-plan"})
		})

		It("does not return an error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
		})

		It("prints messages and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Updating service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("actor warning"))
		})

		It("delegates to the actor", func() {
			Expect(fakeActor.UpdateManagedServiceInstanceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID, actualUpdates := fakeActor.UpdateManagedServiceInstanceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualUpdates).To(Equal(v7action.ServiceInstanceUpdateManagedParams{
				Tags:            types.NewOptionalStringSlice("foo", "bar"),
				Parameters:      types.NewOptionalObject(map[string]interface{}{"baz": "quz"}),
				ServicePlanName: types.NewOptionalString("some-plan"),
			}))
		})

		When("plan is current plan", func() {
			const (
				currentPlan = "current-plan"
			)

			BeforeEach(func() {
				setFlag(&cmd, "-p", flag.OptionalString{IsSet: true, Value: currentPlan})
				fakeActor.UpdateManagedServiceInstanceReturns(
					true,
					v7action.Warnings{"actor warning"},
					nil,
				)
			})

			It("prints warnings and a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(testUI.Out).To(Say("No changes were made"))
			})
		})

		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("bang"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("bang"))
			})
		})

		When("the actor reports the service instance was not found", func() {
			BeforeEach(func() {
				fakeActor.UpdateManagedServiceInstanceReturns(
					false,
					v7action.Warnings{"actor warning"},
					actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
					Name: serviceInstanceName,
				}))
			})
		})

		When("plan not found", func() {
			const (
				invalidPlan = "invalid-plan"
			)

			BeforeEach(func() {
				setFlag(&cmd, "-p", flag.OptionalString{IsSet: true, Value: invalidPlan})
				fakeActor.UpdateManagedServiceInstanceReturns(
					false,
					v7action.Warnings{"actor warning"},
					actionerror.ServicePlanNotFoundError{PlanName: invalidPlan, ServiceBrokerName: "the-broker", OfferingName: "the-offering"},
				)
			})

			It("prints warnings and returns a translatable error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError(actionerror.ServicePlanNotFoundError{
					PlanName:          invalidPlan,
					OfferingName:      "the-offering",
					ServiceBrokerName: "the-broker",
				}))
			})
		})

		When("the actor fails with an unexpected error", func() {
			BeforeEach(func() {
				fakeActor.UpdateManagedServiceInstanceReturns(
					false,
					v7action.Warnings{"actor warning"},
					errors.New("boof"),
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError("boof"))
			})
		})
	})

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
