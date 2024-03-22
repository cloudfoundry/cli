package v7_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("delete-service-key Command", func() {
	var (
		cmd             v7.DeleteServiceKeyCommand
		input           *Buffer
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
	)

	const (
		fakeUserName            = "fake-user-name"
		fakeServiceInstanceName = "fake-service-instance-name"
		fakeServiceKeyName      = "fake-key-name"
		fakeSpaceGUID           = "fake-space-guid"
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.DeleteServiceKeyCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: fakeSpaceGUID})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

		fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
			nil,
			v7action.Warnings{"fake warning"},
			nil,
		)

		setPositionalFlags(&cmd, fakeServiceInstanceName, fakeServiceKeyName)
		setFlag(&cmd, "-f")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	It("checks the user is logged in, and targeting an org and space", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
		actualOrg, actualSpace := fakeSharedActor.CheckTargetArgsForCall(0)
		Expect(actualOrg).To(BeTrue())
		Expect(actualSpace).To(BeTrue())
	})

	It("delegates to the actor", func() {
		Expect(fakeActor.DeleteServiceKeyByServiceInstanceAndNameCallCount()).To(Equal(1))
		actualServiceInstanceName, actualKeyName, actualSpaceGUID := fakeActor.DeleteServiceKeyByServiceInstanceAndNameArgsForCall(0)
		Expect(actualKeyName).To(Equal(fakeServiceKeyName))
		Expect(actualServiceInstanceName).To(Equal(fakeServiceInstanceName))
		Expect(actualSpaceGUID).To(Equal(fakeSpaceGUID))
	})

	Describe("intro message", func() {
		It("prints an intro and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Err).To(Say("fake warning"))

			Expect(testUI.Out).To(Say(
				`Deleting key %s for service instance %s as %s\.\.\.\n`,
				fakeServiceKeyName,
				fakeServiceInstanceName,
				fakeUserName,
			))
		})
	})

	Describe("prompting the user", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-f", false)
		})

		It("prompts the user", func() {
			Expect(testUI.Out).To(Say(`Really delete the service key %s\? \[yN\]:`, fakeServiceKeyName))
		})

		When("user says no", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("n\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not call the actor", func() {
				Expect(fakeActor.DeleteServiceKeyByServiceInstanceAndNameCallCount()).To(BeZero())
			})

			It("says the unbind was cancelled", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Out).To(Say("Delete cancelled\n"))
			})
		})

		When("user says yes", func() {
			BeforeEach(func() {
				_, err := input.Write([]byte("y\n"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("calls the actor", func() {
				Expect(fakeActor.DeleteServiceKeyByServiceInstanceAndNameCallCount()).To(Equal(1))
			})
		})
	})

	When("service key not found", func() {
		BeforeEach(func() {
			fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
				nil,
				v7action.Warnings{"key warning"},
				actionerror.ServiceKeyNotFoundError{
					KeyName:             fakeServiceKeyName,
					ServiceInstanceName: fakeServiceInstanceName,
				},
			)
		})

		It("succeeds with a message", func() {
			Expect(testUI.Err).To(Say("key warning"))
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(SatisfyAll(
				Say(`\n`),
				Say(`Service key %s does not exist for service instance %s\.\n`, fakeServiceKeyName, fakeServiceInstanceName),
				Say(`OK\n`),
			))
		})
	})

	When("service instance not found", func() {
		BeforeEach(func() {
			fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
				nil,
				v7action.Warnings{"instance warning"},
				actionerror.ServiceInstanceNotFoundError{Name: fakeServiceInstanceName},
			)
		})

		It("returns an appropriate error", func() {
			Expect(testUI.Err).To(Say("instance warning"))
			Expect(executeErr).To(MatchError(translatableerror.ServiceInstanceNotFoundError{
				Name: fakeServiceInstanceName,
			}))
		})
	})

	When("there is an error", func() {
		BeforeEach(func() {
			fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
				nil,
				v7action.Warnings{"bad warning"},
				errors.New("boom"),
			)
		})

		It("returns the error", func() {
			Expect(testUI.Err).To(Say("bad warning"))
			Expect(executeErr).To(MatchError("boom"))
		})
	})

	Describe("processing the response stream", func() {
		Context("nil stream", func() {
			It("prints a message and warnings", func() {
				Expect(testUI.Out).To(Say(`OK\n`))
				Expect(testUI.Err).To(Say("fake warning"))
			})
		})

		Context("stream goes to complete", func() {
			BeforeEach(func() {
				eventStream := make(chan v7action.PollJobEvent)
				go func() {
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job processing warning"},
					}
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobComplete,
						Warnings: v7action.Warnings{"job complete warning"},
					}
					close(eventStream)
				}()

				fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
					eventStream,
					v7action.Warnings{"fake warning"},
					nil,
				)
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Out).To(Say(`OK\n`))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("job processing warning"),
					Say("job complete warning"),
				))
			})
		})

		Context("stream goes to polling", func() {
			BeforeEach(func() {
				eventStream := make(chan v7action.PollJobEvent)
				go func() {
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job processing warning"},
					}
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"job polling warning"},
					}
				}()

				fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
					eventStream,
					v7action.Warnings{"fake warning"},
					nil,
				)
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`OK\n`),
					Say(`\n`),
					Say(`Delete in progress\.\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("job processing warning"),
					Say("job polling warning"),
				))
			})
		})

		Context("stream goes to error", func() {
			BeforeEach(func() {
				eventStream := make(chan v7action.PollJobEvent)
				go func() {
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobFailed,
						Warnings: v7action.Warnings{"job failed warning"},
						Err:      errors.New("boom"),
					}
				}()

				fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
					eventStream,
					v7action.Warnings{"fake warning"},
					nil,
				)
			})

			It("prints warnings and returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("job failed warning"),
				))
			})
		})

		When("--wait flag specified", func() {
			BeforeEach(func() {
				eventStream := make(chan v7action.PollJobEvent)
				go func() {
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job processing warning"},
					}
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"job polling warning"},
					}
					eventStream <- v7action.PollJobEvent{
						State:    v7action.JobComplete,
						Warnings: v7action.Warnings{"job complete warning"},
					}
					close(eventStream)
				}()

				fakeActor.DeleteServiceKeyByServiceInstanceAndNameReturns(
					eventStream,
					v7action.Warnings{"fake warning"},
					nil,
				)

				setFlag(&cmd, "--wait")
			})

			It("waits for the event stream to complete", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Waiting for the operation to complete\.\.\.\n`),
					Say(`\n`),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("job processing warning"),
					Say("job polling warning"),
					Say("job complete warning"),
				))
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

	When("getting the username returns an error", func() {
		BeforeEach(func() {
			fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("bad thing"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("bad thing"))
		})
	})
})
