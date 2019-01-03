package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("marketplace Command", func() {
	var (
		cmd             MarketplaceCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeServicesSummariesActor
		binaryName      string
		executeErr      error
		extraArgs       []string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeServicesSummariesActor)

		cmd = MarketplaceCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
		}

		binaryName = "faceman"
		fakeConfig.BinaryNameReturns(binaryName)
		extraArgs = nil
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(extraArgs)
	})

	When("too many arguments are provided", func() {
		BeforeEach(func() {
			extraArgs = []string{"extra"}
		})

		It("returns a TooManyArgumentsError", func() {
			Expect(executeErr).To(MatchError(translatableerror.TooManyArgumentsError{
				ExtraArgument: "extra",
			}))
		})
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.IsLoggedInReturns(false)
		})

		Context("and the -s flag is passed", func() {
			BeforeEach(func() {
				cmd.ServiceName = "service-a"
			})

			When("a service exists that has has multiple plans", func() {
				BeforeEach(func() {
					serviceSummary := v2action.ServiceSummary{
						Service: v2action.Service{
							Label:       "service-a",
							Description: "fake service",
						},
						Plans: []v2action.ServicePlanSummary{
							{
								ServicePlan: v2action.ServicePlan{
									Name:        "plan-a",
									Description: "plan-a-description",
									Free:        false,
								},
							},
							{
								ServicePlan: v2action.ServicePlan{
									Name:        "plan-b",
									Description: "plan-b-description",
									Free:        true,
								},
							},
						},
					}

					fakeActor.GetServiceSummaryByNameReturns(serviceSummary, v2action.Warnings{"warning"}, nil)
				})

				It("outputs a header", func() {
					Expect(testUI.Out).To(Say("Getting service plan information for service service-a\\.\\.\\."))
				})

				It("outputs OK", func() {
					Expect(testUI.Out).To(Say("OK"))
				})

				It("outputs details about the specific service", func() {
					Expect(testUI.Out).Should(Say(
						"service plan\\s+description\\s+free or paid\n" +
							"plan-a\\s+plan-a-description\\s+paid" +
							"\nplan-b\\s+plan-b-description\\s+free"))
				})

				It("outputs any warnings", func() {
					Expect(testUI.Err).To(Say("warning"))
				})
			})

			When("there is an error getting the service", func() {
				BeforeEach(func() {
					fakeActor.GetServiceSummaryByNameReturns(v2action.ServiceSummary{}, v2action.Warnings{"warning"}, errors.New("oops"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("oops"))
				})

				It("outputs any warnings", func() {
					Expect(testUI.Err).To(Say("warning"))
				})
			})
		})

		When("there are no flags passed", func() {
			When("there are no services available", func() {
				BeforeEach(func() {
					fakeActor.GetServicesSummariesReturns([]v2action.ServiceSummary{}, v2action.Warnings{}, nil)
				})

				It("outputs a header", func() {
					Expect(testUI.Out).To(Say("Getting all services from marketplace\\.\\.\\."))
				})

				It("outputs OK", func() {
					Expect(testUI.Out).To(Say("OK"))
				})

				It("outputs that none are available", func() {
					Expect(testUI.Out).To(Say("No service offerings found"))
				})
			})

			When("there are multiple services available", func() {
				BeforeEach(func() {
					servicesSummaries := []v2action.ServiceSummary{
						{
							Service: v2action.Service{
								Label:             "service-a",
								Description:       "fake service-a",
								ServiceBrokerName: "broker-a",
							},
							Plans: []v2action.ServicePlanSummary{
								{
									ServicePlan: v2action.ServicePlan{Name: "plan-a"},
								},
								{
									ServicePlan: v2action.ServicePlan{Name: "plan-b"},
								},
							},
						},
						{
							Service: v2action.Service{
								Label:             "service-b",
								Description:       "fake service-b",
								ServiceBrokerName: "broker-b",
							},
							Plans: []v2action.ServicePlanSummary{
								{
									ServicePlan: v2action.ServicePlan{Name: "plan-c"},
								},
							},
						},
					}

					fakeActor.GetServicesSummariesReturns(servicesSummaries, v2action.Warnings{"warning"}, nil)
				})

				It("outputs a header", func() {
					Expect(testUI.Out).To(Say("Getting all services from marketplace\\.\\.\\."))
				})

				It("outputs OK", func() {
					Expect(testUI.Out).To(Say("OK"))
				})

				It("outputs available services and plans", func() {
					Expect(testUI.Out).Should(Say("service\\s+plans\\s+description\\s+broker\n" +
						"service-a\\s+plan-a, plan-b\\s+fake service-a\\s+broker-a\n" +
						"service-b\\s+plan-c\\s+fake service-b\\s+broker-b"))
				})

				It("outputs a tip to use the -s flag", func() {
					Expect(testUI.Out).To(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
				})

				It("outputs any warnings", func() {
					Expect(testUI.Err).To(Say("warning"))
				})
			})

			When("there is an error getting the available services", func() {
				BeforeEach(func() {
					fakeActor.GetServicesSummariesReturns([]v2action.ServiceSummary{}, v2action.Warnings{"warning"}, errors.New("oops"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("oops"))
				})

				It("outputs any warnings", func() {
					Expect(testUI.Err).To(Say("warning"))
				})
			})
		})
	})

	When("the user is logged in but not targeting an org", func() {
		BeforeEach(func() {
			fakeSharedActor.IsLoggedInReturns(true)
			fakeSharedActor.IsOrgTargetedReturns(false)
		})

		When("no flags are passed", func() {
			It("returns an error saying the user must have a space targeted", func() {
				Expect(executeErr).To(MatchError("Cannot list marketplace services without a targeted space"))
			})
		})

		When("the -s flag is passed", func() {
			BeforeEach(func() {
				cmd.ServiceName = "service-a"
			})

			It("returns an error saying the user must have a space targeted", func() {
				Expect(executeErr).To(MatchError("Cannot list plan information for service-a without a targeted space"))
			})
		})
	})

	When("the user is logged in and targeting and org but not a space", func() {
		BeforeEach(func() {
			fakeSharedActor.IsLoggedInReturns(true)
			fakeSharedActor.IsOrgTargetedReturns(true)
			fakeSharedActor.IsSpaceTargetedReturns(false)
		})

		When("no flags are passed", func() {
			It("returns an error saying the user must have a space targeted", func() {
				Expect(executeErr).To(MatchError("Cannot list marketplace services without a targeted space"))
			})
		})

		When("the -s flag is passed", func() {
			BeforeEach(func() {
				cmd.ServiceName = "service-a"
			})

			It("returns an error saying the user must have a space targeted", func() {
				Expect(executeErr).To(MatchError("Cannot list plan information for service-a without a targeted space"))
			})
		})
	})

	When("the user is logged in and targeting an org and space", func() {
		BeforeEach(func() {
			fakeSharedActor.IsLoggedInReturns(true)
			fakeSharedActor.IsOrgTargetedReturns(true)
			fakeSharedActor.IsSpaceTargetedReturns(true)

			fakeConfig.TargetedOrganizationReturns(configv3.Organization{Name: "org-a"})
			fakeConfig.TargetedSpaceReturns(configv3.Space{Name: "space-a", GUID: "space-guid"})
		})

		When("fetching the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("kaboom"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("kaboom"))
			})
		})

		When("fetching the user succeeds", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{Name: "user-a"}, nil)
			})

			When("the -s flag is passed", func() {
				BeforeEach(func() {
					cmd.ServiceName = "service-a"
				})

				When("a service exists that has has multiple plans", func() {
					BeforeEach(func() {
						serviceSummary := v2action.ServiceSummary{
							Service: v2action.Service{
								Label:       "service-a",
								Description: "fake service",
							},
							Plans: []v2action.ServicePlanSummary{
								{
									ServicePlan: v2action.ServicePlan{
										Name:        "plan-a",
										Description: "plan-a-description",
										Free:        false,
									},
								},
								{
									ServicePlan: v2action.ServicePlan{
										Name:        "plan-b",
										Description: "plan-b-description",
										Free:        true,
									},
								},
							},
						}

						fakeActor.GetServiceSummaryForSpaceByNameReturns(serviceSummary, v2action.Warnings{"warning"}, nil)
					})

					It("outputs a header", func() {
						Expect(testUI.Out).To(Say("Getting service plan information for service service-a as user-a\\.\\.\\."))
					})

					It("outputs OK", func() {
						Expect(testUI.Out).To(Say("OK"))
					})

					It("outputs details about the specific service", func() {
						Expect(testUI.Out).Should(Say(
							"service plan\\s+description\\s+free or paid\n" +
								"plan-a\\s+plan-a-description\\s+paid" +
								"\nplan-b\\s+plan-b-description\\s+free"))
					})

					It("outputs any warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
					})
				})

				When("there is an error getting the service", func() {
					BeforeEach(func() {
						fakeActor.GetServiceSummaryForSpaceByNameReturns(v2action.ServiceSummary{}, v2action.Warnings{"warning"}, errors.New("oops"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("oops"))
					})

					It("outputs any warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
					})
				})
			})

			Context("and there are no flags passed", func() {
				When("there are no services available", func() {
					BeforeEach(func() {
						fakeActor.GetServicesSummariesForSpaceReturns([]v2action.ServiceSummary{}, v2action.Warnings{}, nil)
					})

					It("gets services for the correct space", func() {
						Expect(fakeActor.GetServicesSummariesForSpaceArgsForCall(0)).To(Equal("space-guid"))
					})

					It("outputs a header", func() {
						Expect(testUI.Out).To(Say("Getting services from marketplace in org org-a / space space-a as user-a\\.\\.\\."))
					})

					It("outputs OK", func() {
						Expect(testUI.Out).To(Say("OK"))
					})

					It("outputs that no services are available", func() {
						Expect(testUI.Out).To(Say("No service offerings found"))
					})
				})

				When("there are multiple services available", func() {
					BeforeEach(func() {
						servicesSummaries := []v2action.ServiceSummary{
							{
								Service: v2action.Service{
									Label:             "service-a",
									Description:       "fake service-a",
									ServiceBrokerName: "broker-a",
								},
								Plans: []v2action.ServicePlanSummary{
									{
										ServicePlan: v2action.ServicePlan{Name: "plan-a"},
									},
									{
										ServicePlan: v2action.ServicePlan{Name: "plan-b"},
									},
								},
							},
							{
								Service: v2action.Service{
									Label:             "service-b",
									Description:       "fake service-b",
									ServiceBrokerName: "broker-b",
								},
								Plans: []v2action.ServicePlanSummary{
									{
										ServicePlan: v2action.ServicePlan{Name: "plan-c"},
									},
								},
							},
						}

						fakeActor.GetServicesSummariesForSpaceReturns(servicesSummaries, v2action.Warnings{"warning"}, nil)
					})

					It("gets services for the correct space", func() {
						Expect(fakeActor.GetServicesSummariesForSpaceArgsForCall(0)).To(Equal("space-guid"))
					})

					It("outputs a header", func() {
						Expect(testUI.Out).To(Say("Getting services from marketplace in org org-a / space space-a as user-a\\.\\.\\."))
					})

					It("outputs OK", func() {
						Expect(testUI.Out).To(Say("OK"))
					})

					It("outputs available services and plans", func() {
						Expect(testUI.Out).Should(Say("service\\s+plans\\s+description\\s+broker\n" +
							"service-a\\s+plan-a, plan-b\\s+fake service-a\\s+broker-a\n" +
							"service-b\\s+plan-c\\s+fake service-b\\s+broker-b"))
					})

					It("outputs a tip to use the -s flag", func() {
						Expect(testUI.Out).To(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
					})

					It("outputs any warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
					})
				})

				When("there is an error getting the available services", func() {
					BeforeEach(func() {
						fakeActor.GetServicesSummariesForSpaceReturns([]v2action.ServiceSummary{}, v2action.Warnings{"warning"}, errors.New("oops"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError("oops"))
					})

					It("outputs any warnings", func() {
						Expect(testUI.Err).To(Say("warning"))
					})
				})
			})
		})
	})

})
