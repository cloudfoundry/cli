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
			actualName, actualSpaceGUID := fakeActor.DeleteServiceInstanceArgsForCall(0)
			Expect(actualName).To(Equal(serviceInstanceName))
			Expect(actualSpaceGUID).To(Equal(spaceGUID))
		})

		When("the service instance did not exist", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					nil,
					v7action.Warnings{"delete warning"},
					actionerror.ServiceInstanceNotFoundError{},
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
					nil,
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
				fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
					stream := make(chan v7action.PollJobEvent)

					go func() {
						stream <- v7action.PollJobEvent{
							State: v7action.JobPolling,
						}
						// channel not closed
					}()

					return stream, v7action.Warnings{"delete warning"}, nil
				})
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

			When("there is a warning in the event stream", func() {
				BeforeEach(func() {
					fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
						stream := make(chan v7action.PollJobEvent)

						go func() {
							stream <- v7action.PollJobEvent{
								Warnings: v7action.Warnings{"stream warning"},
							}
							close(stream)
						}()

						return stream, v7action.Warnings{"delete warning"}, nil
					})
				})

				It("prints it", func() {
					Expect(testUI.Err).To(Say("stream warning"))
				})
			})

			When("there is an error in the event stream", func() {
				BeforeEach(func() {
					fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
						stream := make(chan v7action.PollJobEvent)

						go func() {
							stream <- v7action.PollJobEvent{
								Err: errors.New("stream error"),
							}
							close(stream)
						}()

						return stream, v7action.Warnings{"delete warning"}, nil
					})
				})

				It("returns it", func() {
					Expect(executeErr).To(MatchError("stream error"))
				})
			})
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.DeleteServiceInstanceReturns(
					nil,
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

				fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
					stream := make(chan v7action.PollJobEvent)

					go func() {
						stream <- v7action.PollJobEvent{State: v7action.JobPolling}
						stream <- v7action.PollJobEvent{State: v7action.JobComplete}
						close(stream)
					}()

					return stream, v7action.Warnings{"delete warning"}, nil
				})
			})

			It("succeeds with a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("delete warning"))
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Waiting for the operation to complete\.\.\n`),
					Say("\n"),
					Say(`Service instance %s deleted\.\n`, serviceInstanceName),
					Say("OK\n"),
				))
			})

			When("there is a warning in the event stream", func() {
				BeforeEach(func() {
					fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
						stream := make(chan v7action.PollJobEvent)

						go func() {
							stream <- v7action.PollJobEvent{
								Warnings: v7action.Warnings{"stream warning"},
							}
							close(stream)
						}()

						return stream, v7action.Warnings{"delete warning"}, nil
					})
				})

				It("prints it", func() {
					Expect(testUI.Err).To(Say("stream warning"))
				})
			})

			When("there is an error in the event stream", func() {
				BeforeEach(func() {
					fakeActor.DeleteServiceInstanceCalls(func(name, spaceGUID string) (chan v7action.PollJobEvent, v7action.Warnings, error) {
						stream := make(chan v7action.PollJobEvent)

						go func() {
							stream <- v7action.PollJobEvent{
								Err: errors.New("stream error"),
							}
							close(stream)
						}()

						return stream, v7action.Warnings{"delete warning"}, nil
					})
				})

				It("returns it", func() {
					Expect(executeErr).To(MatchError("stream error"))
				})
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

		cmd = DeleteServiceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		setPositionalFlags(&cmd, serviceInstanceName)
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
			Say(`This action impacts all resources scoped to this service instance, including service bindings, service keys and route bindings\.`),
			Say(`This will remove the service instance from any spaces where it has been shared\.`),
			Say(`Really delete the service instance %s\? \[yN\]:`, serviceInstanceName),
		))
	})

	When("the user says yes", func() {
		BeforeEach(func() {
			confirmYes()
		})

		It("informs the user what it is doing", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Deleting service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
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
			Expect(testUI.Out).NotTo(Say("Really delete"))
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
