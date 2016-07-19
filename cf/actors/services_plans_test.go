package actors_test

import (
	"code.cloudfoundry.org/cli/cf/errors"

	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/actors/planbuilder/planbuilderfakes"
	"code.cloudfoundry.org/cli/cf/actors/servicebuilder/servicebuilderfakes"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plans", func() {
	var (
		actor actors.ServicePlanActor

		servicePlanRepo           *apifakes.OldFakeServicePlanRepo
		servicePlanVisibilityRepo *apifakes.FakeServicePlanVisibilityRepository
		orgRepo                   *organizationsfakes.FakeOrganizationRepository

		planBuilder    *planbuilderfakes.FakePlanBuilder
		serviceBuilder *servicebuilderfakes.FakeServiceBuilder

		privateServicePlanVisibilityFields models.ServicePlanVisibilityFields
		publicServicePlanVisibilityFields  models.ServicePlanVisibilityFields
		limitedServicePlanVisibilityFields models.ServicePlanVisibilityFields

		publicServicePlan  models.ServicePlanFields
		privateServicePlan models.ServicePlanFields
		limitedServicePlan models.ServicePlanFields

		publicService           models.ServiceOffering
		mixedService            models.ServiceOffering
		privateService          models.ServiceOffering
		publicAndLimitedService models.ServiceOffering

		org1 models.Organization
		org2 models.Organization

		visibility1 models.ServicePlanVisibilityFields
	)

	BeforeEach(func() {
		servicePlanRepo = new(apifakes.OldFakeServicePlanRepo)
		servicePlanVisibilityRepo = new(apifakes.FakeServicePlanVisibilityRepository)
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		planBuilder = new(planbuilderfakes.FakePlanBuilder)
		serviceBuilder = new(servicebuilderfakes.FakeServiceBuilder)

		actor = actors.NewServicePlanHandler(servicePlanRepo, servicePlanVisibilityRepo, orgRepo, planBuilder, serviceBuilder)

		org1 = models.Organization{}
		org1.Name = "org-1"
		org1.GUID = "org-1-guid"

		org2 = models.Organization{}
		org2.Name = "org-2"
		org2.GUID = "org-2-guid"

		orgRepo.FindByNameReturns(org1, nil)

		publicServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			GUID:            "public-service-plan-visibility-guid",
			ServicePlanGUID: "public-service-plan-guid",
		}

		privateServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			GUID:            "private-service-plan-visibility-guid",
			ServicePlanGUID: "private-service-plan-guid",
		}

		limitedServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			GUID:             "limited-service-plan-visibility-guid",
			ServicePlanGUID:  "limited-service-plan-guid",
			OrganizationGUID: "org-1-guid",
		}

		publicServicePlan = models.ServicePlanFields{
			Name:   "public-service-plan",
			GUID:   "public-service-plan-guid",
			Public: true,
		}

		privateServicePlan = models.ServicePlanFields{
			Name:   "private-service-plan",
			GUID:   "private-service-plan-guid",
			Public: false,
		}

		limitedServicePlan = models.ServicePlanFields{
			Name:   "limited-service-plan",
			GUID:   "limited-service-plan-guid",
			Public: false,
		}

		publicService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-public-service",
				GUID:  "my-public-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				publicServicePlan,
			},
		}

		mixedService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-mixed-service",
				GUID:  "my-mixed-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
				limitedServicePlan,
			},
		}

		privateService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-private-service",
				GUID:  "my-private-service-guid",
			},
			Plans: []models.ServicePlanFields{
				privateServicePlan,
				privateServicePlan,
			},
		}
		publicAndLimitedService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-public-and-limited-service",
				GUID:  "my-public-and-limited-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				publicServicePlan,
				limitedServicePlan,
			},
		}

		visibility1 = models.ServicePlanVisibilityFields{
			GUID:             "visibility-guid-1",
			OrganizationGUID: "org-1-guid",
			ServicePlanGUID:  "limited-service-plan-guid",
		}
	})

	Describe(".UpdateAllPlansForService", func() {
		BeforeEach(func() {
			servicePlanVisibilityRepo.SearchReturns(
				[]models.ServicePlanVisibilityFields{privateServicePlanVisibilityFields}, nil,
			)

			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-mixed-service-guid": {
					publicServicePlan,
					privateServicePlan,
				},
			}
		})

		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))
			err := actor.UpdateAllPlansForService("not-a-service", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Removes the service plan visibilities for any non-public service plans", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
			err := actor.UpdateAllPlansForService("my-mixed-service", true)
			Expect(err).ToNot(HaveOccurred())

			servicePlanVisibilityGUID := servicePlanVisibilityRepo.DeleteArgsForCall(0)
			Expect(servicePlanVisibilityGUID).To(Equal("private-service-plan-visibility-guid"))
		})

		Context("when setting all plans to public", func() {
			It("Sets all non-public service plans to public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateAllPlansForService("my-mixed-service", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGUID, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeFalse())
				Expect(serviceGUID).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeTrue())
			})

			It("Does not try to update service plans if they are all already public", func() {
				servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
					"my-public-service-guid": {
						publicServicePlan,
						publicServicePlan,
					},
				}

				err := actor.UpdateAllPlansForService("my-public-service", true)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("when setting all plans to private", func() {
			It("Sets all public service plans to private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

				err := actor.UpdateAllPlansForService("my-mixed-service", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGUID, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeTrue())
				Expect(serviceGUID).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeFalse())
			})

			It("Does not try to update service plans if they are all already private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(privateService, nil)

				err := actor.UpdateAllPlansForService("my-private-service", false)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})
	})

	Describe(".UpdateOrgForService", func() {
		BeforeEach(func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

			orgRepo.FindByNameReturns(org1, nil)
		})

		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))

			err := actor.UpdateOrgForService("not-a-service", "org-1", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		Context("when giving access to all plans for a single org", func() {
			It("creates a service plan visibility for all plans", func() {
				err := actor.UpdateOrgForService("my-mixed-service", "org-1", true)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(2))

				planGUID, orgGUID := servicePlanVisibilityRepo.CreateArgsForCall(0)
				Expect(planGUID).To(Equal("private-service-plan-guid"))
				Expect(orgGUID).To(Equal("org-1-guid"))

				planGUID, orgGUID = servicePlanVisibilityRepo.CreateArgsForCall(1)
				Expect(planGUID).To(Equal("limited-service-plan-guid"))
				Expect(orgGUID).To(Equal("org-1-guid"))
			})

			It("Does not try to update service plans if they are all already public or the org already has access", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(publicAndLimitedService, nil)

				err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(1))
			})
		})

		Context("when disabling access to all plans for a single org", func() {
			It("deletes the associated visibilities for all limited plans", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(publicAndLimitedService, nil)
				servicePlanVisibilityRepo.SearchReturns([]models.ServicePlanVisibilityFields{visibility1}, nil)
				err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(1))

				services := servicePlanVisibilityRepo.SearchArgsForCall(0)
				Expect(services["organization_guid"]).To(Equal("org-1-guid"))

				visibilityGUID := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(visibilityGUID).To(Equal("visibility-guid-1"))
			})

			It("Does not try to update service plans if they are all public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(publicService, nil)

				err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
			})

			It("Does not try to update service plans if the org already did not have visibility", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(privateService, nil)

				err := actor.UpdateOrgForService("my-private-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
			})
		})
	})

	Describe(".UpdateSinglePlanForService", func() {
		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))
			err := actor.UpdateSinglePlanForService("not-a-service", "public-service-plan", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Returns an error if the plan cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
			err := actor.UpdateSinglePlanForService("my-mixed-service", "not-a-service-plan", true)
			Expect(err.Error()).To(Equal("The plan not-a-service-plan could not be found for service my-mixed-service"))
		})

		Context("when setting a public service plan to public", func() {
			It("Does not try to update the service plan", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("when setting private service plan to public", func() {
			BeforeEach(func() {
				servicePlanVisibilityRepo.SearchReturns(
					[]models.ServicePlanVisibilityFields{privateServicePlanVisibilityFields}, nil)
			})

			It("removes the service plan visibilities for the service plan", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGUID := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGUID).To(Equal("private-service-plan-visibility-guid"))
			})

			It("sets a service plan to public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGUID, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeFalse())
				Expect(serviceGUID).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeTrue())
			})
		})

		Context("when setting a private service plan to private", func() {
			It("Does not try to update the service plan", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("When setting public service plan to private", func() {
			BeforeEach(func() {
				servicePlanVisibilityRepo.SearchReturns(
					[]models.ServicePlanVisibilityFields{publicServicePlanVisibilityFields}, nil)
			})

			It("removes the service plan visibilities for the service plan", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGUID := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGUID).To(Equal("public-service-plan-visibility-guid"))
			})

			It("sets the plan to private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGUID, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeTrue())
				Expect(serviceGUID).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeFalse())
			})
		})
	})

	Describe(".UpdatePlanAndOrgForService", func() {
		BeforeEach(func() {
			orgRepo.FindByNameReturns(org1, nil)
		})

		It("returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))

			err := actor.UpdatePlanAndOrgForService("not-a-service", "public-service-plan", "public-org", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("returns an error if the org cannot be found", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.NewModelNotFoundError("organization", "not-an-org"))
			err := actor.UpdatePlanAndOrgForService("a-real-service", "public-service-plan", "not-an-org", true)
			Expect(err).To(HaveOccurred())
		})

		It("returns an error if the plan cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

			err := actor.UpdatePlanAndOrgForService("a-real-service", "not-a-plan", "org-1", true)
			Expect(err).To(HaveOccurred())
		})

		Context("when disabling access to a single plan for a single org", func() {
			Context("for a public plan", func() {
				It("does not try and delete the visibility", func() {
					serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
					err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", false)
					Expect(err).NotTo(HaveOccurred())

					Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
				})
			})

			Context("for a private plan", func() {
				Context("with no service plan visibilities", func() {
					It("does not try and delete the visibility", func() {
						serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
						err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", false)

						Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("with service plan visibilities", func() {
					BeforeEach(func() {
						servicePlanVisibilityRepo.SearchReturns(
							[]models.ServicePlanVisibilityFields{limitedServicePlanVisibilityFields}, nil)

					})
					It("deletes a service plan visibility", func() {
						serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
						err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-1", false)

						servicePlanVisGUID := servicePlanVisibilityRepo.DeleteArgsForCall(0)
						Expect(err).NotTo(HaveOccurred())
						Expect(servicePlanVisGUID).To(Equal("limited-service-plan-visibility-guid"))
					})

					It("does not call delete on the non-existant service plan visibility", func() {
						serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
						orgRepo.FindByNameReturns(org2, nil)
						servicePlanVisibilityRepo.SearchReturns(nil, nil)
						err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-2", false)
						Expect(err).NotTo(HaveOccurred())

						Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
					})
				})
			})
		})

		Context("when enabling access", func() {
			Context("for a public plan", func() {
				It("does not try and create the visibility", func() {
					serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
					err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", true)

					Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(0))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("for a private plan", func() {
				It("returns None", func() {
					serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
					err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", true)

					Expect(err).NotTo(HaveOccurred())
				})

				It("creates a service plan visibility", func() {
					serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
					err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", true)

					servicePlanGUID, orgGUID := servicePlanVisibilityRepo.CreateArgsForCall(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(servicePlanGUID).To(Equal("private-service-plan-guid"))
					Expect(orgGUID).To(Equal("org-1-guid"))
				})
			})
		})
	})
})
