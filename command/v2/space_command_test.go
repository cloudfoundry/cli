package v2_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v2"
	"code.cloudfoundry.org/cli/command/v2/shared"
	"code.cloudfoundry.org/cli/command/v2/v2fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("space Command", func() {
	var (
		cmd             SpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v2fakes.FakeSpaceActor
		fakeActorV3     *v2fakes.FakeSpaceActorV3
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v2fakes.FakeSpaceActor)
		fakeActorV3 = new(v2fakes.FakeSpaceActorV3)

		cmd = SpaceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			ActorV3:     fakeActorV3,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		fakeConfig.ExperimentalReturns(true)
		fakeActorV3.CloudControllerAPIVersionReturns("3.12.0")
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	// TODO: remove when experimental flag is removed
	It("Displays the experimental warning message", func() {
		Expect(testUI.Out).To(Say(command.ExperimentalWarning))
	})

	Context("when checking the target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(
				sharedaction.NotLoggedInError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(
				command.NotLoggedInError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			config, targetedOrganizationRequired, targetedSpaceRequired := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(config).To(Equal(fakeConfig))
			Expect(targetedOrganizationRequired).To(Equal(true))
			Expect(targetedSpaceRequired).To(Equal(false))
		})
	})

	Context("when the --guid flag is provided", func() {
		BeforeEach(func() {
			cmd.RequiredArgs.Space = "some-space"
			cmd.GUID = true
		})

		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeConfig.TargetedOrganizationReturns(
					configv3.Organization{GUID: "some-org-guid"},
				)
				fakeActor.GetSpaceByOrganizationAndNameReturns(
					v2action.Space{GUID: "some-space-guid"},
					v2action.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("displays the space guid and outputs all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say("some-space-guid"))
				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))

				Expect(fakeActor.GetSpaceByOrganizationAndNameCallCount()).To(Equal(1))
				orgGUID, spaceName := fakeActor.GetSpaceByOrganizationAndNameArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(spaceName).To(Equal("some-space"))
			})
		})

		Context("when getting the space returns an error", func() {
			Context("when the error is translatable", func() {
				BeforeEach(func() {
					fakeActor.GetSpaceByOrganizationAndNameReturns(
						v2action.Space{},
						v2action.Warnings{"warning-1", "warning-2"},
						v2action.SpaceNotFoundError{Name: "some-space"})
				})

				It("returns a translatable error and outputs all warnings", func() {
					Expect(executeErr).To(MatchError(shared.SpaceNotFoundError{Name: "some-space"}))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})

			Context("when the error is not translatable", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get space error")
					fakeActor.GetSpaceByOrganizationAndNameReturns(
						v2action.Space{},
						v2action.Warnings{"warning-1", "warning-2"},
						expectedErr)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
				})
			})
		})
	})

	Context("when the --guid flag is not provided", func() {
		Context("when no errors occur", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{
						Name: "some-user",
					},
					nil)

				cmd.RequiredArgs.Space = "some-space"

				fakeConfig.TargetedOrganizationReturns(
					configv3.Organization{
						GUID: "some-org-guid",
						Name: "some-org",
					},
				)

				fakeActor.GetSpaceSummaryByOrganizationAndNameReturns(
					v2action.SpaceSummary{
						SpaceName:            "some-space",
						SpaceGUID:            "some-space-guid",
						OrgName:              "some-org",
						AppNames:             []string{"app1", "app2", "app3"},
						ServiceInstanceNames: []string{"service1", "service2", "service3"},
						SpaceQuotaName:       "some-space-quota",
						SecurityGroupNames:   []string{"public_networks", "dns", "load_balancer"},
					},
					v2action.Warnings{"warning-1", "warning-2"},
					nil,
				)
			})

			Context("when the v3 actor is nil", func() {
				BeforeEach(func() {
					cmd.ActorV3 = nil
				})
				It("displays the space summary with no isolation segment row", func() {
					Expect(executeErr).To(BeNil())
					Expect(testUI.Out).ToNot(Say("isolation segment:"))
				})
			})

			Context("when api version is above 3.11.0", func() {
				BeforeEach(func() {
					fakeActorV3.GetIsolationSegmentBySpaceReturns(
						v3action.IsolationSegment{
							Name: "some-isolation-segment",
						},
						v3action.Warnings{"v3-warning-1", "v3-warning-2"},
						nil,
					)
					fakeActorV3.CloudControllerAPIVersionReturns("3.12.0")
				})

				It("displays warnings and a table with space name, org, apps, services, isolation segment, space quota and security groups", func() {
					Expect(executeErr).To(BeNil())

					Eventually(testUI.Out).Should(Say("Getting info for space some-space in org some-org as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))
					Expect(testUI.Err).To(Say("v3-warning-1"))
					Expect(testUI.Err).To(Say("v3-warning-2"))

					Expect(testUI.Out).To(Say("name:\\s+some-space"))
					Expect(testUI.Out).To(Say("org:\\s+some-org"))
					Expect(testUI.Out).To(Say("apps:\\s+app1, app2, app3"))
					Expect(testUI.Out).To(Say("services:\\s+service1, service2, service3"))
					Expect(testUI.Out).To(Say("isolation segment:\\s+some-isolation-segment"))
					Expect(testUI.Out).To(Say("space quota:\\s+some-space-quota"))
					Expect(testUI.Out).To(Say("security groups:\\s+public_networks, dns, load_balancer"))

					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
					Expect(fakeActor.GetSpaceSummaryByOrganizationAndNameCallCount()).To(Equal(1))
					orgGUID, spaceName := fakeActor.GetSpaceSummaryByOrganizationAndNameArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))

					Expect(fakeActorV3.GetIsolationSegmentBySpaceCallCount()).To(Equal(1))
					Expect(fakeActorV3.GetIsolationSegmentBySpaceArgsForCall(0)).To(Equal("some-space-guid"))
				})
			})

			Context("when api version is below 3.11.0", func() {
				BeforeEach(func() {
					fakeActorV3.CloudControllerAPIVersionReturns("3.10.0")
				})

				It("displays warnings and a table with space name, org, apps, services, isolation segment, space quota and security groups", func() {
					Expect(executeErr).To(BeNil())

					Eventually(testUI.Out).Should(Say("Getting info for space some-space in org some-org as some-user\\.\\.\\."))
					Expect(testUI.Err).To(Say("warning-1"))
					Expect(testUI.Err).To(Say("warning-2"))

					Expect(testUI.Out).To(Say("name:\\s+some-space"))
					Expect(testUI.Out).To(Say("org:\\s+some-org"))
					Expect(testUI.Out).To(Say("apps:\\s+app1, app2, app3"))
					Expect(testUI.Out).To(Say("services:\\s+service1, service2, service3"))
					Expect(testUI.Out).ToNot(Say("isolation segment:"))
					Expect(testUI.Out).To(Say("space quota:\\s+some-space-quota"))
					Expect(testUI.Out).To(Say("security groups:\\s+public_networks, dns, load_balancer"))

					Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
					Expect(fakeActor.GetSpaceSummaryByOrganizationAndNameCallCount()).To(Equal(1))
					orgGUID, spaceName := fakeActor.GetSpaceSummaryByOrganizationAndNameArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))
					Expect(spaceName).To(Equal("some-space"))
				})
			})
		})
	})

	Context("when getting the current user returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("getting current user error")
			fakeConfig.CurrentUserReturns(
				configv3.User{},
				expectedErr)
		})

		It("returns the error", func() {
			Expect(executeErr).To(MatchError(expectedErr))
		})
	})

	Context("when getting the space summary returns an error", func() {
		Context("when the error is translatable", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByOrganizationAndNameReturns(
					v2action.SpaceSummary{},
					v2action.Warnings{"warning-1", "warning-2"},
					v2action.SpaceNotFoundError{Name: "some-space"})
			})

			It("returns a translatable error and outputs all warnings", func() {
				Expect(executeErr).To(MatchError(shared.SpaceNotFoundError{Name: "some-space"}))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})

		Context("when the error is not translatable", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get space summary error")
				fakeActor.GetSpaceSummaryByOrganizationAndNameReturns(
					v2action.SpaceSummary{},
					v2action.Warnings{"warning-1", "warning-2"},
					expectedErr)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))

				Expect(testUI.Err).To(Say("warning-1"))
				Expect(testUI.Err).To(Say("warning-2"))
			})
		})
	})

	Context("when getting the isolation segment returns an error", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("get isolation segment error")
			fakeActorV3.GetIsolationSegmentBySpaceReturns(
				v3action.IsolationSegment{},
				v3action.Warnings{"v3-warning-1", "v3-warning-2"},
				expectedErr)
		})

		It("returns the error and all warnings", func() {
			Expect(executeErr).To(MatchError(expectedErr))

			Expect(testUI.Err).To(Say("v3-warning-1"))
			Expect(testUI.Err).To(Say("v3-warning-2"))
		})
	})

	Context("when the --security-group-rules flag is provided", func() {
		BeforeEach(func() {
			fakeConfig.CurrentUserReturns(
				configv3.User{
					Name: "some-user",
				},
				nil)

			cmd.RequiredArgs.Space = "some-space"
			cmd.SecurityGroupRules = true

			fakeConfig.TargetedOrganizationReturns(
				configv3.Organization{
					GUID: "some-org-guid",
					Name: "some-org",
				},
			)

			fakeActor.GetSpaceSummaryByOrganizationAndNameReturns(
				v2action.SpaceSummary{
					SpaceName:            "some-space",
					OrgName:              "some-org",
					AppNames:             []string{"app1", "app2", "app3"},
					ServiceInstanceNames: []string{"service1", "service2", "service3"},
					SpaceQuotaName:       "some-space-quota",
					SecurityGroupNames:   []string{"public_networks", "dns", "load_balancer"},
					SecurityGroupRules: []v2action.SecurityGroupRule{
						{
							Description: "Public networks",
							Destination: "0.0.0.0-9.255.255.255",
							Lifecycle:   "staging",
							Name:        "public_networks",
							Ports:       "12345",
							Protocol:    "tcp",
						},
						{
							Description: "Public networks",
							Destination: "0.0.0.0-9.255.255.255",
							Lifecycle:   "running",
							Name:        "public_networks",
							Ports:       "12345",
							Protocol:    "tcp",
						},
						{
							Description: "More public networks",
							Destination: "11.0.0.0-169.253.255.255",
							Lifecycle:   "staging",
							Name:        "more_public_networks",
							Ports:       "54321",
							Protocol:    "udp",
						},
						{
							Description: "More public networks",
							Destination: "11.0.0.0-169.253.255.255",
							Lifecycle:   "running",
							Name:        "more_public_networks",
							Ports:       "54321",
							Protocol:    "udp",
						},
					},
				},
				v2action.Warnings{"warning-1", "warning-2"},
				nil,
			)
		})

		It("displays warnings and security group rules", func() {
			Expect(executeErr).To(BeNil())

			orgGUID, spaceName := fakeActor.GetSpaceSummaryByOrganizationAndNameArgsForCall(0)
			Expect(orgGUID).To(Equal("some-org-guid"))
			Expect(spaceName).To(Equal("some-space"))

			Eventually(testUI.Out).Should(Say("name:\\s+some-space"))
			Eventually(testUI.Out).Should(Say("(?m)^\n^\\s+security group\\s+destination\\s+ports\\s+protocol\\s+lifecycle\\s+description$"))
			Eventually(testUI.Out).Should(Say("#0\\s+public_networks\\s+0.0.0.0-9.255.255.255\\s+12345\\s+tcp\\s+staging\\s+Public networks"))
			Eventually(testUI.Out).Should(Say("(?m)^\\s+public_networks\\s+0.0.0.0-9.255.255.255\\s+12345\\s+tcp\\s+running\\s+Public networks"))
			Eventually(testUI.Out).Should(Say("#1\\s+more_public_networks\\s+11.0.0.0-169.253.255.255\\s+54321\\s+udp\\s+staging\\s+More public networks"))
			Eventually(testUI.Out).Should(Say("(?m)\\s+more_public_networks\\s+11.0.0.0-169.253.255.255\\s+54321\\s+udp\\s+running\\s+More public networks"))
		})
	})
})
