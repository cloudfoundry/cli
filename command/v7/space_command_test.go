package v7_test

import (
	"errors"

	"code.cloudfoundry.org/cli/integration/helpers"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/command/flag"

	"code.cloudfoundry.org/cli/actor/v7action"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("space command", func() {
	var (
		cmd             SpaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
		binaryName      string
		executeErr      error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = SpaceCommand{
			BaseCommand: command.BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
			RequiredArgs: flag.Space{
				Space: "some-space",
			},
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			Name: "some-org",
			GUID: "some-org-guid",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "steve"}, nil)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("checking target fails", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NoOrganizationTargetedError{BinaryName: binaryName})
		})

		It("returns an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NoOrganizationTargetedError{BinaryName: binaryName}))

			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
			checkTargetedOrg, checkTargetedSpace := fakeSharedActor.CheckTargetArgsForCall(0)
			Expect(checkTargetedOrg).To(BeTrue())
			Expect(checkTargetedSpace).To(BeFalse())
		})
	})

	When("the user is not logged in", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("some current user error")
			fakeConfig.CurrentUserReturns(configv3.User{}, expectedErr)
		})

		It("return an error", func() {
			Expect(executeErr).To(Equal(expectedErr))
		})
	})

	When("the --guid flag is passed", func() {
		BeforeEach(func() {
			cmd.GUID = true
		})

		When("getting the space succeeds", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					v7action.Space{GUID: "some-space-guid"},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays warnings and the space guid", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(testUI.Out).To(Say("some-space-guid"))
			})
		})

		When("getting the space fails", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceByNameAndOrganizationReturns(
					v7action.Space{},
					v7action.Warnings{"some-warning"},
					errors.New("space-error"),
				)
			})

			It("displays warnings and returns the error", func() {
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("space-error"))
			})
		})
	})

	When("the --guid flag is not passed", func() {
		var (
			runningSecurityGroup1 resources.SecurityGroup
			runningSecurityGroup2 resources.SecurityGroup
			stagingSecurityGroup  resources.SecurityGroup
			description           string
			ports                 string
		)

		BeforeEach(func() {
			ports = "8080"
			description = "Test security group"
			runningSecurityGroup1 = helpers.NewSecurityGroup(
				"dns",
				"tcp",
				"10.244.1.18",
				&ports,
				&description,
			)
			runningSecurityGroup2 = helpers.NewSecurityGroup(
				"credhub",
				"all",
				"0.0.0.0-5.6.7.8",
				nil,
				&description,
			)
			stagingSecurityGroup = helpers.NewSecurityGroup(
				"sarah",
				"udp",
				"127.0.0.1",
				&ports,
				nil,
			)
		})

		When("the --security-group-rules flag is passed", func() {
			BeforeEach(func() {
				cmd.SecurityGroupRules = true

				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{
						Name:                  "some-space",
						OrgName:               "some-org",
						AppNames:              []string{"app1", "app2", "app3"},
						ServiceInstanceNames:  []string{"instance1", "instance2"},
						RunningSecurityGroups: []resources.SecurityGroup{runningSecurityGroup1, runningSecurityGroup2},
						StagingSecurityGroups: []resources.SecurityGroup{stagingSecurityGroup},
					},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays flavor text", func() {
				Expect(testUI.Out).To(Say("Getting info for space some-space in org some-org as steve..."))
			})

			It("displays warnings", func() {
				Expect(testUI.Err).To(Say("some-warning"))
			})

			It("displays a table of values and security group rules", func() {
				Expect(testUI.Out).To(Say(`name:\s+some-space`))
				Expect(testUI.Out).To(Say(`org:\s+some-org`))
				Expect(testUI.Out).To(Say(`apps:\s+app1, app2, app3`))
				Expect(testUI.Out).To(Say(`services:\s+instance1, instance2`))
				Expect(testUI.Out).To(Say(`isolation segment:`))
				Expect(testUI.Out).To(Say(`running security groups:\s+%s, %s`, runningSecurityGroup1.Name, runningSecurityGroup2.Name))
				Expect(testUI.Out).To(Say(`staging security groups:\s+%s`, stagingSecurityGroup.Name))

				Expect(testUI.Out).To(Say(`security group\s+destination\s+ports\s+protocol\s+lifecycle\s+description`))
				Expect(testUI.Out).To(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
					runningSecurityGroup2.Name, runningSecurityGroup2.Rules[0].Destination, "", runningSecurityGroup2.Rules[0].Protocol, "running", description,
				))
				Expect(testUI.Out).To(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
					runningSecurityGroup1.Name, runningSecurityGroup1.Rules[0].Destination, ports, runningSecurityGroup1.Rules[0].Protocol, "running", description,
				))
				Expect(testUI.Out).To(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s`,
					stagingSecurityGroup.Name, stagingSecurityGroup.Rules[0].Destination, ports, stagingSecurityGroup.Rules[0].Protocol, "staging", "",
				))
			})

			When("there are no security groups to show", func() {
				BeforeEach(func() {
					fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
						v7action.SpaceSummary{},
						v7action.Warnings{"some-warning"},
						nil,
					)
				})

				It("displays an empty message", func() {
					Expect(testUI.Out).To(Say(`running security groups:\s*\n`))
					Expect(testUI.Out).To(Say(`staging security groups:\s*\n`))
					Expect(testUI.Out).To(Say("No security group rules found."))
				})
			})
		})

		When("fetching the space summary succeeds with an isolation segment", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{
						Name:                  "some-space",
						OrgName:               "some-org",
						AppNames:              []string{"app1", "app2", "app3"},
						ServiceInstanceNames:  []string{"instance1", "instance2"},
						IsolationSegmentName:  "iso-seg-name",
						RunningSecurityGroups: []resources.SecurityGroup{runningSecurityGroup1, runningSecurityGroup2},
						StagingSecurityGroups: []resources.SecurityGroup{stagingSecurityGroup},
					},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays flavor text", func() {
				Expect(testUI.Out).To(Say("Getting info for space some-space in org some-org as steve..."))
			})

			It("displays warnings", func() {
				Expect(testUI.Err).To(Say("some-warning"))
			})

			It("displays a table of values", func() {
				Expect(testUI.Out).To(Say(`name:\s+some-space`))
				Expect(testUI.Out).To(Say(`org:\s+some-org`))
				Expect(testUI.Out).To(Say(`apps:\s+app1, app2, app3`))
				Expect(testUI.Out).To(Say(`services:\s+instance1, instance2`))
				Expect(testUI.Out).To(Say(`isolation segment:\s+iso-seg-name`))
				Expect(testUI.Out).To(Say(`running security groups:\s+%s, %s`, runningSecurityGroup1.Name, runningSecurityGroup2.Name))
				Expect(testUI.Out).To(Say(`staging security groups:\s+%s`, stagingSecurityGroup.Name))
			})
		})

		When("fetching the space summary succeeds without an isolation segment", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{
						Name:                  "some-space",
						OrgName:               "some-org",
						AppNames:              []string{"app1", "app2", "app3"},
						ServiceInstanceNames:  []string{"instance1", "instance2"},
						RunningSecurityGroups: []resources.SecurityGroup{runningSecurityGroup1, runningSecurityGroup2},
						StagingSecurityGroups: []resources.SecurityGroup{stagingSecurityGroup},
					},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays flavor text", func() {
				Expect(testUI.Out).To(Say("Getting info for space some-space in org some-org as steve..."))
			})

			It("displays warnings", func() {
				Expect(testUI.Err).To(Say("some-warning"))
			})

			It("displays a table of values", func() {
				Expect(testUI.Out).To(Say(`name:\s+some-space`))
				Expect(testUI.Out).To(Say(`org:\s+some-org`))
				Expect(testUI.Out).To(Say(`apps:\s+app1, app2, app3`))
				Expect(testUI.Out).To(Say(`services:\s+instance1, instance2`))
				Expect(testUI.Out).To(Say(`isolation segment:`))
				Expect(testUI.Out).To(Say(`running security groups:\s+%s, %s`, runningSecurityGroup1.Name, runningSecurityGroup2.Name))
				Expect(testUI.Out).To(Say(`staging security groups:\s+%s`, stagingSecurityGroup.Name))
			})
		})

		When("fetching a space with an applied quota", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{
						Name:                 "some-space",
						OrgName:              "some-org",
						AppNames:             []string{"app1", "app2", "app3"},
						ServiceInstanceNames: []string{"instance1", "instance2"},
						IsolationSegmentName: "iso-seg-name",
						QuotaName:            "applied-quota-name",
					},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays the applied quota", func() {
				Expect(testUI.Out).To(Say(`quota:\s+applied-quota-name`))
			})
		})

		When("fetching a space that has no quota applied", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{
						Name:                 "some-space",
						OrgName:              "some-org",
						AppNames:             []string{"app1", "app2", "app3"},
						ServiceInstanceNames: []string{"instance1", "instance2"},
					},
					v7action.Warnings{"some-warning"},
					nil,
				)
			})

			It("displays a quota row with no value", func() {
				Expect(testUI.Out).To(Say(`quota:\s*\n`))
			})
		})

		When("fetching the space summary fails", func() {
			BeforeEach(func() {
				fakeActor.GetSpaceSummaryByNameAndOrganizationReturns(
					v7action.SpaceSummary{},
					v7action.Warnings{"some-warning"},
					errors.New("get-summary-error"),
				)
			})

			It("displays warnings and returns the error", func() {
				Expect(testUI.Err).To(Say("some-warning"))
				Expect(executeErr).To(MatchError("get-summary-error"))
			})
		})
	})
})
