package v7_test

import (
	"errors"
	"math/rand"

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

var _ = Describe("cleanup-outdated-service-bindings Command", func() {
	var (
		cmd                   v7.CleanupOutdatedServiceBindingsCommand
		testUI                *ui.UI
		input                 *Buffer
		confirmYes, confirmNo func() error
		executeErr            error
		fakeConfig            *commandfakes.FakeConfig
		fakeSharedActor       *commandfakes.FakeSharedActor
		fakeActor             *v7fakes.FakeActor
	)

	const (
		fakeUserName  = "fake-user-name"
		fakeOrgName   = "fake-org-name"
		fakeSpaceName = "fake-space-name"
		fakeSpaceGUID = "fake-space-guid"

		fakeAppName1 = "fake-app-name-1"
		fakeAppGUID1 = "fake-app-guid-1"

		fakeAppName2 = "fake-app-name-2"
		fakeAppGUID2 = "fake-app-guid-2"

		fakeServiceInstanceGUID1 = "fake-service-instance-guid-1"

		fakeServiceInstanceName2 = "fake-service-instance-name-2"
		fakeServiceInstanceGUID2 = "fake-service-instance-guid-2"

		fakeBindingGUID1 = "fake-binding-guid-1"
		fakeTimestamp1   = "2026-01-01T12:00:00Z"
		fakeBindingGUID2 = "fake-binding-guid-2"
		fakeTimestamp2   = "2026-01-02T12:00:00Z"
		fakeBindingGUID3 = "fake-binding-guid-3"
		fakeTimestamp3   = "2026-01-03T12:00:00Z"
		fakeBindingGUID4 = "fake-binding-guid-4"
		fakeTimestamp4   = "2026-01-04T12:00:00Z"
		fakeBindingGUID5 = "fake-binding-guid-5"
		fakeTimestamp5   = "2026-01-05T12:00:00Z"
		fakeBindingGUID6 = "fake-binding-guid-6"
		fakeTimestamp6   = "2026-01-06T12:00:00Z"
	)

	BeforeEach(func() {
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		confirmYes = func() error { _, err := input.Write([]byte("y\n")); return err }
		confirmNo = func() error { _, err := input.Write([]byte("N\n")); return err }

		cmd = v7.CleanupOutdatedServiceBindingsCommand{
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

		fakeActor.ListAppBindingsReturns(
			[]resources.ServiceCredentialBinding{
				// without any optional parameters, fakeBindingGUID1, 2 and 4 should be deleted (keeping last 1 per service instance)
				{GUID: fakeBindingGUID1, CreatedAt: fakeTimestamp1, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID1},
				{GUID: fakeBindingGUID2, CreatedAt: fakeTimestamp2, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
				{GUID: fakeBindingGUID3, CreatedAt: fakeTimestamp3, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID1},
				{GUID: fakeBindingGUID4, CreatedAt: fakeTimestamp4, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
				{GUID: fakeBindingGUID5, CreatedAt: fakeTimestamp5, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
				// binding for another app - should be ignored
				{GUID: fakeBindingGUID6, CreatedAt: fakeTimestamp6, AppName: fakeAppName2, AppGUID: fakeAppGUID2, ServiceInstanceGUID: fakeServiceInstanceGUID2},
			},
			v7action.Warnings{"list warning"},
			nil,
		)

		fakeActor.DeleteServiceAppBindingReturnsOnCall(0,
			nil,
			v7action.Warnings{"delete warning 1"},
			nil,
		)
		fakeActor.DeleteServiceAppBindingReturnsOnCall(1,
			nil,
			v7action.Warnings{"delete warning 2"},
			nil,
		)
		fakeActor.DeleteServiceAppBindingReturnsOnCall(2,
			nil,
			v7action.Warnings{"delete warning 3"},
			nil,
		)

		setPositionalFlags(&cmd, fakeAppName1)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("user and target setup", func() {
		BeforeEach(func() {
			err := confirmYes()
			Expect(err).ToNot(HaveOccurred())
		})

		It("checks the user is logged in, and targeting an org and space", func() {
			Expect(executeErr).NotTo(HaveOccurred())

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			actualOrg, actualSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(actualOrg).To(BeTrue())
			Expect(actualSpace).To(BeTrue())
		})
	})

	Describe("intro message", func() {
		BeforeEach(func() {
			err := confirmYes()
			Expect(err).ToNot(HaveOccurred())
		})

		It("prints an intro and warnings and a user confirmation message", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			// Warnings from list + delete
			Expect(testUI.Err).To(SatisfyAll(
				Say("list warning"),
			))

			Expect(testUI.Out).To(Say(`Cleaning up outdated service bindings for app %s in org %s / space %s as %s\.\.\.\n`, fakeAppName1, fakeOrgName, fakeSpaceName, fakeUserName))
			Expect(testUI.Out).To(Say(`Really delete all outdated service bindings\?`))
		})
	})

	Describe("GetOutdatedServiceBindings function", func() {
		var (
			bindings []resources.ServiceCredentialBinding
		)

		BeforeEach(func() {
			bindings = []resources.ServiceCredentialBinding{
				{GUID: "g1", AppGUID: "a1", ServiceInstanceGUID: "s1", CreatedAt: "2026-01-01T00:00:00Z"},
				{GUID: "g2", AppGUID: "a1", ServiceInstanceGUID: "s1", CreatedAt: "2026-01-02T00:00:00Z"},
				{GUID: "g3", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-01T00:00:00Z"},
				{GUID: "g4", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-03T00:00:00Z"},
				{GUID: "g5", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-02T00:00:00Z"},
			}
			r := rand.New(rand.NewSource(GinkgoRandomSeed()))
			r.Shuffle(len(bindings), func(i, j int) { bindings[i], bindings[j] = bindings[j], bindings[i] })
		})

		It("returns an empty list if an empty list of bindings is passed", func() {
			outdatedBindings := v7.GetOutdatedServiceBindings([]resources.ServiceCredentialBinding{}, 1)
			Expect(outdatedBindings).To(BeEmpty())
		})

		It("returns all but the newest per app/service pair when keepLast=1, sorted by 1. ServiceInstanceGUID and 2. CreatedAt", func() {
			outdatedBindings := v7.GetOutdatedServiceBindings(bindings, 1)

			Expect(outdatedBindings).To(HaveExactElements(
				resources.ServiceCredentialBinding{GUID: "g1", AppGUID: "a1", ServiceInstanceGUID: "s1", CreatedAt: "2026-01-01T00:00:00Z"},
				resources.ServiceCredentialBinding{GUID: "g3", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-01T00:00:00Z"},
				resources.ServiceCredentialBinding{GUID: "g5", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-02T00:00:00Z"},
			))
		})

		It("respects keepLast > 1", func() {
			outdatedBindings := v7.GetOutdatedServiceBindings(bindings, 2)

			Expect(outdatedBindings).To(HaveExactElements(
				resources.ServiceCredentialBinding{GUID: "g3", AppGUID: "a1", ServiceInstanceGUID: "s2", CreatedAt: "2026-01-01T00:00:00Z"},
			))
		})
	})

	Context("multiple bindings exist", func() {
		When("user confirms with yes", func() {
			BeforeEach(func() {
				err := confirmYes()
				Expect(err).ToNot(HaveOccurred())
			})

			When("no --service-instance parameter is provided", func() {
				It("lists bindings then delegates deletion of outdated bindings to the actor by GUID", func() {
					Expect(fakeActor.ListAppBindingsCallCount()).To(Equal(1))
					Expect(fakeActor.ListAppBindingsArgsForCall(0)).To(Equal(v7action.ListAppBindingParams{
						SpaceGUID: fakeSpaceGUID,
						AppName:   fakeAppName1,
					}))

					Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(3))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(0)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID1}))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(1)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID2}))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(2)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID4}))
				})

				It("prints deleting messages for all outdated bindings", func() {
					Expect(testUI.Out).To(Say(`Found 3 outdated service bindings\.`))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID1))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID2))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID4))
				})
			})

			When("--service-instance parameter is set", func() {
				BeforeEach(func() {
					fakeActor.ListServiceAppBindingsReturns(
						[]resources.ServiceCredentialBinding{
							// fakeBindingGUID2 and 4 should be deleted
							{GUID: fakeBindingGUID2, CreatedAt: fakeTimestamp2, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
							{GUID: fakeBindingGUID4, CreatedAt: fakeTimestamp4, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
							{GUID: fakeBindingGUID5, CreatedAt: fakeTimestamp5, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID2},
							// binding for another app - should be ignored
							{GUID: fakeBindingGUID6, CreatedAt: fakeTimestamp6, AppName: fakeAppName2, AppGUID: fakeAppGUID2, ServiceInstanceGUID: fakeServiceInstanceGUID2},
						},
						v7action.Warnings{"list warning"},
						nil,
					)

					setFlag(&cmd, "--service-instance", fakeServiceInstanceName2)
				})

				It("prints deleting messages for all outdated bindings", func() {
					Expect(testUI.Out).To(Say(`Found 2 outdated service bindings\.`))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID2))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID4))
				})

				It("lists bindings for the given service and then delegates deletion of outdated bindings to the actor by GUID", func() {
					Expect(fakeActor.ListAppBindingsCallCount()).To(Equal(0))
					Expect(fakeActor.ListServiceAppBindingsCallCount()).To(Equal(1))
					Expect(fakeActor.ListServiceAppBindingsArgsForCall(0)).To(Equal(v7action.ListServiceAppBindingParams{
						SpaceGUID:           fakeSpaceGUID,
						ServiceInstanceName: fakeServiceInstanceName2,
						AppName:             fakeAppName1,
					}))

					Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(2))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(0)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID2}))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(1)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID4}))
				})
			})

			When("--keep-last parameter is provided", func() {
				BeforeEach(func() {
					setFlag(&cmd, "--keep-last", func() *int { i := 2; return &i }())
				})

				It("lists bindings then delegates deletion of outdated bindings to the actor by GUID", func() {
					Expect(fakeActor.ListAppBindingsCallCount()).To(Equal(1))
					Expect(fakeActor.ListAppBindingsArgsForCall(0)).To(Equal(v7action.ListAppBindingParams{
						SpaceGUID: fakeSpaceGUID,
						AppName:   fakeAppName1,
					}))

					Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(1))
					Expect(fakeActor.DeleteServiceAppBindingArgsForCall(0)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID2}))
				})

				It("prints deleting messages for all outdated bindings", func() {
					Expect(testUI.Out).To(Say(`Found 1 outdated service binding\.`))
					Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID2))
				})
			})
		})

		When("user confirms with no", func() {
			BeforeEach(func() {
				err := confirmNo()
				Expect(err).ToNot(HaveOccurred())
			})

			It("aborts the operation", func() {
				Expect(executeErr).To(BeNil())
				Expect(testUI.Out).To(Say("Outdated service bindings have not been deleted."))
			})

			It("does not delete any bindings", func() {
				Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(0))
			})
		})

		When("the --force flag is set", func() {
			BeforeEach(func() {
				setFlag(&cmd, "--force")
			})

			It("deletes the outdated bindings without asking the user for confirmation", func() {
				Expect(executeErr).To(BeNil())
				Expect(testUI.Out).To(Say(`Found 3 outdated service bindings\.`))
				Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID1))
				Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID2))
				Expect(testUI.Out).To(Say(`Deleting service binding %s\.\.\.`, fakeBindingGUID4))
			})

			It("delegates deletion of outdated bindings to the actor by GUID", func() {
				Expect(fakeActor.DeleteServiceAppBindingCallCount()).To(Equal(3))
				Expect(fakeActor.DeleteServiceAppBindingArgsForCall(0)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID1}))
				Expect(fakeActor.DeleteServiceAppBindingArgsForCall(1)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID2}))
				Expect(fakeActor.DeleteServiceAppBindingArgsForCall(2)).To(Equal(v7action.DeleteServiceAppBindingParams{ServiceBindingGUID: fakeBindingGUID4}))
			})
		})
	})

	Describe("processing the response streams", func() {
		BeforeEach(func() {
			err := confirmYes()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("nil stream", func() {
			It("prints per-binding delete message, OK, and warnings", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Found 3 outdated service bindings\.\n`),
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID1),
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID2),
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID4),
					Say(`OK\n`),
				))
				Expect(testUI.Err).To(SatisfyAll(
					Say("list warning"),
					Say("delete warning 1"),
					Say("delete warning 2"),
					Say("delete warning 3"),
				))
			})
		})

		Context("one stream goes to complete, one to polling and the last to error", func() {
			BeforeEach(func() {
				eventStream1 := make(chan v7action.PollJobEvent)
				go func() {
					eventStream1 <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job 1 processing warning"},
					}
					eventStream1 <- v7action.PollJobEvent{
						State:    v7action.JobComplete,
						Warnings: v7action.Warnings{"job 1 complete warning"},
					}
					close(eventStream1)
				}()

				fakeActor.DeleteServiceAppBindingReturnsOnCall(0,
					eventStream1,
					v7action.Warnings{"delete 1 warning"},
					nil,
				)

				eventStream2 := make(chan v7action.PollJobEvent)
				go func() {
					eventStream2 <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job 2 processing warning"},
					}
					eventStream2 <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"job 2 polling warning"},
					}
				}()

				fakeActor.DeleteServiceAppBindingReturnsOnCall(1,
					eventStream2,
					v7action.Warnings{"delete 2 warning"},
					nil,
				)

				eventStream3 := make(chan v7action.PollJobEvent)
				go func() {
					eventStream3 <- v7action.PollJobEvent{
						State:    v7action.JobFailed,
						Warnings: v7action.Warnings{"job 3 failed warning"},
						Err:      errors.New("boom"),
					}
				}()

				fakeActor.DeleteServiceAppBindingReturnsOnCall(2,
					eventStream3,
					v7action.Warnings{"delete 3 warning"},
					nil,
				)

				fakeActor.GetServiceInstanceByGUIDReturns(
					resources.ServiceInstance{Name: fakeServiceInstanceName2},
					v7action.Warnings{"get service instance warning"},
					nil,
				)
			})

			It("processes all jobs and prints all messages", func() {
				Expect(testUI.Out).To(SatisfyAll(
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID1),
					Say(`OK\n`),
					Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID2),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Unbinding in progress. Use 'cf service %s' to check operation status\.\n`, fakeServiceInstanceName2),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("list warning"),
					Say("delete 1 warning"),
					Say("job 1 processing warning"),
					Say("job 1 complete warning"),
					Say("delete 2 warning"),
					Say("job 2 processing warning"),
					Say("job 2 polling warning"),
					Say("delete 3 warning"),
					Say("job 3 failed warning"),
				))

				Expect(executeErr).To(MatchError("boom"))
			})
		})

		Context("one stream goes to complete", func() {
			BeforeEach(func() {
				eventStream1 := make(chan v7action.PollJobEvent)
				go func() {
					eventStream1 <- v7action.PollJobEvent{
						State:    v7action.JobProcessing,
						Warnings: v7action.Warnings{"job 1 processing warning"},
					}
					eventStream1 <- v7action.PollJobEvent{
						State:    v7action.JobPolling,
						Warnings: v7action.Warnings{"job 1 polling warning"},
					}
					eventStream1 <- v7action.PollJobEvent{
						State:    v7action.JobComplete,
						Warnings: v7action.Warnings{"job 1 complete warning"},
					}
					close(eventStream1)
				}()

				fakeActor.ListAppBindingsReturns(
					[]resources.ServiceCredentialBinding{
						{GUID: fakeBindingGUID1, CreatedAt: fakeTimestamp1, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID1},
						{GUID: fakeBindingGUID3, CreatedAt: fakeTimestamp3, AppName: fakeAppName1, AppGUID: fakeAppGUID1, ServiceInstanceGUID: fakeServiceInstanceGUID1},
					},
					v7action.Warnings{"list warning"},
					nil,
				)

				fakeActor.DeleteServiceAppBindingReturnsOnCall(0,
					eventStream1,
					v7action.Warnings{"delete 1 warning"},
					nil,
				)
			})

			When("--wait flag specified", func() {
				BeforeEach(func() {
					setFlag(&cmd, "--wait")
				})

				It("waits for the event stream to complete", func() {
					Expect(testUI.Out).To(SatisfyAll(
						Say(`Deleting service binding %s\.\.\.\n`, fakeBindingGUID1),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`OK\n`),
					))

					Expect(testUI.Err).To(SatisfyAll(
						Say("list warning"),
						Say("delete 1 warning"),
						Say("job 1 processing warning"),
						Say("job 1 polling warning"),
						Say("job 1 complete warning"),
					))
				})
			})
		})
	})
})
