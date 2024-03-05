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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("purge-service-instance command", func() {
	const (
		serviceInstanceName = "service-instance-name"
		orgName             = "fake-org-name"
		spaceName           = "fake-space-name"
		spaceGUID           = "fake-space-guid"
		username            = "fake-username"
	)

	var (
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		cmd             PurgeServiceInstanceCommand
		executeErr      error
	)

	testActorInteractions := func() {
		It("delegates to the actor", func() {
			Expect(fakeActor.PurgeServiceInstanceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID := fakeActor.PurgeServiceInstanceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
		})

		When("the service instance did not exist", func() {
			BeforeEach(func() {
				fakeActor.PurgeServiceInstanceReturns(
					v7action.Warnings{"purge warning"},
					actionerror.ServiceInstanceNotFoundError{},
				)
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("purge warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say("\n"),
					Say(`Service instance %s did not exist\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})
		})

		When("the service instance is successfully purged", func() {
			BeforeEach(func() {
				fakeActor.PurgeServiceInstanceReturns(
					v7action.Warnings{"purge warning"},
					nil,
				)
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("purge warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say("\n"),
					Say(`Service instance %s purged\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.PurgeServiceInstanceReturns(
					v7action.Warnings{"purge warning"},
					errors.New("bang"),
				)
			})

			It("fails with warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(testUI.Err).To(Say("purge warning"))
				Expect(testUI.Out).NotTo(Say("OK"))
			})
		})
	}

	confirmYes := func() {
		_, err := input.Write([]byte("y\n"))
		Expect(err).NotTo(HaveOccurred())
	}

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: spaceName, GUID: spaceGUID})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)

		cmd = PurgeServiceInstanceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, serviceInstanceName)

		_ = executeErr
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

	It("prompts the user", func() {
		Expect(testUI.Out).To(SatisfyAll(
			Say(`WARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys.\n`),
			Say(`\n`),
			Say(`Really purge service instance %s from Cloud Foundry\? \[yN\]:`, serviceInstanceName),
		))
	})

	When("the user says yes", func() {
		BeforeEach(func() {
			confirmYes()
		})

		It("outputs the attempted operation", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Purging service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
			))
		})

		testActorInteractions()
	})

	When("the user says no", func() {
		BeforeEach(func() {
			_, err := input.Write([]byte("n\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not call the actor", func() {
			Expect(fakeActor.PurgeServiceInstanceCallCount()).To(BeZero())
		})

		It("says the delete was cancelled", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Purge cancelled\n"))
		})
	})

	When("the -f flag is specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-f")
		})

		It("does not prompt the user", func() {
			Expect(testUI.Out).NotTo(Say("Really purge"))
		})

		testActorInteractions()
	})

	Context("errors", func() {
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
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("boom"))
				confirmYes()
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})
	})
})
