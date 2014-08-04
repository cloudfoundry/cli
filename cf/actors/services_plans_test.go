package actors_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plans", func() {
	var (
		actor                     actors.ServicePlanActor
		serviceRepo               *fakes.FakeServiceRepo
		servicePlanRepo           *fakes.FakeServicePlanRepo
		servicePlanVisibilityRepo *fakes.FakeServicePlanVisibilityRepository
		orgRepo                   *fakes.FakeOrgRepository

		privateServicePlanVisibilityFields models.ServicePlanVisibilityFields
		publicServicePlanVisibilityFields  models.ServicePlanVisibilityFields

		publicServicePlan  models.ServicePlanFields
		privateServicePlan models.ServicePlanFields

		publicService models.ServiceOffering
		mixedService  models.ServiceOffering
	)

	BeforeEach(func() {
		serviceRepo = &fakes.FakeServiceRepo{}
		servicePlanRepo = &fakes.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		actor = actors.NewServicePlanHandler(serviceRepo, servicePlanRepo, servicePlanVisibilityRepo, orgRepo)

		publicServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			Guid:            "public-service-plan-visibility-guid",
			ServicePlanGuid: "public-service-plan-guid",
		}

		privateServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			Guid:            "private-service-plan-visibility-guid",
			ServicePlanGuid: "private-service-plan-guid",
		}

		publicServicePlan = models.ServicePlanFields{
			Name:   "public-service-plan",
			Guid:   "public-service-plan-guid",
			Public: true,
		}

		privateServicePlan = models.ServicePlanFields{
			Name:   "private-service-plan",
			Guid:   "private-service-plan-guid",
			Public: false,
			OrgNames: []string{
				"org-1",
				"org-2",
			},
		}

		publicService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-public-service",
				Guid:  "my-public-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				publicServicePlan,
			},
		}

		mixedService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-mixed-service",
				Guid:  "my-mixed-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
			},
		}
	})

	Describe(".UpdateAllPlansForService", func() {
		BeforeEach(func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = mixedService

			servicePlanVisibilityRepo.ListReturns(
				[]models.ServicePlanVisibilityFields{privateServicePlanVisibilityFields}, nil)

			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-mixed-service-guid": {
					publicServicePlan,
					privateServicePlan,
				},
			}
		})

		It("Returns an error if the service cannot be found", func() {
			serviceRepo.FindServiceOfferingByLabelApiResponse = errors.New("service was not found")

			_, err := actor.UpdateAllPlansForService("not-a-service")
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Sets all non-public service plans to public", func() {
			_, err := actor.UpdateAllPlansForService("my-mixed-service")
			Expect(err).ToNot(HaveOccurred())

			servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
			Expect(servicePlan.Public).To(BeFalse())
			Expect(serviceGuid).To(Equal("my-mixed-service-guid"))
			Expect(public).To(BeTrue())
		})

		It("Returns true if all the plans were public", func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = publicService
			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-public-service-guid": {
					publicServicePlan,
					publicServicePlan,
				},
			}

			servicesOriginallyPublic, err := actor.UpdateAllPlansForService("my-public-service")
			Expect(err).NotTo(HaveOccurred())
			Expect(servicesOriginallyPublic).To(BeTrue())
		})

		It("Returns false if any of the plans were not public", func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = mixedService
			servicesOriginallyPublic, err := actor.UpdateAllPlansForService("my-mixed-service")
			Expect(err).NotTo(HaveOccurred())
			Expect(servicesOriginallyPublic).To(BeFalse())
		})

		It("Does not try to update service plans if they are all already public", func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = publicService
			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-public-service-guid": {
					publicServicePlan,
					publicServicePlan,
				},
			}

			_, err := actor.UpdateAllPlansForService("my-public-service")
			Expect(err).ToNot(HaveOccurred())

			Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
		})

		It("Removes the service plan visibilities for any non-public service plans", func() {
			_, err := actor.UpdateAllPlansForService("my-mixed-service")
			Expect(err).ToNot(HaveOccurred())

			servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
			Expect(servicePlanVisibilityGuid).To(Equal("private-service-plan-visibility-guid"))
		})

	})

	Describe(".UpdateSinglePlanForService", func() {
		BeforeEach(func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = mixedService

			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-mixed-service-guid": {
					publicServicePlan,
					privateServicePlan,
				},
			}
		})

		It("Returns an error if the service cannot be found", func() {
			serviceRepo.FindServiceOfferingByLabelApiResponse = errors.New("service was not found")

			_, err := actor.UpdateSinglePlanForService("not-a-service", "public-service-plan", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Returns false if the original plan was private", func() {
			serviceOriginallyPublic, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceOriginallyPublic).To(BeFalse())
		})

		It("Returns true if the original plan was public", func() {
			serviceOriginallyPublic, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(serviceOriginallyPublic).To(BeTrue())
		})

		It("Returns an error if the plan cannot be found", func() {
			_, err := actor.UpdateSinglePlanForService("my-mixed-service", "not-a-service-plan", true)
			Expect(err.Error()).To(Equal("The plan not-a-service-plan could not be found for service my-mixed-service"))
		})

		Context("when setting a public service plan to public", func() {
			It("Does not try to update the service plan", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("when setting private service plan to public", func() {
			BeforeEach(func() {
				servicePlanVisibilityRepo.ListReturns(
					[]models.ServicePlanVisibilityFields{privateServicePlanVisibilityFields}, nil)
			})

			It("removes the service plan visibilities for the service plan", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGuid).To(Equal("private-service-plan-visibility-guid"))
			})

			It("sets a service plan to public", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeFalse())
				Expect(serviceGuid).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeTrue())
			})
		})

		Context("when setting a private service plan to private", func() {
			It("Does not try to update the service plan", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("When setting public service plan to private", func() {
			BeforeEach(func() {
				servicePlanVisibilityRepo.ListReturns(
					[]models.ServicePlanVisibilityFields{publicServicePlanVisibilityFields}, nil)
			})

			It("removes the service plan visibilities for the service plan", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGuid).To(Equal("public-service-plan-visibility-guid"))
			})

			It("sets the plan to private", func() {
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeTrue())
				Expect(serviceGuid).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeFalse())
			})
		})
	})

	Describe(".UpdatePlanAndOrgForService", func() {
		BeforeEach(func() {
			serviceRepo.FindServiceOfferingByLabelServiceOffering = mixedService

			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-mixed-service-guid": {
					publicServicePlan,
					privateServicePlan,
				},
			}
		})
	})
})
