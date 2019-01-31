package composite_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	. "code.cloudfoundry.org/cli/actor/v2action/composite"
	"code.cloudfoundry.org/cli/actor/v2action/composite/compositefakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Broker Summary Actions", func() {
	var (
		fakeServiceActor    *compositefakes.FakeServiceActor
		fakeBrokerActor     *compositefakes.FakeBrokerActor
		fakeOrgActor        *compositefakes.FakeOrganizationActor
		fakeVisibilityActor *compositefakes.FakeVisibilityActor
		actor               *ServiceBrokerSummaryCompositeActor
	)

	BeforeEach(func() {
		fakeOrgActor = new(compositefakes.FakeOrganizationActor)
		fakeServiceActor = new(compositefakes.FakeServiceActor)
		fakeBrokerActor = new(compositefakes.FakeBrokerActor)
		fakeVisibilityActor = new(compositefakes.FakeVisibilityActor)
		actor = &ServiceBrokerSummaryCompositeActor{
			ServiceActor:    fakeServiceActor,
			BrokerActor:     fakeBrokerActor,
			OrgActor:        fakeOrgActor,
			VisibilityActor: fakeVisibilityActor,
		}
	})

	Describe("GetServiceBrokerSummaries", func() {
		var (
			broker       string
			service      string
			organization string

			summaries  []v2action.ServiceBrokerSummary
			warnings   v2action.Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			summaries, warnings, executeErr = actor.GetServiceBrokerSummaries(broker, service, organization)
		})

		When("no broker, service, organization is specified", func() {
			BeforeEach(func() {
				broker = ""
				service = ""
				organization = ""
			})

			When("fetching the service broker is successful", func() {
				BeforeEach(func() {
					fakeBrokerActor.GetServiceBrokersReturns(
						[]v2action.ServiceBroker{
							{Name: "broker-1", GUID: "broker-guid-1"},
							{Name: "broker-2", GUID: "broker-guid-2"},
						},
						v2action.Warnings{"get-brokers-warning"}, nil)
				})

				When("the brokers contain no services", func() {
					It("returns expected Service Brokers", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-brokers-warning"))
						Expect(summaries).To(ConsistOf(
							v2action.ServiceBrokerSummary{
								ServiceBroker: v2action.ServiceBroker{
									Name: "broker-1",
									GUID: "broker-guid-1",
								},
							},
							v2action.ServiceBrokerSummary{
								ServiceBroker: v2action.ServiceBroker{
									Name: "broker-2",
									GUID: "broker-guid-2",
								},
							},
						))

						Expect(fakeBrokerActor.GetServiceBrokersCallCount()).To(Equal(1))
					})
				})

				When("the brokers contain services", func() {
					When("fetching the services is successful", func() {
						When("the services contain no plans", func() {
							BeforeEach(func() {
								fakeServiceActor.GetServicesWithPlansReturnsOnCall(0,
									v2action.ServicesWithPlans{
										v2action.Service{Label: "service-1", GUID: "service-guid-1"}: nil,
										v2action.Service{Label: "service-2", GUID: "service-guid-2"}: nil,
									}, v2action.Warnings{"service-warning-1"}, nil)
								fakeServiceActor.GetServicesWithPlansReturnsOnCall(1,
									v2action.ServicesWithPlans{
										v2action.Service{Label: "service-3", GUID: "service-guid-3"}: nil,
										v2action.Service{Label: "service-4", GUID: "service-guid-4"}: nil,
									}, v2action.Warnings{"service-warning-2"}, nil)
							})

							It("returns expected Services for their given brokers", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("get-brokers-warning", "service-warning-1", "service-warning-2"))
								Expect(summaries[0].Services).To(ConsistOf(
									[]v2action.ServiceSummary{
										{
											Service: v2action.Service{Label: "service-1", GUID: "service-guid-1"},
										},
										{
											Service: v2action.Service{Label: "service-2", GUID: "service-guid-2"},
										},
									},
								))
								Expect(summaries[1].Services).To(ConsistOf(
									[]v2action.ServiceSummary{
										{
											Service: v2action.Service{Label: "service-3", GUID: "service-guid-3"},
										},
										{
											Service: v2action.Service{Label: "service-4", GUID: "service-guid-4"},
										},
									},
								))

								Expect(fakeServiceActor.GetServicesWithPlansCallCount()).To(Equal(2))
								Expect(fakeServiceActor.GetServicesWithPlansArgsForCall(0)).To(ConsistOf(
									v2action.Filter{
										Type:     constant.ServiceBrokerGUIDFilter,
										Operator: constant.EqualOperator,
										Values:   []string{"broker-guid-1"},
									},
								))
								Expect(fakeServiceActor.GetServicesWithPlansArgsForCall(1)).To(ConsistOf(
									v2action.Filter{
										Type:     constant.ServiceBrokerGUIDFilter,
										Operator: constant.EqualOperator,
										Values:   []string{"broker-guid-2"},
									},
								))
							})
						})

						When("the services contain plans", func() {
							When("fetching service plans is successful", func() {
								When("all plans are public", func() {
									BeforeEach(func() {
										fakeServiceActor.GetServicesWithPlansReturnsOnCall(0, v2action.ServicesWithPlans{
											v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
												{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
												{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: true},
											},
											v2action.Service{Label: "service-2", GUID: "service-guid-2"}: nil,
										}, v2action.Warnings{"service-warning-1"}, nil)

										fakeServiceActor.GetServicesWithPlansReturnsOnCall(1, v2action.ServicesWithPlans{
											v2action.Service{Label: "service-3", GUID: "service-guid-3"}: []v2action.ServicePlan{
												{GUID: "service-plan-guid-3", Name: "service-plan-3", Public: true},
												{GUID: "service-plan-guid-4", Name: "service-plan-4", Public: true},
											},
											v2action.Service{Label: "service-4", GUID: "service-guid-4"}: nil,
										}, v2action.Warnings{"service-warning-2"}, nil)
									})

									It("returns the expected Service with Plans without visibilities", func() {
										Expect(executeErr).ToNot(HaveOccurred())
										Expect(warnings).To(ConsistOf(
											"get-brokers-warning", "service-warning-1", "service-warning-2"))

										Expect(summaries[0].Services).To(ConsistOf(
											v2action.ServiceSummary{
												Service: v2action.Service{Label: "service-1", GUID: "service-guid-1"},
												Plans: []v2action.ServicePlanSummary{
													{
														ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
													},
													{
														ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: true},
													},
												},
											},
											v2action.ServiceSummary{
												Service: v2action.Service{Label: "service-2", GUID: "service-guid-2"},
											},
										))
										Expect(summaries[1].Services).To(ConsistOf(
											v2action.ServiceSummary{
												Service: v2action.Service{Label: "service-3", GUID: "service-guid-3"},
												Plans: []v2action.ServicePlanSummary{
													{
														ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-3", Name: "service-plan-3", Public: true},
													},
													{
														ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-4", Name: "service-plan-4", Public: true},
													},
												},
											},
											v2action.ServiceSummary{
												Service: v2action.Service{Label: "service-4", GUID: "service-guid-4"},
											},
										))

										Expect(fakeServiceActor.GetServicesWithPlansCallCount()).To(Equal(2))
										Expect(fakeServiceActor.GetServicesWithPlansArgsForCall(0)).To(ConsistOf(
											v2action.Filter{
												Type:     constant.ServiceBrokerGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"broker-guid-1"},
											},
										))
										Expect(fakeServiceActor.GetServicesWithPlansArgsForCall(1)).To(ConsistOf(
											v2action.Filter{
												Type:     constant.ServiceBrokerGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"broker-guid-2"},
											},
										))

										Expect(fakeVisibilityActor.GetServicePlanVisibilitiesCallCount()).To(Equal(0))
									})
								})

								When("there are non-public plans", func() {
									BeforeEach(func() {
										fakeServiceActor.GetServicesWithPlansReturnsOnCall(0, v2action.ServicesWithPlans{
											v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
												{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
												{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: false},
											},
											v2action.Service{Label: "service-2", GUID: "service-guid-2"}: nil,
										}, v2action.Warnings{"service-warning-1"}, nil)
										fakeServiceActor.GetServicesWithPlansReturnsOnCall(1, v2action.ServicesWithPlans{},
											v2action.Warnings{"service-warning-2"}, nil)
									})

									When("fetching orgs and service plan visibilities is successful", func() {
										BeforeEach(func() {
											fakeVisibilityActor.GetServicePlanVisibilitiesReturnsOnCall(0,
												[]v2action.ServicePlanVisibility{
													{OrganizationGUID: "org-guid-1"},
													{OrganizationGUID: "org-guid-2"},
												},
												v2action.Warnings{"service-plan-visibility-1"},
												nil)
											fakeVisibilityActor.GetServicePlanVisibilitiesReturnsOnCall(1,
												[]v2action.ServicePlanVisibility{
													{OrganizationGUID: "org-guid-3"},
													{OrganizationGUID: "org-guid-4"},
												},
												v2action.Warnings{"service-plan-visibility-2"},
												nil)

											fakeOrgActor.GetOrganizationReturnsOnCall(0,
												v2action.Organization{Name: "org-1"},
												v2action.Warnings{"org-warning-1"},
												nil)
											fakeOrgActor.GetOrganizationReturnsOnCall(1,
												v2action.Organization{Name: "org-2"},
												v2action.Warnings{"org-warning-2"},
												nil)
											fakeOrgActor.GetOrganizationReturnsOnCall(2,
												v2action.Organization{Name: "org-3"},
												v2action.Warnings{"org-warning-3"},
												nil)
											fakeOrgActor.GetOrganizationReturnsOnCall(3,
												v2action.Organization{Name: "org-4"},
												v2action.Warnings{"org-warning-4"},
												nil)
										})

										It("returns the expected Service Plans", func() {
											Expect(executeErr).ToNot(HaveOccurred())
											Expect(warnings).To(ConsistOf(
												"get-brokers-warning",
												"service-warning-1", "service-warning-2",
												"service-plan-visibility-1", "service-plan-visibility-2",
												"org-warning-1", "org-warning-2", "org-warning-3", "org-warning-4",
											))

											Expect(summaries[0].Services).To(ConsistOf(
												v2action.ServiceSummary{
													Service: v2action.Service{Label: "service-1", GUID: "service-guid-1"},
													Plans: []v2action.ServicePlanSummary{
														{
															ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
															VisibleTo:   []string{"org-1", "org-2"},
														},
														{
															ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: false},
															VisibleTo:   []string{"org-3", "org-4"},
														},
													},
												},
												v2action.ServiceSummary{
													Service: v2action.Service{Label: "service-2", GUID: "service-guid-2"},
												},
											))

											Expect(fakeVisibilityActor.GetServicePlanVisibilitiesCallCount()).To(Equal(2))
											Expect(fakeVisibilityActor.GetServicePlanVisibilitiesArgsForCall(0)).To(Equal("service-plan-guid-1"))
											Expect(fakeVisibilityActor.GetServicePlanVisibilitiesArgsForCall(1)).To(Equal("service-plan-guid-2"))
										})
									})

									When("fetching the service plan visibilities fails", func() {
										BeforeEach(func() {
											fakeVisibilityActor.GetServicePlanVisibilitiesReturns(
												nil,
												v2action.Warnings{"service-plan-visibility-1"},
												errors.New("boom"))
										})

										It("returns the error and warnings", func() {
											Expect(executeErr).To(MatchError("boom"))
											Expect(warnings).To(ConsistOf(
												"get-brokers-warning",
												"service-warning-1",
												"service-plan-visibility-1",
											))
										})
									})
								})

								When("fetching the organizations fails", func() {
									BeforeEach(func() {
										fakeServiceActor.GetServicesWithPlansReturns(v2action.ServicesWithPlans{
											v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
												{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
											},
											v2action.Service{Label: "service-2", GUID: "service-guid-2"}: nil,
										}, v2action.Warnings{"service-warning-1"}, nil)
										fakeVisibilityActor.GetServicePlanVisibilitiesReturns(
											[]v2action.ServicePlanVisibility{
												{OrganizationGUID: "org-guid-1"},
												{OrganizationGUID: "org-guid-2"},
											},
											v2action.Warnings{"service-plan-visibility-1"},
											nil)

										fakeOrgActor.GetOrganizationReturnsOnCall(0,
											v2action.Organization{},
											v2action.Warnings{"org-warning-1"},
											errors.New("boom"))
									})

									It("returns the error and warnings", func() {
										Expect(executeErr).To(MatchError("boom"))
										Expect(warnings).To(ConsistOf(
											"get-brokers-warning",
											"service-warning-1",
											"service-plan-visibility-1",
											"org-warning-1",
										))
									})
								})
							})
						})
					})
				})

				When("fetching services with plans fails", func() {
					BeforeEach(func() {
						fakeServiceActor.GetServicesWithPlansReturns(
							nil,
							v2action.Warnings{"service-warning-1"},
							errors.New("boom"),
						)
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError("boom"))
						Expect(warnings).To(ConsistOf("get-brokers-warning", "service-warning-1"))
					})
				})
			})

			When("fetching service brokers fails", func() {
				BeforeEach(func() {
					fakeBrokerActor.GetServiceBrokersReturns(nil, v2action.Warnings{"test-warning"}, errors.New("explode"))
				})

				It("returns the warnings and propagates the error", func() {
					Expect(warnings).To(ConsistOf("test-warning"))
					Expect(executeErr).To(MatchError("explode"))
				})
			})
		})

		When("service broker is specified", func() {
			BeforeEach(func() {
				broker = "broker-1"
				service = ""
				organization = ""

				fakeBrokerActor.GetServiceBrokerByNameReturns(
					v2action.ServiceBroker{Name: "broker-1", GUID: "broker-guid-1"},
					v2action.Warnings{"get-broker-warning"},
					nil)
			})

			It("returns expected Service Brokers", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-broker-warning"))
				Expect(summaries).To(ConsistOf(
					v2action.ServiceBrokerSummary{
						ServiceBroker: v2action.ServiceBroker{
							Name: "broker-1",
							GUID: "broker-guid-1",
						},
					}))

				Expect(fakeBrokerActor.GetServiceBrokerByNameCallCount()).To(Equal(1))
				Expect(fakeBrokerActor.GetServiceBrokerByNameArgsForCall(0)).To(Equal("broker-1"))
			})

			When("fetching the service broker fails", func() {
				BeforeEach(func() {
					fakeBrokerActor.GetServiceBrokerByNameReturns(
						v2action.ServiceBroker{},
						v2action.Warnings{"get-broker-warning"},
						fmt.Errorf("it broke"))
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("it broke"))
					Expect(warnings).To(ConsistOf("get-broker-warning"))
				})
			})
		})

		When("a service name is specified", func() {
			BeforeEach(func() {
				broker = ""
				service = "service-1"
				organization = ""

				fakeBrokerActor.GetServiceBrokersReturns(
					[]v2action.ServiceBroker{
						{Name: "broker-1", GUID: "broker-guid-1"},
					},
					v2action.Warnings{"get-brokers-warning"}, nil)
			})

			When("the specified service exists", func() {
				BeforeEach(func() {
					fakeServiceActor.ServiceExistsWithNameReturns(true,
						v2action.Warnings{"service-exists-with-name-warning"}, nil)
				})

				It("fetches services on brokers filtered by name", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-brokers-warning", "service-exists-with-name-warning"))

					Expect(fakeServiceActor.GetServicesWithPlansCallCount()).To(Equal(1))
					Expect(fakeServiceActor.GetServicesWithPlansArgsForCall(0)).To(ConsistOf(
						v2action.Filter{
							Type:     constant.ServiceBrokerGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"broker-guid-1"},
						},
						v2action.Filter{
							Type:     constant.LabelFilter,
							Operator: constant.EqualOperator,
							Values:   []string{"service-1"},
						},
					))
				})

				When("a particular broker has no services with that name", func() {
					BeforeEach(func() {
						fakeServiceActor.GetServicesWithPlansReturns(v2action.ServicesWithPlans{}, nil, nil)
					})

					It("does not return a summary for that broker", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(summaries).To(HaveLen(0))
					})
				})
			})

			When("checking whether the service exists returns an error", func() {
				BeforeEach(func() {
					fakeServiceActor.ServiceExistsWithNameReturns(false,
						v2action.Warnings{"service-exists-with-name-warning"}, errors.New("boom"))
				})

				It("propagates the error with warnings", func() {
					Expect(executeErr).To(MatchError("boom"))
					Expect(warnings).To(ConsistOf("service-exists-with-name-warning"))
				})
			})

			When("a service with the specified name does not exist", func() {
				BeforeEach(func() {
					fakeServiceActor.ServiceExistsWithNameReturns(false,
						v2action.Warnings{"service-exists-with-name-warning"}, nil)
				})

				It("returns an appropriate error message", func() {
					Expect(executeErr).To(MatchError(actionerror.ServiceNotFoundError{Name: "service-1"}))
					Expect(warnings).To(ConsistOf("service-exists-with-name-warning"))
				})
			})
		})

		When("an org is specified", func() {
			BeforeEach(func() {
				broker = ""
				service = ""
				organization = "my-org"

				fakeBrokerActor.GetServiceBrokersReturns(
					[]v2action.ServiceBroker{
						{Name: "broker-1", GUID: "broker-guid-1"},
					},
					v2action.Warnings{"get-brokers-warning"}, nil)
			})

			When("the specified org exists", func() {
				BeforeEach(func() {
					fakeOrgActor.OrganizationExistsWithNameReturns(
						true,
						v2action.Warnings{"org-warning-1"},
						nil)
				})

				When("plans are public", func() {
					BeforeEach(func() {
						fakeServiceActor.GetServicesWithPlansReturns(v2action.ServicesWithPlans{
							v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
								{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
							},
						}, v2action.Warnings{"get-service-plans-warning"}, nil)

					})

					It("returns those plans", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("org-warning-1", "get-brokers-warning", "get-service-plans-warning"))
						Expect(summaries[0].Services).To(ConsistOf(
							v2action.ServiceSummary{
								Service: v2action.Service{Label: "service-1", GUID: "service-guid-1"},
								Plans: []v2action.ServicePlanSummary{
									{
										ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
									},
								},
							},
						))
					})
				})

				When("all plans for all services for all brokers are private with no visibilities", func() {
					BeforeEach(func() {
						fakeServiceActor.GetServicesWithPlansReturns(v2action.ServicesWithPlans{
							v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
								{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
							},
						}, v2action.Warnings{"get-service-plans-warning"}, nil)
					})

					It("returns no broker summaries", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("org-warning-1", "get-brokers-warning", "get-service-plans-warning"))
						Expect(len(summaries)).To(Equal(0))
					})
				})

				When("a plan is visible in the provided org", func() {
					BeforeEach(func() {
						fakeServiceActor.GetServicesWithPlansReturns(v2action.ServicesWithPlans{
							v2action.Service{Label: "service-1", GUID: "service-guid-1"}: []v2action.ServicePlan{
								{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
							},
						}, v2action.Warnings{"get-service-plans-warning"}, nil)

						fakeVisibilityActor.GetServicePlanVisibilitiesReturns([]v2action.ServicePlanVisibility{
							{OrganizationGUID: "my-org-guid"},
						}, v2action.Warnings{"get-service-plan-visibility-warning"}, nil)

						fakeOrgActor.GetOrganizationReturns(v2action.Organization{Name: "my-org"},
							v2action.Warnings{"org-warning-2"}, nil)
					})

					It("should return that plan", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("org-warning-1", "get-brokers-warning", "get-service-plans-warning", "get-service-plan-visibility-warning", "org-warning-2"))
						Expect(summaries[0].Services).To(ConsistOf(
							v2action.ServiceSummary{
								Service: v2action.Service{Label: "service-1", GUID: "service-guid-1"},
								Plans: []v2action.ServicePlanSummary{
									{
										ServicePlan: v2action.ServicePlan{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: false},
										VisibleTo:   []string{"my-org"},
									},
								},
							},
						))
					})
				})
			})

			When("checking if the org exists returns an error", func() {
				BeforeEach(func() {
					fakeOrgActor.OrganizationExistsWithNameReturns(
						false,
						v2action.Warnings{"org-warning-1"},
						errors.New("boom"))
				})

				It("propagtes the error and warnings", func() {
					Expect(executeErr).To(MatchError("boom"))
					Expect(warnings).To(ConsistOf("org-warning-1"))
				})
			})

			When("the specified org does not exist", func() {
				BeforeEach(func() {
					fakeOrgActor.OrganizationExistsWithNameReturns(
						false,
						v2action.Warnings{"org-warning-1"},
						nil)
				})

				It("returns the warnings and propogates the error", func() {
					Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: "my-org"}))
					Expect(warnings).To(ConsistOf("org-warning-1"))

					Expect(fakeBrokerActor.GetServiceBrokersCallCount()).To(Equal(0))
					Expect(fakeOrgActor.OrganizationExistsWithNameCallCount()).To(Equal(1))
					Expect(fakeOrgActor.OrganizationExistsWithNameArgsForCall(0)).To(Equal("my-org"))
				})
			})
		})

		When("an org, broker and service are specified", func() {
			BeforeEach(func() {
				broker = "broker-1"
				service = "service-1"
				organization = "my-org"

				fakeOrgActor.OrganizationExistsWithNameReturns(
					true,
					v2action.Warnings{"org-warning"},
					nil)
				fakeServiceActor.ServiceExistsWithNameReturns(
					true,
					v2action.Warnings{"service-exists-with-name-warning"},
					nil)
				fakeBrokerActor.GetServiceBrokerByNameReturns(
					v2action.ServiceBroker{Name: "broker-1", GUID: "broker-guid-1"},
					v2action.Warnings{"get-broker-warning"},
					nil)
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf("org-warning", "service-exists-with-name-warning", "get-broker-warning"))
			})
		})
	})
})
