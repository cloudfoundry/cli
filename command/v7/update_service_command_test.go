package v7_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/flag"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo/v2"
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

		setPositionalFlags(&cmd, flag.TrimmedString(serviceInstanceName))

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: orgName})
		fakeConfig.TargetedSpaceReturns(configv3.Space{
			Name: spaceName,
			GUID: spaceGUID,
		})
		fakeActor.GetCurrentUserReturns(configv3.User{Name: username}, nil)

		fakeStream := make(chan v7action.PollJobEvent)
		close(fakeStream)
		fakeActor.UpdateManagedServiceInstanceReturns(
			fakeStream,
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

	When("upgrade flag specified", func() {
		BeforeEach(func() {
			setFlag(&cmd, "--upgrade")
		})

		It("prints a message and returns an error", func() {
			Expect(executeErr).To(MatchError(
				fmt.Sprintf(
					`Upgrading is no longer supported via updates, please run "cf upgrade-service %s" instead.`,
					serviceInstanceName,
				),
			))
		})
	})

	When("no parameters specified", func() {
		It("prints a message and exits 0", func() {
			Expect(executeErr).NotTo(HaveOccurred())
			Expect(testUI.Out).To(Say("No flags specified. No changes were made"))
		})
	})

	Describe("updates", func() {
		BeforeEach(func() {
			setFlag(&cmd, "-t", `foo,bar`)
			setFlag(&cmd, "-c", `{"baz": "quz"}`)
			setFlag(&cmd, "-p", "some-plan")
		})

		It("does not return an error", func() {
			Expect(executeErr).NotTo(HaveOccurred())
		})

		It("prints messages and warnings", func() {
			Expect(testUI.Out).To(SatisfyAll(
				Say(`Updating service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
				Say(`\n`),
				Say(`Update of service instance %s complete\.`, serviceInstanceName),
				Say(`OK\n`),
			))

			Expect(testUI.Err).To(Say("actor warning"))
		})

		It("delegates to the actor", func() {
			Expect(fakeActor.UpdateManagedServiceInstanceCallCount()).To(Equal(1))
			actualUpdates := fakeActor.UpdateManagedServiceInstanceArgsForCall(0)
			Expect(actualUpdates).To(Equal(v7action.UpdateManagedServiceInstanceParams{
				ServiceInstanceName: serviceInstanceName,
				SpaceGUID:           spaceGUID,
				Tags:                types.NewOptionalStringSlice("foo", "bar"),
				Parameters:          types.NewOptionalObject(map[string]interface{}{"baz": "quz"}),
				ServicePlanName:     "some-plan",
			}))
		})

		When("stream goes to polling", func() {
			BeforeEach(func() {
				fakeStream := make(chan v7action.PollJobEvent)
				fakeActor.UpdateManagedServiceInstanceReturns(
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
					Say(`Updating service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`Update in progress. Use 'cf services' or 'cf service %s' to check operation status\.`, serviceInstanceName),
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
				fakeActor.UpdateManagedServiceInstanceReturns(
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
				fakeActor.UpdateManagedServiceInstanceReturns(
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
					Say(`Updating service instance %s in org %s / space %s as %s...\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`Waiting for the operation to complete\.\.\n`),
					Say(`\n`),
					Say(`Update of service instance %s complete\.\n`, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(testUI.Err).To(SatisfyAll(
					Say("a warning"),
					Say("poll warning"),
					Say("failed warning"),
				))
			})
		})

		When("plan is current plan", func() {
			const currentPlan = "current-plan"

			BeforeEach(func() {
				setFlag(&cmd, "-p", currentPlan)
				fakeActor.UpdateManagedServiceInstanceReturns(
					nil,
					v7action.Warnings{"actor warning"},
					actionerror.ServiceInstanceUpdateIsNoop{},
				)
			})

			It("prints warnings and a message", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(testUI.Out).To(Say("No changes were made"))
			})
		})

		When("getting the user fails", func() {
			BeforeEach(func() {
				fakeActor.GetCurrentUserReturns(configv3.User{}, errors.New("bang"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("bang"))
			})
		})

		When("the actor reports the service instance was not found", func() {
			BeforeEach(func() {
				fakeActor.UpdateManagedServiceInstanceReturns(
					nil,
					v7action.Warnings{"actor warning"},
					actionerror.ServiceInstanceNotFoundError{Name: serviceInstanceName},
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError(actionerror.ServiceInstanceNotFoundError{
					Name: serviceInstanceName,
				}))
			})
		})

		When("plan not found", func() {
			const (
				invalidPlan = "invalid-plan"
			)

			BeforeEach(func() {
				setFlag(&cmd, "-p", invalidPlan)
				fakeActor.UpdateManagedServiceInstanceReturns(
					nil,
					v7action.Warnings{"actor warning"},
					actionerror.ServicePlanNotFoundError{PlanName: invalidPlan, ServiceBrokerName: "the-broker", OfferingName: "the-offering"},
				)
			})

			It("prints warnings and returns a translatable error", func() {
				Expect(testUI.Err).To(Say("actor warning"))
				Expect(executeErr).To(MatchError(actionerror.ServicePlanNotFoundError{
					PlanName:          invalidPlan,
					OfferingName:      "the-offering",
					ServiceBrokerName: "the-broker",
				}))
			})
		})

		When("the actor fails with an unexpected error", func() {
			BeforeEach(func() {
				fakeActor.UpdateManagedServiceInstanceReturns(
					nil,
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

	When("checking the target returns an error", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(errors.New("explode"))
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError("explode"))
		})
	})
})
