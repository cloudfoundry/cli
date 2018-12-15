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
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a"},
					{Name: "plan-b"},
				}

				fakeCloudControllerClient.GetServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("returns summaries including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
								},
							},
						},
					},
				))
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-plans-warning"))
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
	})

	Describe("GetServicesSummariesForSpace", func() {
		var (
			servicesSummaries []ServiceSummary
			warnings          Warnings
			err               error
			spaceGUID         = "space-123"
		)

		JustBeforeEach(func() {
			servicesSummaries, warnings, err = actor.GetServicesSummariesForSpace(spaceGUID)
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

		When("there are services with plans", func() {
			BeforeEach(func() {
				services := []ccv2.Service{
					{
						Label:       "service-a",
						Description: "service-a-description",
					},
					{
						Label:       "service-b",
						Description: "service-b-description",
					},
				}

				plans := []ccv2.ServicePlan{
					{Name: "plan-a"},
					{Name: "plan-b"},
				}

				fakeCloudControllerClient.GetSpaceServicesReturns(services, ccv2.Warnings{"get-services-warning"}, nil)
				fakeCloudControllerClient.GetServicePlansReturns(plans, ccv2.Warnings{"get-plans-warning"}, nil)
			})

			It("returns summaries including plans and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesSummaries).To(ConsistOf(
					ServiceSummary{
						Service: Service{
							Label:       "service-a",
							Description: "service-a-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
								},
							},
						},
					},
					ServiceSummary{
						Service: Service{
							Label:       "service-b",
							Description: "service-b-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
								},
							},
						},
					},
				))
				Expect(warnings).To(ConsistOf("get-services-warning", "get-plans-warning", "get-plans-warning"))
			})

			It("retrieves the services for the correct space", func() {
				requestedSpaceGUID, _ := fakeCloudControllerClient.GetSpaceServicesArgsForCall(0)
				Expect(requestedSpaceGUID).To(Equal(spaceGUID))
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
						Label:       "service",
						Description: "service-description",
					},
				}
				plans := []ccv2.ServicePlan{
					{Name: "plan-a"},
					{Name: "plan-b"},
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
							Label:       "service",
							Description: "service-description",
						},
						Plans: []ServicePlanSummary{
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
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
	})

	Describe("GetServiceSummaryForSpaceByName", func() {
		var (
			serviceName    string
			spaceGUID      string
			serviceSummary ServiceSummary
			warnings       Warnings
			err            error
		)

		BeforeEach(func() {
			serviceName = "service"
			spaceGUID = "space-123"
		})

		JustBeforeEach(func() {
			serviceSummary, warnings, err = actor.GetServiceSummaryForSpaceByName(spaceGUID, serviceName)
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

		When("the service exists", func() {
			var services []ccv2.Service

			BeforeEach(func() {
				services = []ccv2.Service{
					{
						Label:       "service",
						Description: "service-description",
					},
				}
				plans := []ccv2.ServicePlan{
					{Name: "plan-a"},
					{Name: "plan-b"},
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
									Name: "plan-a",
								},
							},
							ServicePlanSummary{
								ServicePlan: ServicePlan{
									Name: "plan-b",
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
	})
})
