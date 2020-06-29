package v7_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v7"
	"code.cloudfoundry.org/cli/command/v7/v7fakes"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/pkg/errors"
)

var _ = Describe("marketplace command", func() {
	var (
		cmd             MarketplaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v7fakes.FakeActor
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v7fakes.FakeActor)

		cmd = MarketplaceCommand{
			BaseCommand: BaseCommand{
				UI:          testUI,
				Config:      fakeConfig,
				SharedActor: fakeSharedActor,
				Actor:       fakeActor,
			},
		}

		fakeConfig.TargetedSpaceReturns(configv3.Space{
			GUID: "fake-space-guid",
			Name: "fake-space-name",
		})

		fakeConfig.TargetedOrganizationReturns(configv3.Organization{
			GUID: "fake-org-guid",
			Name: "fake-org-name",
		})

		fakeConfig.CurrentUserReturns(configv3.User{Name: "fake-username"}, nil)
	})

	Describe("pre-flight checks", func() {
		var executeErr error

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		It("checks the login status", func() {
			Expect(fakeSharedActor.IsLoggedInCallCount()).To(Equal(1))
		})

		When("logged in", func() {
			BeforeEach(func() {
				fakeSharedActor.IsLoggedInReturns(true)
			})

			It("gets the user", func() {
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(1))
			})

			It("checks the target", func() {
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))
				checkOrg, checkSpace := fakeSharedActor.CheckTargetArgsForCall(0)
				Expect(checkOrg).To(BeTrue())
				Expect(checkSpace).To(BeTrue())
			})

			When("getting the user fails", func() {
				BeforeEach(func() {
					fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("fake get user error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("fake get user error"))
				})
			})

			When("checking the target fails", func() {
				BeforeEach(func() {
					fakeSharedActor.CheckTargetReturns(errors.New("fake target error"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError("fake target error"))
				})
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				fakeSharedActor.IsLoggedInReturns(false)
			})

			It("does not try to get the username or check the target", func() {
				Expect(fakeConfig.CurrentUserCallCount()).To(Equal(0))
				Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(0))
			})
		})

		When("the -e and --no-plans flags to be specified together", func() {
			BeforeEach(func() {
				setFlag(&cmd, "--no-plans")
				setFlag(&cmd, "-e", "foo")
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError(translatableerror.ArgumentCombinationError{
					Args: []string{"--no-plans", "-e"},
				}))
			})
		})
	})

	DescribeTable(
		"printing the action",
		func(loggedIn bool, flags map[string]interface{}, message string) {
			fakeSharedActor.IsLoggedInReturns(loggedIn)

			for k, v := range flags {
				setFlag(&cmd, k, v)
			}

			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(testUI.Out).To(Say(message))
		},
		Entry(
			"not logged in with no flags",
			false,
			map[string]interface{}{},
			`Getting all service offerings from marketplace\.\.\.`,
		),
		Entry(
			"not logged in with -e flag",
			false,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
			},
			`Getting service plan information for service offering fake-service-offering-name\.\.\.`,
		),
		Entry(
			"not logged in with -b flag",
			false,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
			},
			`Getting all service offerings from marketplace for service broker fake-service-broker-name\.\.\.`,
		),
		Entry(
			"not logged in with -e and -b flag",
			false,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
				"-e": "fake-service-offering-name",
			},
			`Getting service plan information for service offering fake-service-offering-name from service broker fake-service-broker-name\.\.\.`,
		),
		Entry(
			"logged in with no flags",
			true,
			map[string]interface{}{},
			`Getting all service offerings from marketplace in org fake-org-name / space fake-space-name as fake-username\.\.\.`,
		),
		Entry(
			"logged in with -e flag",
			true,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
			},
			`Getting service plan information for service offering fake-service-offering-name in org fake-org-name / space fake-space-name as fake-username\.\.\.`,
		),
		Entry(
			"logged in with -b flag",
			true,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
			},
			`Getting all service offerings from marketplace for service broker fake-service-broker-name in org fake-org-name / space fake-space-name as fake-username\.\.\.`,
		),
		Entry(
			"logged in with -e and -b flag",
			true,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
				"-b": "fake-service-broker-name",
			},
			`Getting service plan information for service offering fake-service-offering-name from service broker fake-service-broker-name in org fake-org-name / space fake-space-name as fake-username\.\.\.`,
		),
	)

	DescribeTable(
		"sending the filter to the actor",
		func(loggedIn bool, flags map[string]interface{}, expectedFilter v7action.MarketplaceFilter) {
			fakeSharedActor.IsLoggedInReturns(loggedIn)

			for k, v := range flags {
				setFlag(&cmd, k, v)
			}

			err := cmd.Execute(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeActor.MarketplaceCallCount()).To(Equal(1))
			Expect(fakeActor.MarketplaceArgsForCall(0)).To(Equal(expectedFilter))
		},
		Entry(
			"not logged in with no flags",
			false,
			map[string]interface{}{},
			v7action.MarketplaceFilter{},
		),
		Entry(
			"not logged in with -e flag",
			false,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
			},
			v7action.MarketplaceFilter{
				ServiceOfferingName: "fake-service-offering-name",
			},
		),
		Entry(
			"not logged in with -b flag",
			false,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
			},
			v7action.MarketplaceFilter{
				ServiceBrokerName: "fake-service-broker-name",
			},
		),
		Entry(
			"not logged in with -e and -b flag",
			false,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
				"-e": "fake-service-offering-name",
			},
			v7action.MarketplaceFilter{
				ServiceOfferingName: "fake-service-offering-name",
				ServiceBrokerName:   "fake-service-broker-name",
			},
		),
		Entry(
			"logged in with no flags",
			true,
			map[string]interface{}{},
			v7action.MarketplaceFilter{
				SpaceGUID: "fake-space-guid",
			},
		),
		Entry(
			"logged in with -e flag",
			true,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
			},
			v7action.MarketplaceFilter{
				SpaceGUID:           "fake-space-guid",
				ServiceOfferingName: "fake-service-offering-name",
			},
		),
		Entry(
			"logged in with -b flag",
			true,
			map[string]interface{}{
				"-b": "fake-service-broker-name",
			},
			v7action.MarketplaceFilter{
				SpaceGUID:         "fake-space-guid",
				ServiceBrokerName: "fake-service-broker-name",
			},
		),
		Entry(
			"logged in with -e and -b flag",
			true,
			map[string]interface{}{
				"-e": "fake-service-offering-name",
				"-b": "fake-service-broker-name",
			},
			v7action.MarketplaceFilter{
				SpaceGUID:           "fake-space-guid",
				ServiceOfferingName: "fake-service-offering-name",
				ServiceBrokerName:   "fake-service-broker-name",
			},
		),
	)

	Describe("handling the result from the actor", func() {
		var executeErr error

		JustBeforeEach(func() {
			executeErr = cmd.Execute(nil)
		})

		When("the actor returns warnings", func() {
			BeforeEach(func() {
				fakeActor.MarketplaceReturns(
					[]v7action.ServiceOfferingWithPlans{{}},
					v7action.Warnings{"warning 1", "warning 2"},
					nil,
				)
			})

			It("prints then", func() {
				Expect(testUI.Err).To(Say(`warning 1`))
				Expect(testUI.Err).To(Say(`warning 2`))
			})
		})

		When("the actor returns an error", func() {
			BeforeEach(func() {
				fakeActor.MarketplaceReturns(
					[]v7action.ServiceOfferingWithPlans{{}},
					v7action.Warnings{"warning 1", "warning 2"},
					errors.New("awful error"),
				)
			})

			It("prints warnings and returns an error", func() {
				Expect(testUI.Err).To(Say(`warning 1`))
				Expect(testUI.Err).To(Say(`warning 2`))
				Expect(executeErr).To(MatchError("awful error"))
			})
		})

		When("no offerings are returned", func() {
			BeforeEach(func() {
				fakeActor.MarketplaceReturns(
					[]v7action.ServiceOfferingWithPlans{},
					v7action.Warnings{"warning 1", "warning 2"},
					nil,
				)
			})

			It("says that no service offerings were found", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`\n\n`))
				Expect(testUI.Out).To(Say(`No service offerings found.`))

				Expect(testUI.Err).To(Say("warning 1"))
				Expect(testUI.Err).To(Say("warning 2"))
			})
		})

		When("showing the service offerings table", func() {
			BeforeEach(func() {
				fakeActor.MarketplaceReturns(
					[]v7action.ServiceOfferingWithPlans{
						{
							GUID:              "offering-guid-1",
							Name:              "offering-1",
							Description:       "about offering 1",
							ServiceBrokerName: "service-broker-1",
							Plans: []resources.ServicePlan{
								{
									GUID: "plan-guid-1",
									Name: "plan-1",
								},
							},
						},
						{
							GUID:              "offering-guid-2",
							Name:              "offering-2",
							Description:       "about offering 2",
							ServiceBrokerName: "service-broker-2",
							Plans: []resources.ServicePlan{
								{
									GUID: "plan-guid-2",
									Name: "plan-2",
								},
								{
									GUID: "plan-guid-3",
									Name: "plan-3",
								},
							},
						},
					},
					v7action.Warnings{"warning 1", "warning 2"},
					nil,
				)
			})

			It("prints a table showing service offerings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`\n\n`))
				Expect(testUI.Out).To(Say(`offering\s+plans\s+description\s+broker`))
				Expect(testUI.Out).To(Say(`offering-1\s+plan-1\s+about offering 1\s+service-broker-1`))
				Expect(testUI.Out).To(Say(`offering-2\s+plan-2, plan-3\s+about offering 2\s+service-broker-2`))
				Expect(testUI.Out).To(Say(`\n\n`))
				Expect(testUI.Out).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))

				Expect(testUI.Err).To(Say("warning 1"))
				Expect(testUI.Err).To(Say("warning 2"))
			})

			When("the --no-plans flag is specified", func() {
				BeforeEach(func() {
					setFlag(&cmd, "--no-plans")
				})

				It("prints a table showing service offerings without plan names", func() {
					Expect(executeErr).NotTo(HaveOccurred())

					Expect(testUI.Out).To(Say(`\n\n`))
					Expect(testUI.Out).To(Say(`offering\s+description\s+broker`))
					Expect(testUI.Out).To(Say(`offering-1\s+about offering 1\s+service-broker-1`))
					Expect(testUI.Out).To(Say(`offering-2\s+about offering 2\s+service-broker-2`))
					Expect(testUI.Out).To(Say(`\n\n`))
					Expect(testUI.Out).To(Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`))

					Expect(testUI.Err).To(Say("warning 1"))
					Expect(testUI.Err).To(Say("warning 2"))
				})
			})
		})

		When("showing the service plans table", func() {
			BeforeEach(func() {
				setFlag(&cmd, "-e", "fake-service-offering-name")

				fakeActor.MarketplaceReturns(
					[]v7action.ServiceOfferingWithPlans{
						{
							GUID:              "offering-guid-1",
							Name:              "interesting-name",
							Description:       "about offering 1",
							ServiceBrokerName: "service-broker-1",
							Plans: []resources.ServicePlan{
								{
									GUID:        "plan-guid-1",
									Name:        "plan-1",
									Description: "best available plan",
									Available:   true,
									Free:        true,
								},
							},
						},
						{
							GUID:              "offering-guid-2",
							Name:              "interesting-name",
							Description:       "about offering 2",
							ServiceBrokerName: "service-broker-2",
							Plans: []resources.ServicePlan{
								{
									GUID:        "plan-guid-2",
									Name:        "plan-2",
									Description: "just another plan",
									Available:   true,
									Free:        false,
									Costs: []resources.ServicePlanCost{
										{
											Currency: "USD",
											Amount:   100,
											Unit:     "Monthly",
										},
										{
											Currency: "USD",
											Amount:   0.999,
											Unit:     "1GB of messages over 20GB",
										},
									},
								},
								{
									GUID:      "plan-guid-3",
									Name:      "plan-3",
									Free:      true,
									Available: false,
								},
								{
									GUID:      "plan-guid-4",
									Name:      "plan-4",
									Free:      false,
									Available: true,
								},
							},
						},
					},
					v7action.Warnings{"warning 1", "warning 2"},
					nil,
				)
			})

			It("prints a table showing service plans", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(testUI.Out).To(Say(`\n\n`))
				Expect(testUI.Out).To(Say(`broker: service-broker-1`))
				Expect(testUI.Out).To(Say(`plan\s+description\s+free or paid\s+costs\s+available`))
				Expect(testUI.Out).To(Say(`plan-1\s+best available plan\s+free\s+yes`))
				Expect(testUI.Out).To(Say(`\n\n`))
				Expect(testUI.Out).To(Say(`broker: service-broker-2`))
				Expect(testUI.Out).To(Say(`plan\s+description\s+free or paid\s+costs\s+available`))
				Expect(testUI.Out).To(Say(`plan-2\s+just another plan\s+paid\s+USD 100.00/Monthly, USD 1.00/1GB of messages over 20GB\s+yes`))
				Expect(testUI.Out).To(Say(`plan-3\s+free\s+no`))
				Expect(testUI.Out).To(Say(`plan-4\s+paid\s+yes`))

				Expect(testUI.Err).To(Say("warning 1"))
				Expect(testUI.Err).To(Say("warning 2"))
			})
		})
	})
})
