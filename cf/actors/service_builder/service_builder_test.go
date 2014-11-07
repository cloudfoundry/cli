package service_builder_test

import (
	plan_builder_fakes "github.com/cloudfoundry/cli/cf/actors/plan_builder/fakes"
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"

	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Builder", func() {
	var (
		planBuilder     *plan_builder_fakes.FakePlanBuilder
		serviceBuilder  service_builder.ServiceBuilder
		serviceRepo     *testapi.FakeServiceRepo
		service1        models.ServiceOffering
		v1Service       models.ServiceOffering
		planWithoutOrgs models.ServicePlanFields
		plan1           models.ServicePlanFields
		plan2           models.ServicePlanFields
	)

	BeforeEach(func() {
		serviceRepo = &testapi.FakeServiceRepo{}
		planBuilder = &plan_builder_fakes.FakePlanBuilder{}

		serviceBuilder = service_builder.NewBuilder(serviceRepo, planBuilder)
		service1 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-service1",
				Guid:       "service-guid1",
				BrokerGuid: "my-service-broker-guid1",
			},
		}

		v1Service = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "v1Service",
				Guid:       "v1Service-guid",
				BrokerGuid: "my-service-broker-guid1",
				Provider:   "IAmV1",
			},
		}

		serviceRepo.FindServiceOfferingsByLabelName = "my-service1"
		serviceRepo.FindServiceOfferingsByLabelServiceOfferings = models.ServiceOfferings{service1, v1Service}

		serviceRepo.GetServiceOfferingByGuidReturns = struct {
			ServiceOffering models.ServiceOffering
			Error           error
		}{
			service1,
			nil,
		}

		serviceRepo.ListServicesFromBrokerReturns = map[string][]models.ServiceOffering{
			"my-service-broker-guid1": []models.ServiceOffering{service1},
		}

		plan1 = models.ServicePlanFields{
			Name:                "service-plan1",
			Guid:                "service-plan1-guid",
			ServiceOfferingGuid: "service-guid1",
			OrgNames:            []string{"org1", "org2"},
		}

		plan2 = models.ServicePlanFields{
			Name:                "service-plan2",
			Guid:                "service-plan2-guid",
			ServiceOfferingGuid: "service-guid1",
		}

		planWithoutOrgs = models.ServicePlanFields{
			Name:                "service-plan-without-orgs",
			Guid:                "service-plan-without-orgs-guid",
			ServiceOfferingGuid: "service-guid1",
		}

		planBuilder.GetPlansVisibleToOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
		planBuilder.GetPlansForServiceWithOrgsReturns([]models.ServicePlanFields{plan1, plan2}, nil)
		planBuilder.GetPlansForServiceForOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
	})

	Describe(".GetServicesForSpace", func() {
		BeforeEach(func() {
			serviceRepo.GetServiceOfferingsForSpaceReturns = struct {
				ServiceOfferings []models.ServiceOffering
				Error            error
			}{
				[]models.ServiceOffering{service1, service1},
				nil,
			}
		})

		It("returns the services for the space", func() {
			services, err := serviceBuilder.GetServicesForSpace("spaceGuid")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
		})
	})

	Describe(".GetServicesForSpaceWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.GetServiceOfferingsForSpaceReturns = struct {
				ServiceOfferings []models.ServiceOffering
				Error            error
			}{
				[]models.ServiceOffering{service1, service1},
				nil,
			}

			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the services for the space, populated with plans", func() {
			services, err := serviceBuilder.GetServicesForSpaceWithPlans("spaceGuid")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Plans[0]).To(Equal(planWithoutOrgs))
			Expect(services[1].Plans[0]).To(Equal(planWithoutOrgs))
		})
	})

	Describe(".GetAllServices", func() {
		BeforeEach(func() {
			serviceRepo.GetAllServiceOfferingsReturns = struct {
				ServiceOfferings []models.ServiceOffering
				Error            error
			}{
				[]models.ServiceOffering{service1, service1},
				nil,
			}
		})

		It("returns the named service, populated with plans", func() {
			services, err := serviceBuilder.GetAllServices()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
		})
	})

	Describe(".GetAllServicesWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.GetAllServiceOfferingsReturns = struct {
				ServiceOfferings []models.ServiceOffering
				Error            error
			}{
				[]models.ServiceOffering{service1, service1},
				nil,
			}

			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{plan1}, nil)
		})

		It("returns the named service, populated with plans", func() {
			services, err := serviceBuilder.GetAllServicesWithPlans()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Plans[0]).To(Equal(plan1))
		})
	})

	Describe(".GetServiceByNameWithPlans", func() {
		BeforeEach(func() {
			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{plan2}, nil)
		})

		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameWithPlans("my-service1")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServiceByNameWithPlansWithOrgNames", func() {
		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameWithPlansWithOrgNames("my-service1")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".GetServiceByNameForSpace", func() {
		Context("mixed v2 and v1 services", func() {
			BeforeEach(func() {
				service2 := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label: "service",
						Guid:  "service-guid-v2",
					},
				}

				service1 := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label:    "service",
						Guid:     "service-guid",
						Provider: "a provider",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns = struct {
					ServiceOfferings models.ServiceOfferings
					Error            error
				}{
					models.ServiceOfferings{service1, service2},
					nil,
				}
			})

			It("returns the nv2 service", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGuid")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(service.Plans)).To(Equal(0))
				Expect(service.Guid).To(Equal("service-guid-v2"))
			})
		})

		Context("v2 services", func() {
			BeforeEach(func() {
				service := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label: "service",
						Guid:  "service-guid",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns = struct {
					ServiceOfferings models.ServiceOfferings
					Error            error
				}{
					models.ServiceOfferings{service},
					nil,
				}
			})

			It("returns the named service", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGuid")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(service.Plans)).To(Equal(0))
				Expect(service.Guid).To(Equal("service-guid"))
			})
		})

		Context("v1 services", func() {
			BeforeEach(func() {
				service := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label:    "service",
						Guid:     "service-guid",
						Provider: "a provider",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns = struct {
					ServiceOfferings models.ServiceOfferings
					Error            error
				}{
					models.ServiceOfferings{service},
					nil,
				}
			})

			It("returns the an error", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGuid")
				Expect(service).To(Equal(models.ServiceOffering{}))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe(".GetServiceByNameForSpaceWithPlans", func() {
		BeforeEach(func() {
			service := models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label: "serviceWithPlans",
				},
			}

			serviceRepo.FindServiceOfferingsForSpaceByLabelReturns = struct {
				ServiceOfferings models.ServiceOfferings
				Error            error
			}{
				models.ServiceOfferings{service},
				nil,
			}

			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the named service", func() {
			service, err := serviceBuilder.GetServiceByNameForSpaceWithPlans("serviceWithPlans", "spaceGuid")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(service.Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServicesByNameForSpaceWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.FindServiceOfferingsForSpaceByLabelReturns = struct {
				ServiceOfferings models.ServiceOfferings
				Error            error
			}{
				models.ServiceOfferings{service1, v1Service},
				nil,
			}

			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the named service", func() {
			services, err := serviceBuilder.GetServicesByNameForSpaceWithPlans("serviceWithPlans", "spaceGuid")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Label).To(Equal("my-service1"))
			Expect(services[0].Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(services[0].Plans[0].OrgNames).To(BeNil())
			Expect(services[1].Label).To(Equal("v1Service"))
			Expect(services[1].Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(services[1].Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServiceByNameForOrg", func() {
		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameForOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			Expect(planBuilder.GetPlansForServiceForOrgCallCount()).To(Equal(1))
			servName, orgName := planBuilder.GetPlansForServiceForOrgArgsForCall(0)
			Expect(servName).To(Equal("service-guid1"))
			Expect(orgName).To(Equal("org1"))

			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".GetServicesForBroker", func() {
		It("returns all the services for a broker, fully populated", func() {
			services, err := serviceBuilder.GetServicesForBroker("my-service-broker-guid1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".GetServiceVisibleToOrg", func() {
		It("Returns a service populated with plans visible to the provided org", func() {
			service, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				planBuilder.GetPlansVisibleToOrgReturns(nil, nil)
				service, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(service).To(Equal(models.ServiceOffering{}))
			})
		})
	})

	Describe(".GetServicesVisibleToOrg", func() {
		It("Returns services with plans visible to the provided org", func() {
			planBuilder.GetPlansVisibleToOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
			services, err := serviceBuilder.GetServicesVisibleToOrg("org1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				planBuilder.GetPlansVisibleToOrgReturns(nil, nil)
				services, err := serviceBuilder.GetServicesVisibleToOrg("org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(services).To(BeNil())
			})
		})
	})
})
