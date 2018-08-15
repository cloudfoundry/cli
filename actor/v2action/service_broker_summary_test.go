package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Broker Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServiceBrokerSummaries", func() {
		var (
			broker       string
			service      string
			organization string

			summaries  []ServiceBrokerSummary
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			summaries, warnings, executeErr = actor.GetServiceBrokerSummaries(broker, service, organization)
		})

		When("no broker, service, organization is specified", func() {
			When("fetching the service broker is successful", func() {
				BeforeEach(func() {
					broker = ""
					service = ""
					organization = ""

					fakeCloudControllerClient.GetServiceBrokersReturns(
						[]ccv2.ServiceBroker{
							{Name: "broker-1", GUID: "broker-guid-1"},
							{Name: "broker-2", GUID: "broker-guid-2"},
						},
						ccv2.Warnings{"get-brokers-warning"},
						nil)
				})

				When("there broker contains no services", func() {
					It("returns expected Service Brokers", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get-brokers-warning"))
						Expect(summaries).To(ConsistOf(
							ServiceBrokerSummary{
								ServiceBroker: ServiceBroker{
									Name: "broker-1",
									GUID: "broker-guid-1",
								},
							},
							ServiceBrokerSummary{
								ServiceBroker: ServiceBroker{
									Name: "broker-2",
									GUID: "broker-guid-2",
								},
							},
						))

						Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)).To(BeEmpty())
					})
				})

				When("the brokers contain services", func() {
					When("fetching the services is successful", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetServicesReturnsOnCall(0,
								[]ccv2.Service{
									{Label: "service-1", GUID: "service-guid-1"},
									{Label: "service-2", GUID: "service-guid-2"},
								},
								ccv2.Warnings{"service-warning-1"},
								nil)
							fakeCloudControllerClient.GetServicesReturnsOnCall(1,
								[]ccv2.Service{
									{Label: "service-3", GUID: "service-guid-3"},
									{Label: "service-4", GUID: "service-guid-4"},
								},
								ccv2.Warnings{"service-warning-2"},
								nil)
						})

						It("returns expected Services for their given brokers", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("get-brokers-warning", "service-warning-1", "service-warning-2"))
							Expect(summaries[0].Services).To(ConsistOf(
								[]ServiceSummary{
									{
										Service: Service{Label: "service-1", GUID: "service-guid-1"},
									},
									{
										Service: Service{Label: "service-2", GUID: "service-guid-2"},
									},
								},
							))
							Expect(summaries[1].Services).To(ConsistOf(
								[]ServiceSummary{
									{
										Service: Service{Label: "service-3", GUID: "service-guid-3"},
									},
									{
										Service: Service{Label: "service-4", GUID: "service-guid-4"},
									},
								},
							))

							Expect(fakeCloudControllerClient.GetServicesCallCount()).To(Equal(2))
							Expect(fakeCloudControllerClient.GetServicesArgsForCall(0)).To(ConsistOf(ccv2.Filter{
								Type:     constant.ServiceBrokerGUIDFilter,
								Operator: constant.EqualOperator,
								Values:   []string{"broker-guid-1"},
							}))
							Expect(fakeCloudControllerClient.GetServicesArgsForCall(1)).To(ConsistOf(ccv2.Filter{
								Type:     constant.ServiceBrokerGUIDFilter,
								Operator: constant.EqualOperator,
								Values:   []string{"broker-guid-2"},
							}))
						})

						When("the services contain plans", func() {
							When("fetching service plans is successful", func() {
								When("all plans are public", func() {
									BeforeEach(func() {
										fakeCloudControllerClient.GetServicePlansReturnsOnCall(0,
											[]ccv2.ServicePlan{
												{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
												{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: true},
											},
											ccv2.Warnings{"service-plan-warning-1"},
											nil)
										fakeCloudControllerClient.GetServicePlansReturnsOnCall(2,
											[]ccv2.ServicePlan{
												{GUID: "service-plan-guid-3", Name: "service-plan-3", Public: true},
												{GUID: "service-plan-guid-4", Name: "service-plan-4", Public: true},
											},
											ccv2.Warnings{"service-plan-warning-2"},
											nil)
									})

									It("returns the expected Service Plans without visibilities", func() {
										Expect(executeErr).ToNot(HaveOccurred())
										Expect(warnings).To(ConsistOf(
											"get-brokers-warning", "service-warning-1",
											"service-warning-2", "service-plan-warning-1", "service-plan-warning-2"))

										Expect(summaries[0].Services[0].Plans).To(ConsistOf(
											[]ServicePlanSummary{
												{
													ServicePlan: ServicePlan{GUID: "service-plan-guid-1", Name: "service-plan-1", Public: true},
												},
												{
													ServicePlan: ServicePlan{GUID: "service-plan-guid-2", Name: "service-plan-2", Public: true},
												},
											},
										))
										Expect(summaries[1].Services[0].Plans).To(ConsistOf(
											[]ServicePlanSummary{
												{
													ServicePlan: ServicePlan{GUID: "service-plan-guid-3", Name: "service-plan-3", Public: true},
												},
												{
													ServicePlan: ServicePlan{GUID: "service-plan-guid-4", Name: "service-plan-4", Public: true},
												},
											},
										))

										Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(4))
										Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(ccv2.Filter{
											Type:     constant.ServiceGUIDFilter,
											Operator: constant.EqualOperator,
											Values:   []string{"service-guid-1"},
										}))
										Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(1)).To(ConsistOf(ccv2.Filter{
											Type:     constant.ServiceGUIDFilter,
											Operator: constant.EqualOperator,
											Values:   []string{"service-guid-2"},
										}))

										Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(0))
									})
								})

								When("there are non-public services plans", func() {
									BeforeEach(func() {
										// Gets the service plans
										fakeCloudControllerClient.GetServicePlansReturnsOnCall(0,
											[]ccv2.ServicePlan{
												{GUID: "service-plan-guid-1", Name: "service-plan-1"},
												{GUID: "service-plan-guid-2", Name: "service-plan-2"},
											},
											ccv2.Warnings{"service-plan-warning-1"},
											nil)
									})

									When("fetching orgs and service plan visibilities is successful", func() {
										BeforeEach(func() {
											// Gets the visibilities for the plans
											fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0,
												[]ccv2.ServicePlanVisibility{
													{OrganizationGUID: "org-guid-1"},
													{OrganizationGUID: "org-guid-2"},
												},
												ccv2.Warnings{"service-plan-visibility-1"},
												nil)
											fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(1,
												[]ccv2.ServicePlanVisibility{
													{OrganizationGUID: "org-guid-3"},
													{OrganizationGUID: "org-guid-4"},
												},
												ccv2.Warnings{"service-plan-visibility-2"},
												nil)

											// Gets the OrgNames for the visibilities
											fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
												ccv2.Organization{Name: "org-1"},
												ccv2.Warnings{"org-warning-1"},
												nil)
											fakeCloudControllerClient.GetOrganizationReturnsOnCall(1,
												ccv2.Organization{Name: "org-2"},
												ccv2.Warnings{"org-warning-2"},
												nil)
											fakeCloudControllerClient.GetOrganizationReturnsOnCall(2,
												ccv2.Organization{Name: "org-3"},
												ccv2.Warnings{"org-warning-3"},
												nil)
											fakeCloudControllerClient.GetOrganizationReturnsOnCall(3,
												ccv2.Organization{Name: "org-4"},
												ccv2.Warnings{"org-warning-4"},
												nil)
										})

										It("returns the expected Service Plans", func() {
											Expect(executeErr).ToNot(HaveOccurred())
											Expect(warnings).To(ConsistOf(
												"get-brokers-warning",
												"service-warning-1", "service-warning-2",
												"service-plan-warning-1",
												"service-plan-visibility-1", "service-plan-visibility-2",
												"org-warning-1", "org-warning-2", "org-warning-3", "org-warning-4",
											))

											Expect(summaries[0].Services[0].Plans[0].VisibleTo).To(ConsistOf("org-1", "org-2"))
											Expect(summaries[0].Services[0].Plans[1].VisibleTo).To(ConsistOf("org-3", "org-4"))

											Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(4))
											Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(ccv2.Filter{
												Type:     constant.ServiceGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"service-guid-1"},
											}))
											Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(1)).To(ConsistOf(ccv2.Filter{
												Type:     constant.ServiceGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"service-guid-2"},
											}))

											Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(2))
											Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)).To(ConsistOf(ccv2.Filter{
												Type:     constant.ServicePlanGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"service-plan-guid-1"},
											}))
											Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(1)).To(ConsistOf(ccv2.Filter{
												Type:     constant.ServicePlanGUIDFilter,
												Operator: constant.EqualOperator,
												Values:   []string{"service-plan-guid-2"},
											}))
										})
									})

									When("fetching the service plan visibilities fails", func() {
										BeforeEach(func() {
											// Gets the visibilities for the plans
											fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0,
												nil,
												ccv2.Warnings{"service-plan-visibility-1"},
												errors.New("boom"))
										})

										It("returns the error and warnings", func() {
											Expect(executeErr).To(MatchError("boom"))
											Expect(warnings).To(ConsistOf(
												"get-brokers-warning",
												"service-warning-1",
												"service-plan-warning-1",
												"service-plan-visibility-1",
											))
										})
									})

									When("fetching the organizations fails", func() {
										BeforeEach(func() {
											fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0,
												[]ccv2.ServicePlanVisibility{
													{OrganizationGUID: "org-guid-1"},
													{OrganizationGUID: "org-guid-2"},
												},
												ccv2.Warnings{"service-plan-visibility-1"},
												nil)

											fakeCloudControllerClient.GetOrganizationReturnsOnCall(0,
												ccv2.Organization{},
												ccv2.Warnings{"org-warning-1"},
												errors.New("boom"))
										})

										It("returns the error and warnings", func() {
											Expect(executeErr).To(MatchError("boom"))
											Expect(warnings).To(ConsistOf(
												"get-brokers-warning",
												"service-warning-1",
												"service-plan-warning-1",
												"service-plan-visibility-1",
												"org-warning-1",
											))
										})
									})
								})
							})
						})

						When("fetching service plans fails", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.GetServicePlansReturnsOnCall(0,
									nil,
									ccv2.Warnings{"service-plan-warning-1"},
									errors.New("boom"))
							})

							It("returns the error and warnings", func() {
								Expect(executeErr).To(MatchError("boom"))
								Expect(warnings).To(ConsistOf("get-brokers-warning", "service-warning-1", "service-plan-warning-1"))
							})
						})
					})

					When("fetching services fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetServicesReturnsOnCall(0,
								nil,
								ccv2.Warnings{"service-warning-1"},
								errors.New("boom"))
						})

						It("returns the error and warnings", func() {
							Expect(executeErr).To(MatchError("boom"))
							Expect(warnings).To(ConsistOf("get-brokers-warning", "service-warning-1"))
						})
					})
				})
			})

			When("fetching service brokers fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturns(nil, ccv2.Warnings{"test-warning"}, errors.New("explode"))
				})

				It("returns the warnings and propagates the error", func() {
					Expect(warnings).To(ConsistOf("test-warning"))
					Expect(executeErr).To(MatchError("explode"))
				})
			})
		})

		When("only a service broker is specified", func() {
			BeforeEach(func() {
				broker = "broker-1"
				service = ""
				organization = ""

				fakeCloudControllerClient.GetServiceBrokersReturns(
					[]ccv2.ServiceBroker{
						{Name: "broker-1", GUID: "broker-guid-1"},
					},
					ccv2.Warnings{"get-brokers-warning"},
					nil)
			})

			It("returns expected Service Brokers", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-brokers-warning"))
				Expect(summaries).To(ConsistOf(
					ServiceBrokerSummary{
						ServiceBroker: ServiceBroker{
							Name: "broker-1",
							GUID: "broker-guid-1",
						},
					}))

				Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)).To(ConsistOf(ccv2.Filter{
					Type:     constant.NameFilter,
					Operator: constant.EqualOperator,
					Values:   []string{broker},
				}))
			})
		})
	})
})
