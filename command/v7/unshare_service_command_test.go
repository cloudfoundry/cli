package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"

	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unshare-service command", func() {
	var (
		cmd             UnshareServiceCommand
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		executeErr      error
	)

	const (
		expectedServiceInstanceName = "fake-service-instance-name"
		expectedSpaceName           = "fake-space-name"
		expectedTargetedSpaceGuid   = "fake-space-guid"
		expectedTargetedOrgName     = "fake-org-name"
		expectedTargetedOrgGuid     = "fake-org-guid"
		expectedUser                = "fake-username"
	)

	testActorInteractions := func() {
		It("calls the actor to share in specified space and targeted org", func() {
			Expect(fakeActor.UnshareServiceInstanceFromSpaceAndOrgCallCount()).To(Equal(1))

			actualServiceInstance, actualTargetedSpace, actualTargetedOrg, actualSharingParams := fakeActor.UnshareServiceInstanceFromSpaceAndOrgArgsForCall(0)
			Expect(actualServiceInstance).To(Equal(expectedServiceInstanceName))
			Expect(actualTargetedSpace).To(Equal(expectedTargetedSpaceGuid))
			Expect(actualTargetedOrg).To(Equal(expectedTargetedOrgGuid))
			Expect(actualSharingParams).To(Equal(v7action.ServiceInstanceSharingParams{
				SpaceName: expectedSpaceName,
				OrgName:   types.OptionalString{},
			}))
		})

		When("the unshare completes successfully", func() {
			BeforeEach(func() {
				fakeActor.UnshareServiceInstanceFromSpaceAndOrgReturns(v7action.Warnings{"warning one", "warning two"}, nil)
			})

			It("returns an OK message", func() {
				Expect(executeErr).To(BeNil())

				Expect(testUI.Out).To(
					Say(`Unsharing service instance %s from org %s / space %s as %s`,
						expectedServiceInstanceName,
						expectedTargetedOrgName,
						expectedSpaceName,
						expectedUser))
				Expect(testUI.Out).To(Say(`OK`))
				Expect(testUI.Err).To(SatisfyAll(
					Say("warning one"),
					Say("warning two"),
				))
			})
		})

		When("organization flag is specified", func() {
			var expectedSpecifiedOrgName = "fake-org-name"

			BeforeEach(func() {
				setFlag(&cmd, "-o", types.NewOptionalString(expectedSpecifiedOrgName))
			})

			It("calls the actor to share in specified space and org", func() {
				Expect(fakeActor.UnshareServiceInstanceFromSpaceAndOrgCallCount()).To(Equal(1))

				_, _, _, actualSharingParams := fakeActor.UnshareServiceInstanceFromSpaceAndOrgArgsForCall(0)
				Expect(actualSharingParams).To(Equal(v7action.ServiceInstanceSharingParams{
					SpaceName: expectedSpaceName,
					OrgName:   types.NewOptionalString(expectedSpecifiedOrgName),
				}))
			})
		})

		When("the actor errors", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(nil)
				fakeActor.UnshareServiceInstanceFromSpaceAndOrgReturns(v7action.Warnings{"actor warning"}, errors.New("test error"))
			})

			It("prints all warnings and fails with an error", func() {
				Expect(executeErr).To(Not(BeNil()))
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError("test error"))

			})
		})
	}

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: expectedTargetedOrgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: expectedSpaceName, GUID: expectedTargetedSpaceGuid})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: expectedUser}, nil)

		cmd = UnshareServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}
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

	Context("user not targeting space", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("space not targeted"))
		})

		It("fails the command", func() {
			Expect(executeErr).To(Not(BeNil()))
			Expect(executeErr).To(MatchError("space not targeted"))
		})
	})

	Context("user is targeting a space and org", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.ServiceInstance = expectedServiceInstanceName
			setFlag(&cmd, "-s", expectedSpaceName)

			fakeSharedActor.CheckTargetReturns(nil)
			fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: expectedTargetedSpaceGuid})
			fakeConfig.TargetedOrganizationReturns(configv3.Organization{GUID: expectedTargetedOrgGuid, Name: expectedTargetedOrgName})
			fakeActor.GetCurrentUserReturns(configv3.User{Name: expectedUser}, nil)
		})

		It("prompts the user", func() {
			Expect(testUI.Err).To(Say(`WARNING: Unsharing this service instance will remove any existing bindings originating from the service instance in the space "%s". This could cause apps to stop working.`, expectedSpaceName))
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Really unshare the service instance %s from space %s\? \[yN\]:`,
					expectedServiceInstanceName,
					expectedSpaceName),
			))
		})

		When("the user says yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			testActorInteractions()
		})

		When("the user says no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not call the actor", func() {
				Expect(fakeActor.UnshareServiceInstanceFromSpaceAndOrgCallCount()).To(BeZero())
			})

			It("says the delete was cancelled", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("Unshare cancelled\n"))
			})
		})

		When("the -f flag is specified", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-f")
			})

			It("does not prompt the user", func() {
				Expect(testUI.Out).NotTo(Say("Really unshare"))
			})

			testActorInteractions()
		})
	})

	Context("pre-unshare errors", func() {
		When("checking the target returns an error", func() {
			BeforeEach(func() {
				fakeSharedActor.CheckTargetReturns(errors.New("explode"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("explode"))
			})
		})

		When("getting the username fails", func() {
			BeforeEach(func() {
				input.Write([]byte("y\n"))
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("boom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})
	})
})
