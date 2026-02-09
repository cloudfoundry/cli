package v7_test

import (
	"errors"
	"code.cloudfoundry.org/cli/v8/actor/actionerror"
	"code.cloudfoundry.org/cli/v8/actor/v7action"
	"code.cloudfoundry.org/cli/v8/command/commandfakes"
	v7 "code.cloudfoundry.org/cli/v8/command/v7"
	"code.cloudfoundry.org/cli/v8/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"code.cloudfoundry.org/cli/v8/util/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("unbind-service Command", func() {
	var (
		cmd             v7.UnbindServiceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		executeErr      error
		fakeActor       *v7fakes.FakeActor
	)

	const (
		fakeUserName            = "fake-user-name"
		fakeServiceInstanceName = "fake-service-instance-name"
		fakeAppName             = "fake-app-name"
		fakeOrgName             = "fake-org-name"
		fakeSpaceName           = "fake-space-name"
		fakeSpaceGUID           = "fake-space-guid"
		fakeBindingGUID         = "fake-binding-guid"
		fakeBindingGUID2        = "fake-binding-guid-2"
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(NewBuffer(), NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = v7.UnbindServiceCommand{
			BaseCommand: v7.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: fakeSpaceName,
			GUID: fakeSpaceGUID,
		})

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: fakeOrgName})

		fakeActor.GetCurrentUserReturns(configv3.User{Name: fakeUserName}, nil)

		fakeActor.ListServiceAppBindingsReturns(
			[]resources.ServiceCredentialBinding{{GUID: fakeBindingGUID}},
			v7action.Warnings{"fake warning"},
			nil,
		)
		fakeActor.DeleteServiceAppBindingReturns(
			nil,
			v7action.Warnings{"delete warning"},
			nil,
		)

		setPositionalFlags(&cmd, fakeAppName, fakeServiceInstanceName)
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

	Context("one binding exists", func() {
		It("lists binding then delegates deletion to the actor by GUID", func() {
			Expect(fakeActor.ListServiceAppBindingsCallCount()).To(Equal(1))
			Expect(fakeActor.ListServiceAppBindingsArgsForCall(0)).To(Equal(v7action.ListServiceAppBindingParams{
				SpaceGUID:           fakeSpaceGUID,
				ServiceInstanceName: fakeServiceInstanceName,
				AppName:             fakeAppName,
			}))

			Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(1))
			Expect(fakeActor.DeleteServiceAppBindingArgsForCall(0)).To(Equal(v7action.DeleteServiceAppBindingParams{
				ServiceBindingGUID: fakeBindingGUID,
			}))
		})
	})

	Describe("intro message", func() {
		It("prints an intro and warnings", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			// Warnings from list + delete
			Expect(testUI.Err).To(SatisfyAll(
				Say("fake warning"),
				Say("delete warning"),
			))

			Expect(testUI.Out).To(Say(
				`Unbinding app %s from service %s in org %s / space %s as %s\.\.\.\n`,
				fakeAppName,
				fakeServiceInstanceName,
				fakeOrgName,
				fakeSpaceName,
				fakeUserName,
			))
		})
	})

	When("binding does not exist", func() {
		BeforeEach(func() {
			fakeActor.ListServiceAppBindingsReturns(
				nil,
				v7action.Warnings{"fake warning"},
				actionerror.ServiceBindingNotFoundError{},
			)
		})

		It("prints a message and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(
					`Binding between %s and %s does not exist\n`,
					fakeServiceInstanceName,
					fakeAppName,
				),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("fake warning"))
		})
	})

	Describe("processing the response stream", func() {
		Context("nil stream", func() {
			It("prints per-binding delete message, OK, and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID),
					Say(`OK\n`),
				))
				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("delete warning"),
				))
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

				fakeActor.DeleteServiceAppBindingReturns(
					eventStream,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("prints delete message, OK, and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("delete warning"),
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

				fakeActor.DeleteServiceAppBindingReturns(
					eventStream,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("prints delete message, OK, in-progress note, and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Unbinding in progress. Use 'cf service %s' to check operation status\.\n`, fakeServiceInstanceName),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("delete warning"),
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

				fakeActor.DeleteServiceAppBindingReturns(
					eventStream,
					v7action.Warnings{"delete warning"},
					nil,
				)
			})

			It("prints warnings and returns the error", func() {
				Expect(executeErr).To(MatchError("boom"))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("delete warning"),
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

				fakeActor.DeleteServiceAppBindingReturns(
					eventStream,
					v7action.Warnings{"delete warning"},
					nil,
				)

				setFlag(&cmd, "--wait")
			})

			It("waits for the event stream to complete", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID),
					Say(`Waiting for the operation to complete\.+\n`),
					Say(`\n`),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("fake warning"),
					Say("delete warning"),
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

	When("list returns error", func() {
		BeforeEach(func() {
			fakeActor.ListServiceAppBindingsReturns(
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

	When("delete returns error", func() {
		BeforeEach(func() {
			fakeActor.DeleteServiceAppBindingReturns(
				nil,
				v7action.Warnings{"delete warning"},
				errors.New("boom"),
			)
		})

		It("prints warnings and returns the error", func() {
			Expect(testUI.Err).To(SatisfyAll(
				Say("fake warning"),
				Say("delete warning"),
			))
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

	Context("multiple bindings exist", func() {
		BeforeEach(func() {
			fakeActor.ListServiceAppBindingsReturns(
				[]resources.ServiceCredentialBinding{{GUID: fakeBindingGUID}, {GUID: fakeBindingGUID2}},
				v7action.Warnings{"fake warning"},
				nil,
			)
		})

		It("deletes each binding by GUID and prints messages for both", func() {
			Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(2))

			firstArgs := fakeActor.DeleteServiceAppBindingArgsForCall(0)
			secondArgs := fakeActor.DeleteServiceAppBindingArgsForCall(1)
			Expect([]string{firstArgs.ServiceBindingGUID, secondArgs.ServiceBindingGUID}).To(ConsistOf(fakeBindingGUID, fakeBindingGUID2))

			Expect(testUI.Out).To(SatisfyAll(
				Say(`Deleting service binding %s\.+\n`, fakeBindingGUID),
				Say(`OK\n`),
				Say(`Deleting service binding %s\.+\n`, fakeBindingGUID2),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("fake warning"))
		})
	})

	Context("--guid selects a single binding among multiple", func() {
		BeforeEach(func() {
			fakeActor.ListServiceAppBindingsReturns(
				[]resources.ServiceCredentialBinding{{GUID: fakeBindingGUID}, {GUID: fakeBindingGUID2}},
				v7action.Warnings{"fake warning"},
				nil,
			)
			setFlag(&cmd, "--guid", fakeBindingGUID2)
		})

		It("deletes only the specified GUID", func() {
			Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(1))
			args := fakeActor.DeleteServiceAppBindingArgsForCall(0)
			Expect(args.ServiceBindingGUID).To(Equal(fakeBindingGUID2))

			Expect(testUI.Out).To(SatisfyAll(
				Say(`Deleting service binding %s\.+\n`, fakeBindingGUID2),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(SatisfyAll(
				Say("fake warning"),
				Say("delete warning"),
			))
		})
	})

	Context("--guid does not match any binding", func() {
		BeforeEach(func() {
			fakeActor.ListServiceAppBindingsReturns(
				[]resources.ServiceCredentialBinding{{GUID: fakeBindingGUID}},
				v7action.Warnings{"fake warning"},
				nil,
			)
			setFlag(&cmd, "--guid", "unknown-guid")
		})

		It("prints a specific not-found message and OK, without calling delete", func() {
			Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(0))

			Expect(testUI.Out).To(SatisfyAll(
				Say(`Service binding with GUID unknown-guid does not exist\n`),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("fake warning"))
		})
	})
})
