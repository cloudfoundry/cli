package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("create-service-key Command", func() {
	var (
		cmd             v7.CreateServiceKeyCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
	)

	const (
		fakeUserName            = "fake-user-name"
		fakeServiceInstanceName = "fake-service-instance-name"
		fakeServiceKeyName      = "fake-service-key-name"
		fakeSpaceGUID           = "fake-space-guid"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.CreateServiceKeyCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{GUID: fakeSpaceGUID})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

		fakeActor.CreateServiceKeyReturns(
			nil,
			v7action.Warnings{"fake warning"},
			nil,
		)

		setPositionalFlags(&cmd, fakeServiceInstanceName, fakeServiceKeyName)
		setFlag(&cmd, "-c", `{"foo": "bar"}`)
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
		Expect(fakeActor.CreateServiceKeyCallCount()).To(Equal(1))
		Expect(fakeActor.CreateServiceKeyArgsForCall(0)).To(Equal(v7action.CreateServiceKeyParams{
			SpaceGUID:           fakeSpaceGUID,
			ServiceInstanceName: fakeServiceInstanceName,
			ServiceKeyName:      fakeServiceKeyName,
			Parameters:          types.NewOptionalObject(map[string]interface{}{"foo": "bar"}),
		}))
	})

	It("displays messages and warnings", func() {
		Expect(executeErr).NotTo(HaveOccurred())

		Expect(testUI.Out).To(SatisfyAll(
			Say(`Creating service key %s for service instance %s as %s\.\.\.\n`, fakeServiceKeyName, fakeServiceInstanceName, fakeUserName),
			Say(`OK\n`),
		))

		Expect(testUI.Err).To(Say("fake warning"))
	})

	When("key already exists", func() {
		BeforeEach(func() {
			fakeActor.CreateServiceKeyReturns(
				nil,
				v7action.Warnings{"fake warning"},
				actionerror.ResourceAlreadyExistsError{},
			)
		})

		It("prints a message and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`\n`),
				Say(`Service key %s already exists\n`, fakeServiceKeyName),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("fake warning"))
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

				fakeActor.CreateServiceKeyReturns(
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

				fakeActor.CreateServiceKeyReturns(
					eventStream,
					v7action.Warnings{"fake warning"},
					nil,
				)
			})

			It("prints a message and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`OK\n`),
					Say(`\n`),
					Say(`Create in progress\.\n`),
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

				fakeActor.CreateServiceKeyReturns(
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

				fakeActor.CreateServiceKeyReturns(
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

	When("actor returns error", func() {
		BeforeEach(func() {
			fakeActor.CreateServiceKeyReturns(
				nil,
				v7action.Warnings{"fake warning"},
				errors.New("boom"),
			)
		})

		It("prints warnings and returns the error", func() {
			Expect(testUI.Err).To(Say("fake warning"))
			Expect(executeErr).To(MatchError("boom"))
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
