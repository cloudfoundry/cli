package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
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

		setPositionalFlags(&cmd, serviceInstanceName)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})
		fakeConfig.CurrentUserReturns(configv3.User{Name: username}, nil)

		fakeActor.UpdateManagedServiceInstanceReturns(
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

	When("plan is set", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-p", flag.OptionalString{IsSet: true, Value: "coolplan"})
		})

		It("fails", func() {
			Expect(executeErr).To(MatchError("not implemented"))
		})
	})

	Describe("updates", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-t", flag.Tags{IsSet: true, Value: []string{"foo", "bar"}})
			setFlag(&cmd, "-c", flag.JSONOrFileWithValidation{
				IsSet: true,
				Value: map[string]interface{}{"baz": "quz"},
			})
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
				Tags:       types.NewOptionalStringSlice("foo", "bar"),
				Parameters: types.NewOptionalObject(map[string]interface{}{"baz": "quz"}),
			}))
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
					v7action.Warnings{"actor warning"},
					actionerror.ServiceInstanceNotFoundError{},
				)
			})

			It("prints warnings and returns a translatable error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError(translatableerror.ServiceInstanceNotFoundError{
					Name: serviceInstanceName,
				}))
			})
		})

		When("the actor fails with an unexpected error", func() {
			BeforeEach(func() {
				fakeActor.UpdateManagedServiceInstanceReturns(
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

	When("upgrade is specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-u", true)
		})

		It("fails", func() {
			Expect(executeErr).To(MatchError("not implemented"))
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
