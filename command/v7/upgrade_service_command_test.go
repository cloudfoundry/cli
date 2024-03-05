package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("upgrade-service command", func() {
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
		cmd             UpgradeServiceCommand
		executeErr      error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = UpgradeServiceCommand{
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
		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	testActorInteractions := func() {
		It("delegates to the actor", func() {
			Expect(fakeActor.UpgradeManagedServiceInstanceCallCount()).To(Equal(1))
			actualName, actualSpaceGUID := fakeActor.UpgradeManagedServiceInstanceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
		})

		When("the service instance does not exist", func() {
			BeforeEach(func() {
				fakeActor.UpgradeManagedServiceInstanceReturns(
					nil,
					v7action.Warnings{"upgrade warning"},
					actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("prints warnings and returns a translatable error", func() {
				Expect(testUI.Err).To(Say("upgrade warning"))
				Expect(executeErr).To(MatchError(translatableerror.ServiceInstanceNotFoundError{
					Name: serviceInstanceName,
				}))
			})
		})

		When("the actor returns an unexpected error", func() {
			BeforeEach(func() {
				fakeActor.UpgradeManagedServiceInstanceReturns(
					make(chan v7action.PollJobEvent),
					v7action.Warnings{"upgrade warning"},
					errors.New("bang"),
				)
			})

			It("fails with warnings", func() {
				Expect(executeErr).To(MatchError("bang"))
				Expect(testUI.Err).To(Say("upgrade warning"))
				Expect(testUI.Out).NotTo(Say("OK"))
			})
		})

		When("stream goes to polling", func() {
			BeforeEach(func() {
				fakeStream := make(chan v7action.PollJobEvent)
				fakeActor.UpgradeManagedServiceInstanceReturns(
					fakeStream,
					v7action.Warnings{"actor warning"},
					nil,
				)

				go func() {
					fakeStream <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"poll warning"},
					}
				}()
			})

			It("prints messages and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Upgrading service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`Upgrade in progress. Use 'cf services' or 'cf service %s' to check operation status\.`, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("actor warning"),
					Say("poll warning"),
				))
			})
		})

		When("error in event stream", func() {
			BeforeEach(func() {
				setFlag(&cmd, "--wait")

				fakeStream := make(chan v7action.PollJobEvent)
				fakeActor.UpgradeManagedServiceInstanceReturns(
					fakeStream,
					v7action.Warnings{"a warning"},
					nil,
				)

				go func() {
					fakeStream <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"poll warning"},
					}
					fakeStream <- v7action.PollJobEvent{
						State:    v7action.JobFailed,
						Warnings: v7action.Warnings{"failed warning"},
						Err:      errors.New("boom"),
					}
				}()
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("boom"))
				Expect(testUI.Err).To(SatisfyAll(
					Say("poll warning"),
					Say("failed warning"),
				))
			})
		})

		When("--wait flag specified", func() {
			BeforeEach(func() {
				setFlag(&cmd, "--wait")

				fakeStream := make(chan v7action.PollJobEvent)
				fakeActor.UpgradeManagedServiceInstanceReturns(
					fakeStream,
					v7action.Warnings{"a warning"},
					nil,
				)

				go func() {
					fakeStream <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"poll warning"},
					}
					fakeStream <- v7action.PollJobEvent{
						State:    v7action.JobComplete,
						Warnings: v7action.Warnings{"failed warning"},
					}
					close(fakeStream)
				}()
			})

			It("prints messages and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Upgrading service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`Waiting for the operation to complete\.\.\n`),
					Say(`\n`),
					Say(`Upgrade of service instance %s complete\.\n`, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("a warning"),
					Say("poll warning"),
					Say("failed warning"),
				))
			})
		})
	}

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		orgChecked, spaceChecked := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(orgChecked).To(BeTrue())
		Expect(spaceChecked).To(BeTrue())
	})

	It("prompts the user for confirmation", func() {
		Expect(testUI.Out).To(SatisfyAll(
			Say(`Warning: This operation may be long running and will block further operations on the service instance until it's completed`),
			Say(`Do you really want to upgrade the service instance %s\? \[yN\]:`, serviceInstanceName),
		))
	})

	When("the user confirms when prompted", func() {
		BeforeEach(func() {
			_, err := input.Write([]byte("y\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("outputs the attempted operation", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Upgrading service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
			))
		})

		testActorInteractions()
	})

	When("the user cancels when prompted", func() {
		BeforeEach(func() {
			_, err := input.Write([]byte("n\n"))
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not call the actor", func() {
			Expect(fakeActor.DeleteServiceInstanceCallCount()).To(BeZero())
		})

		It("outputs that the upgrade was cancelled", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("Upgrade cancelled\n"))
		})
	})

	When("the -f flags is passed", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-f")
		})

		It("does not prompt the user", func() {
			Expect(testUI.Out).NotTo(Say("Do you really want"))
		})

		testActorInteractions()
	})

	When("checking the target returns an error", func() {
		It("returns the error", func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
			executeErr := cmd.Execute(nil)
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
