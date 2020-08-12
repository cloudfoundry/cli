package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-service command", func() {
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
		cmd             DeleteServiceCommand
		executeErr      error
	)

	testActorInteractions := func() {
		It("delegates to the actor", func() {
			Expect(fakeActor.DeleteServiceInstanceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID, actualWait := fakeActor.DeleteServiceInstanceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
			Expect(actualWait).To(BeFalse())
		})

		When("the service instance did not exist", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					v7action.ServiceInstanceDidNotExist,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("delete warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say("\n"),
					Say(`Service instance %s did not exist\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})
		})

		When("the service instance is successfully deleted", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					v7action.ServiceInstanceGone,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("delete warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say("\n"),
					Say(`Service instance %s deleted\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})
		})

		When("the service instance deletion is in progress", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					v7action.ServiceInstanceDeleteInProgress,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("delete warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say("\n"),
					Say(`Delete in progress. Use 'cf services' or 'cf service %s' to check operation status\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					v7action.ServiceInstanceUnknownState,
					v7action.Warnings{"delete warning"},
					errors.New("bang"),
				)
			})

			It("fails with warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(testUI.Err).To(Say("delete warning"))
				Expect(testUI.Out).NotTo(Say("OK"))
			})
		})

		When("the -w flag is specified", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-w")
			})

			It("passes the wait flag to the actor", func() {
				Expect(fakeActor.DeleteServiceInstanceCallCount()).To(Equal(1))
				_, _, actualWait := fakeActor.DeleteServiceInstanceArgsForCall(0)
				Expect(actualWait).To(BeTrue())
			})
		})
	}

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{Name: spaceName, GUID: spaceGUID})
		fakeConfig.CurrentUserReturns(configv3.User{Name: username}, nil)

		cmd = DeleteServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, types.NewTrimmedString(serviceInstanceName))
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
			Say(`Deleting service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
			Say(`\n`),
			Say(`Really delete the service instance %s\? \[yN\]:`, serviceInstanceName),
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
			Expect(fakeActor.DeleteServiceInstanceCallCount()).To(BeZero())
		})

		It("says the delete was cancelled", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Delete cancelled\n"))
		})
	})

	When("the -f flag is specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-f")
		})

		It("does not prompt the user", func() {
			Expect(testUI.Out).NotTo(Say("really delete"))
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
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("boom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))
			})
		})
	})
})
