package v2action_test

import (
	"errors"
	"reflect"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Summary Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("GetServicesSummaries", func() {
		var (
			servicesSummaries []ServiceSummary
			warnings          Warnings
			err               error
		)

		JustBeforeEach(func() {
			servicesSummaries, warnings, err = actor.GetServicesSummaries()
		})

		When("there are no services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, nil)
			})

			It("returns an empty list of summaries and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(HaveLen(0))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("fetching services returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, errors.New("oops"))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError("oops"))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("there are services with plans", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:        "service-a-guid",
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						GUID:        "service-b-guid",
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a", ServiceGUID: "service-a-guid", Public: true},
					{Name: "plan-b", ServiceGUID: "service-b-guid", Public: true},
					{Name: "plan-c", ServiceGUID: "service-b-guid", Public: true},
					{Name: "plan-d", ServiceGUID: "service-a-guid", Public: true},
				}

				fakeCloudControllerClient.GetServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("returns summaries including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							GUID:        "service-a-guid",
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-a",
									Public:      true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-d",
									Public:      true,
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							GUID:        "service-b-guid",
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									Name:        "plan-b",
									Public:      true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									Name:        "plan-c",
									Public:      true,
								},
							},
						},
					},
				))

				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
			})

			Context("and fetching plans returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, ccv2.Warnings{"get-plans-warning"}, errors.New("plan-oops"))
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError("plan-oops"))
					Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
				})
			})
		})

		AfterEach(func() {
			Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(0))
		})
	})

	Describe("GetServicesSummariesForSpace", func() {
		var (
			servicesSummaries []ServiceSummary
			warnings          Warnings
			err               error
			spaceGUID         = "space-123"
			organizationGUID  = "org-guid-123"
		)

		JustBeforeEach(func() {
			servicesSummaries, warnings, err = actor.GetServicesSummariesForSpace(spaceGUID, organizationGUID)
		})

		When("there are no services", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, nil)
			})

			It("returns an empty list of summaries and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(HaveLen(0))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("fetching services returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, errors.New("oops"))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError("oops"))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		It("retrieves the services for the correct space", func() {
			requestedSpaceGUID, _ := fakeCloudControllerClient.GetSpaceServicesArgsForCall(0)
			Expect(requestedSpaceGUID).To(Equal(spaceGUID))
		})

		Context("and fetching plans returns an error", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						Label:       "service-a",
						Description: "service-a-description",
					},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, ccv2.Warnings{"get-plans-warning"}, errors.New("plan-oops"))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError("plan-oops"))
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
			})
		})

		When("there are services with public plans", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:        "service-a-guid",
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						GUID:        "service-b-guid",
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a", ServiceGUID: "service-b-guid", Public: true},
					{Name: "plan-b", ServiceGUID: "service-a-guid", Public: true},
					{Name: "plan-c", ServiceGUID: "service-a-guid", Public: true},
					{Name: "plan-d", ServiceGUID: "service-b-guid", Public: true},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("returns summaries including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							GUID:        "service-a-guid",
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-b",
									Public:      true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-c",
									Public:      true,
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							GUID:        "service-b-guid",
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									Name:        "plan-a",
									Public:      true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									Name:        "plan-d",
									Public:      true,
								},
							},
						},
					},
				))

				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
			})

			It("uses the IN filter to get all the plans for all services and then match them up", func() {
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.ServiceGUIDFilter,
						Operator: constant.InOperator,
						Values:   []string{"service-a-guid", "service-b-guid"},
					},
				))
			})

			It("does not request service plan visibilities", func() {
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(0))
			})
		})

		When("there are services with one non-public plan", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:        "service-a-guid",
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						GUID:        "service-b-guid",
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a", ServiceGUID: "service-a-guid", Public: true},
					{Name: "plan-b", ServiceGUID: "service-a-guid", Public: false},
					{Name: "plan-c", ServiceGUID: "service-b-guid", Public: true},
					{Name: "plan-d", ServiceGUID: "service-b-guid", Public: false},
				}

				brokers := []ccv2.ServiceBroker{
					{Name: "normal-broker"},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
				fakeCloudControllerClient.GetServiceBrokersReturns(brokers, ccv2.Warnings{"get-brokers-warning"}, nil)
			})

			It("returns summaries excluding non public plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							GUID:        "service-a-guid",
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-a",
									Public:      true,
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							GUID:        "service-b-guid",
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									Name:        "plan-c",
									Public:      true,
								},
							},
						},
					},
				))

				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-brokers-warning", "get-brokers-warning"))
			})
		})

		When("there are services with non-public plan but visible to the org", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:        "service-a-guid",
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						GUID:        "service-b-guid",
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{GUID: "plan-a-guid", ServiceGUID: "service-a-guid", Name: "plan-a", Public: false},
					{GUID: "plan-b-guid", ServiceGUID: "service-a-guid", Name: "plan-b", Public: false},
					{GUID: "plan-c-guid", ServiceGUID: "service-b-guid", Name: "plan-c", Public: false},
					{GUID: "plan-d-guid", ServiceGUID: "service-b-guid", Name: "plan-d", Public: false},
				}

				visibilities1 := []ccv2.ServicePlanVisibility{
					{OrganizationGUID: "org-guid-1", ServicePlanGUID: "plan-a-guid"},
					{OrganizationGUID: "org-guid-1", ServicePlanGUID: "plan-b-guid"},
				}

				visibilities2 := []ccv2.ServicePlanVisibility{
					{OrganizationGUID: "org-guid-1", ServicePlanGUID: "plan-c-guid"},
				}

				brokers := []ccv2.ServiceBroker{
					{Name: "normal-broker"},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
				fakeCloudControllerClient.GetServiceBrokersReturns(brokers, ccv2.Warnings{"get-brokers-warning"}, nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0, visibilities1, ccv2.Warnings{"get-visibilities-a-warning"}, nil)
				fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(1, visibilities2, ccv2.Warnings{"get-visibilities-b-warning"}, nil)
			})

			It("returns summaries with plans visible for the org", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							GUID:        "service-a-guid",
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									GUID:        "plan-a-guid",
									ServiceGUID: "service-a-guid",
									Name:        "plan-a",
									Public:      false,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									GUID:        "plan-b-guid",
									Name:        "plan-b",
									Public:      false,
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							GUID:        "service-b-guid",
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-b-guid",
									GUID:        "plan-c-guid",
									Name:        "plan-c",
									Public:      false,
								},
							},
						},
					},
				))
			})

			It("returns all warnings", func() {
				Expect(warnings).To(ConsistOf(
					"get-services-warning",
					"get-plans-warning",
					"get-brokers-warning",
					"get-visibilities-a-warning",
					"get-brokers-warning",
					"get-visibilities-b-warning",
				))
			})

			It("gets service plans using IN filter for all services at once", func() {
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.ServiceGUIDFilter,
						Operator: constant.InOperator,
						Values:   []string{"service-a-guid", "service-b-guid"},
					},
				))
			})

			It("gets plan visibilities for the non-public plan for the org", func() {
				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(2))

				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.ServicePlanGUIDFilter,
						Operator: constant.InOperator,
						Values:   []string{"plan-a-guid", "plan-b-guid"},
					},
					ccv2.Filter{
						Type:     constant.OrganizationGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{organizationGUID},
					},
				))

				Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(1)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.ServicePlanGUIDFilter,
						Operator: constant.InOperator,
						Values:   []string{"plan-c-guid", "plan-d-guid"},
					},
					ccv2.Filter{
						Type:     constant.OrganizationGUIDFilter,
						Operator: constant.EqualOperator,
						Values:   []string{organizationGUID},
					},
				))
			})

			When("getting visibilities fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlanVisibilitiesReturnsOnCall(0, []ccv2.ServicePlanVisibility{}, ccv2.Warnings{"get-visibilities-warning"}, errors.New("oopsie"))
				})

				It("returns errors and warnings", func() {
					Expect(err).To(MatchError(errors.New("oopsie")))
					Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-brokers-warning", "get-visibilities-warning"))
				})
			})
		})

		When("when there are space scoped services", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						GUID:              "service-a-guid",
						Label:             "service-a",
						Description:       "service-a-description",
						ServiceBrokerName: "broker-a",
					},
					{
						GUID:              "service-b-guid",
						Label:             "service-b",
						Description:       "service-b-description",
						ServiceBrokerName: "broker-b",
					},
				}

				plans := []ccv2.ServicePlan{
					{GUID: "plan-a-guid", ServiceGUID: "service-a-guid", Name: "plan-a", Public: false},
					{GUID: "plan-b-guid", ServiceGUID: "service-a-guid", Name: "plan-b", Public: false},
					{GUID: "plan-c-guid", ServiceGUID: "service-b-guid", Name: "plan-c", Public: false},
					{GUID: "plan-d-guid", ServiceGUID: "service-b-guid", Name: "plan-d", Public: false},
				}

				brokersA := []ccv2.ServiceBroker{
					{GUID: "broker-a-guid", Name: "broker-a", SpaceGUID: spaceGUID},
				}

				brokersB := []ccv2.ServiceBroker{
					{GUID: "broker-b-guid", Name: "broker-b", SpaceGUID: "different-space-guid"},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
				fakeCloudControllerClient.GetServiceBrokersReturnsOnCall(0, brokersA, ccv2.Warnings{"get-brokers-a-warning"}, nil)
				fakeCloudControllerClient.GetServiceBrokersReturnsOnCall(1, brokersB, ccv2.Warnings{"get-brokers-b-warning"}, nil)
			})

			It("returns summaries including plans only for brokers scoped to the current space", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							GUID:              "service-a-guid",
							Label:             "service-a",
							Description:       "service-a-description",
							ServiceBrokerName: "broker-a",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									GUID:        "plan-a-guid",
									ServiceGUID: "service-a-guid",
									Name:        "plan-a",
									Public:      false,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									GUID:        "plan-b-guid",
									ServiceGUID: "service-a-guid",
									Name:        "plan-b",
									Public:      false,
								},
							},
						},
					},
				))
			})

			It("gets all plans at once using IN operator for service GUIDs", func() {
				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.ServiceGUIDFilter,
						Operator: constant.InOperator,
						Values:   []string{"service-a-guid", "service-b-guid"},
					},
				))
			})

			It("applies filter by broker name when fetching list of brokers", func() {
				Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(0)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"broker-a"},
					},
				))

				Expect(fakeCloudControllerClient.GetServiceBrokersArgsForCall(1)).To(ConsistOf(
					ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"broker-b"},
					},
				))
			})

			It("returns warnings", func() {
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-brokers-a-warning", "get-brokers-b-warning"))
			})

			When("getting broker fails with an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBrokersReturnsOnCall(0, []ccv2.ServiceBroker{}, ccv2.Warnings{"get-brokers-warning"}, errors.New("oopsie"))
				})

				It("returns error and warnings", func() {
					Expect(err).To(MatchError(errors.New("oopsie")))
					Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-brokers-warning"))
				})
			})
		})

	})

	Describe("GetServiceSummaryByName", func() {
		var (
			serviceName    string
			serviceSummary ServiceSummary
			warnings       Warnings
			err            error
		)

		BeforeEach(func() {
			serviceName = "service"
		})

		JustBeforeEach(func() {
			serviceSummary, warnings, err = actor.GetServiceSummaryByName(serviceName)
		})

		When("there is no service matching the provided name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, nil)
			})

			It("returns an error that the service with the provided name is missing", func() {
				Expect(err).To(MatchError(actionerror.ServiceNotFoundError{Name: serviceName}))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("fetching a service returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, errors.New("oops"))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError("oops"))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("the service exists", func() {
			var services []ccv2.Service

			BeforeEach(func() {
				services = []ccv2.Service{
					{
						GUID:        "service-a-guid",
						Label:       "service",
						Description: "service-description",
					},
				}
				plans := []ccv2.ServicePlan{
					{Name: "plan-a", ServiceGUID: "service-a-guid", Public: true},
					{Name: "plan-b", ServiceGUID: "service-a-guid", Public: true},
				}

				fakeCloudControllerClient.GetServicesStub = func(filters ...ccv2.Filter) ([]ccv2.Service, ccv2.Warnings, error) {
					filterToMatch := ccv2.Filter{
						Type:     constant.LabelFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"service"},
					}

					if len(filters) == 1 && reflect.DeepEqual(filters[0], filterToMatch) {
						return services, ccv2.Warnings{"get-services-warning"}, nil
					}

					return []ccv2.Service{}, nil, nil
				}

				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("filters for the service, returning a service summary including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceSummary).To(Equal(
					ServiceSummary{
						Service: Service{
							GUID:        "service-a-guid",
							Label:       "service",
							Description: "service-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-a",
									Public:      true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									ServiceGUID: "service-a-guid",
									Name:        "plan-b",
									Public:      true,
								},
							},
						},
					}))

				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
			})

			Context("and fetching plans returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
					fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, ccv2.Warnings{"get-plans-warning"}, errors.New("plan-oops"))
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError("plan-oops"))
					Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
				})
			})
		})

		AfterEach(func() {
			Expect(fakeCloudControllerClient.GetServiceBrokersCallCount()).To(Equal(0))
		})
	})

	Describe("GetServiceSummaryForSpaceByName", func() {
		var (
			serviceName      = "service"
			spaceGUID        = "space-123"
			organizationGUID = "org-guid-123"
			serviceSummary   ServiceSummary
			warnings         Warnings
			err              error
		)

		JustBeforeEach(func() {
			serviceSummary, warnings, err = actor.GetServiceSummaryForSpaceByName(spaceGUID, serviceName, organizationGUID)
		})

		When("there is no service matching the provided name", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, nil)
			})

			It("returns an error that the service with the provided name is missing", func() {
				Expect(err).To(MatchError(actionerror.ServiceNotFoundError{Name: serviceName}))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("fetching a service returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceServicesReturns([]ccv2.Service{}, ccv2.Warnings{"get-services-warning"}, errors.New("oops"))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError("oops"))
				Expect(warnings).To(ConsistOf("get-services-warning"))
			})
		})

		When("the service exists with all plans public", func() {
			var services []ccv2.Service

			BeforeEach(func() {
				services = []ccv2.Service{
					{
						Label:       "service",
						Description: "service-description",
					},
				}
				plans := []ccv2.ServicePlan{
					{Name: "plan-a", Public: true},
					{Name: "plan-b", Public: true},
				}

				fakeCloudControllerClient.GetSpaceServicesStub = func(guid string, filters ...ccv2.Filter) ([]ccv2.Service, ccv2.Warnings, error) {
					filterToMatch := ccv2.Filter{
						Type:     constant.LabelFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"service"},
					}

					if len(filters) == 1 && reflect.DeepEqual(filters[0], filterToMatch) && spaceGUID == guid {
						return services, ccv2.Warnings{"get-services-warning"}, nil
					}

					return []ccv2.Service{}, nil, nil
				}

				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("filters for the service within the space, returning a service summary including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceSummary).To(Equal(
					ServiceSummary{
						Service: Service{
							Label:       "service",
							Description: "service-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name:   "plan-a",
									Public: true,
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name:   "plan-b",
									Public: true,
								},
							},
						},
					}))
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
			})

			Context("and fetching plans returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
					fakeCloudControllerClient.GetServicePlansReturns([]ccv2.ServicePlan{}, ccv2.Warnings{"get-plans-warning"}, errors.New("plan-oops"))
				})

				It("returns the error and all warnings", func() {
					Expect(err).To(MatchError("plan-oops"))
					Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning"))
				})
			})
		})

		When("the service exists with one non-public plan", func() {
			var services []ccv2.Service

			BeforeEach(func() {
				services = []ccv2.Service{
					{
						Label:       "service",
						Description: "service-description",
					},
				}
				plans := []ccv2.ServicePlan{
					{GUID: "plan-a-guid", Name: "plan-a", Public: true},
					{GUID: "plan-b-guid", Name: "plan-b", Public: false},
				}

				brokers := []ccv2.ServiceBroker{
					{Name: "normal-broker"},
				}

				fakeCloudControllerClient.GetSpaceServicesStub = func(guid string, filters ...ccv2.Filter) ([]ccv2.Service, ccv2.Warnings, error) {
					filterToMatch := ccv2.Filter{
						Type:     constant.LabelFilter,
						Operator: constant.EqualOperator,
						Values:   []string{"service"},
					}

					if len(filters) == 1 && reflect.DeepEqual(filters[0], filterToMatch) && spaceGUID == guid {
						return services, ccv2.Warnings{"get-services-warning"}, nil
					}

					return []ccv2.Service{}, nil, nil
				}

				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
				fakeCloudControllerClient.GetServiceBrokersReturns(brokers, ccv2.Warnings{"get-brokers-warning"}, nil)
			})

			It("returns service summary excluding non public plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceSummary).To(Equal(
					ServiceSummary{
						Service: Service{
							Label:       "service",
							Description: "service-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									GUID:   "plan-a-guid",
									Name:   "plan-a",
									Public: true,
								},
							},
						},
					}))
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-brokers-warning"))
			})

			When("the service broker is space-scoped", func() {
				BeforeEach(func() {
					services[0].ServiceBrokerName = "broker-a"

					brokers := []ccv2.ServiceBroker{
						{GUID: "broker-a-guid", Name: "broker-a", SpaceGUID: spaceGUID},
					}

					fakeCloudControllerClient.GetServiceBrokersReturns(brokers, ccv2.Warnings{"get-brokers-warning"}, nil)
				})

				It("returns summaries with all plans related to this space-scoped broker", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(serviceSummary).To(Equal(
						ServiceSummary{
							Service: Service{
								Label:             "service",
								Description:       "service-description",
								ServiceBrokerName: "broker-a",
							},
							Plans: []ServicePlanSummary{
								ServicePlanSummary{
									ServicePlan: ServicePlan{
										GUID:   "plan-a-guid",
										Name:   "plan-a",
										Public: true,
									},
								},
								ServicePlanSummary{
									ServicePlan: ServicePlan{
										GUID:   "plan-b-guid",
										Name:   "plan-b",
										Public: false,
									},
								},
							},
						}))
				})

				It("returns all warnings", func() {
					Expect(warnings).To(ConsistOf(
						"get-services-warning",
						"get-plans-warning",
						"get-brokers-warning",
					))
				})
			})

			When("the non-public plan is visible to the org", func() {
				BeforeEach(func() {
					visibilities := []ccv2.ServicePlanVisibility{
						{OrganizationGUID: "org-guid-1", ServicePlanGUID: "plan-b-guid"},
					}

					fakeCloudControllerClient.GetServicePlanVisibilitiesReturns(visibilities, ccv2.Warnings{"get-visibilities-warning"}, nil)
				})

				It("returns summaries with plans visible for the org", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(serviceSummary).To(Equal(
						ServiceSummary{
							Service: Service{
								Label:       "service",
								Description: "service-description",
							},
							Plans: []ServicePlanSummary{
								ServicePlanSummary{
									ServicePlan: ServicePlan{
										GUID:   "plan-a-guid",
										Name:   "plan-a",
										Public: true,
									},
								},
								ServicePlanSummary{
									ServicePlan: ServicePlan{
										GUID:   "plan-b-guid",
										Name:   "plan-b",
										Public: false,
									},
								},
							},
						}))
				})

				It("returns all warnings", func() {
					Expect(warnings).To(ConsistOf(
						"get-services-warning",
						"get-plans-warning",
						"get-brokers-warning",
						"get-visibilities-warning",
					))
				})

				It("gets plan visibilities for the non-public plan for the org", func() {
					Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesCallCount()).To(Equal(1))

					Expect(fakeCloudControllerClient.GetServicePlanVisibilitiesArgsForCall(0)).To(ConsistOf(
						ccv2.Filter{
							Type:     constant.ServicePlanGUIDFilter,
							Operator: constant.InOperator,
							Values:   []string{"plan-b-guid"},
						},
						ccv2.Filter{
							Type:     constant.OrganizationGUIDFilter,
							Operator: constant.EqualOperator,
							Values:   []string{organizationGUID},
						},
					))
				})
			})
		})
	})
})
