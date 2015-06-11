package actors_test

import (
	"github.com/cloudfoundry/cli/cf/errors"

	"github.com/cloudfoundry/cli/cf/actors"
	fake_plan_builder "github.com/cloudfoundry/cli/cf/actors/plan_builder/fakes"
	fake_service_builder "github.com/cloudfoundry/cli/cf/actors/service_builder/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	fake_orgs "github.com/cloudfoundry/cli/cf/api/organizations/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plans", func() {
	var (
		actor actors.ServicePlanActor

		servicePlanRepo           *testapi.FakeServicePlanRepo
		servicePlanVisibilityRepo *testapi.FakeServicePlanVisibilityRepository
		orgRepo                   *fake_orgs.FakeOrganizationRepository

		planBuilder    *fake_plan_builder.FakePlanBuilder
		serviceBuilder *fake_service_builder.FakeServiceBuilder

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
		servicePlanRepo = &testapi.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &testapi.FakeServicePlanVisibilityRepository{}
		orgRepo = &fake_orgs.FakeOrganizationRepository{}
		planBuilder = &fake_plan_builder.FakePlanBuilder{}
		serviceBuilder = &fake_service_builder.FakeServiceBuilder{}

		actor = actors.NewServicePlanHandler(servicePlanRepo, servicePlanVisibilityRepo, orgRepo, planBuilder, serviceBuilder)

		org1 = models.Organization{}
		org1.Name = "org-1"
		org1.Guid = "org-1-guid"

		org2 = models.Organization{}
		org2.Name = "org-2"
		org2.Guid = "org-2-guid"

		orgRepo.FindByNameReturns(org1, nil)

		publicServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			Guid:            "public-service-plan-visibility-guid",
			ServicePlanGuid: "public-service-plan-guid",
		}

		privateServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			Guid:            "private-service-plan-visibility-guid",
			ServicePlanGuid: "private-service-plan-guid",
		}

		limitedServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
			Guid:             "limited-service-plan-visibility-guid",
			ServicePlanGuid:  "limited-service-plan-guid",
			OrganizationGuid: "org-1-guid",
		}

		publicServicePlan = models.ServicePlanFields{
			Name:   "public-service-plan",
			Guid:   "public-service-plan-guid",
			Public: true,
		}

		privateServicePlan = models.ServicePlanFields{
			Name:     "private-service-plan",
			Guid:     "private-service-plan-guid",
			Public:   false,
			OrgNames: []string{},
		}

		limitedServicePlan = models.ServicePlanFields{
			Name:   "limited-service-plan",
			Guid:   "limited-service-plan-guid",
			Public: false,
			OrgNames: []string{
				"org-1",
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
				limitedServicePlan,
			},
		}

		privateService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-private-service",
				Guid:  "my-private-service-guid",
			},
			Plans: []models.ServicePlanFields{
				privateServicePlan,
				privateServicePlan,
			},
		}
		publicAndLimitedService = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-public-and-limited-service",
				Guid:  "my-public-and-limited-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				publicServicePlan,
				limitedServicePlan,
			},
		}

		visibility1 = models.ServicePlanVisibilityFields{
			Guid:             "visibility-guid-1",
			OrganizationGuid: "org-1-guid",
			ServicePlanGuid:  "limited-service-plan-guid",
		}
	})

	Describe(".UpdateAllPlansForService", func() {
		BeforeEach(func() {
			servicePlanVisibilityRepo.SearchReturns(
				[]models.ServicePlanVisibilityFields{privateServicePlanVisibilityFields}, nil)

			servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
				"my-mixed-service-guid": {
					publicServicePlan,
					privateServicePlan,
				},
			}
		})

		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))
			_, err := actor.UpdateAllPlansForService("not-a-service", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Removes the service plan visibilities for any non-public service plans", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
			_, err := actor.UpdateAllPlansForService("my-mixed-service", true)
			Expect(err).ToNot(HaveOccurred())

			servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
			Expect(servicePlanVisibilityGuid).To(Equal("private-service-plan-visibility-guid"))
		})

		Context("when setting all plans to public", func() {
			It("Sets all non-public service plans to public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				_, err := actor.UpdateAllPlansForService("my-mixed-service", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeFalse())
				Expect(serviceGuid).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeTrue())
			})

			It("Returns true if all the plans were public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(publicService, nil)

				servicesOriginallyPublic, err := actor.UpdateAllPlansForService("my-public-service", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesOriginallyPublic).To(BeTrue())
			})

			It("Returns false if any of the plans were not public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

				servicesOriginallyPublic, err := actor.UpdateAllPlansForService("my-mixed-service", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(servicesOriginallyPublic).To(BeFalse())
			})

			It("Does not try to update service plans if they are all already public", func() {
				servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
					"my-public-service-guid": {
						publicServicePlan,
						publicServicePlan,
					},
				}

				_, err := actor.UpdateAllPlansForService("my-public-service", true)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})

		Context("when setting all plans to private", func() {
			It("Sets all public service plans to private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

				_, err := actor.UpdateAllPlansForService("my-mixed-service", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
				Expect(servicePlan.Public).To(BeTrue())
				Expect(serviceGuid).To(Equal("my-mixed-service-guid"))
				Expect(public).To(BeFalse())
			})

			It("Returns true if all plans were already private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(privateService, nil)

				allPlansAlreadyPrivate, err := actor.UpdateAllPlansForService("my-private-service", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(allPlansAlreadyPrivate).To(BeTrue())
			})

			It("Returns false if any of the plans were not private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)

				allPlansAlreadyPrivate, err := actor.UpdateAllPlansForService("my-mixed-service", false)
				Expect(err).NotTo(HaveOccurred())
				Expect(allPlansAlreadyPrivate).To(BeFalse())
			})

			It("Does not try to update service plans if they are all already private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(privateService, nil)

				_, err := actor.UpdateAllPlansForService("my-private-service", false)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanRepo.UpdateCallCount()).To(Equal(0))
			})
		})
	})

	Describe(".UpdateOrgForService", func() {
		BeforeEach(func() {
			serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)

			orgRepo.FindByNameReturns(org1, nil)
		})

		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameForOrgReturns(models.ServiceOffering{}, errors.New("service was not found"))

			_, err := actor.UpdateOrgForService("not-a-service", "org-1", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		Context("when giving access to all plans for a single org", func() {
			It("creates a service plan visibility for all private plans", func() {
				_, err := actor.UpdateOrgForService("my-mixed-service", "org-1", true)
				Expect(err).ToNot(HaveOccurred())

				Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(1))

				planGuid, orgGuid := servicePlanVisibilityRepo.CreateArgsForCall(0)
				Expect(planGuid).To(Equal("private-service-plan-guid"))
				Expect(orgGuid).To(Equal("org-1-guid"))
			})

			It("Returns true if all the plans were already public", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(publicService, nil)
				allPlansSet, err := actor.UpdateOrgForService("my-public-service", "org-1", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(allPlansSet).To(BeTrue())
			})

			It("Returns false if any of the plans were not public", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(privateService, nil)
				allPlansSet, err := actor.UpdateOrgForService("my-private-service", "org-1", true)
				Expect(err).NotTo(HaveOccurred())
				Expect(allPlansSet).To(BeFalse())
			})

			It("Does not try to update service plans if they are all already public or the org already has access", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(publicAndLimitedService, nil)

				allPlansWereSet, err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(0))
				Expect(allPlansWereSet).To(BeTrue())
			})
		})

		Context("when disabling access to all plans for a single org", func() {
			It("deletes the associated visibilities for all limited plans", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(publicAndLimitedService, nil)
				servicePlanVisibilityRepo.SearchReturns([]models.ServicePlanVisibilityFields{visibility1}, nil)
				allPlansSet, err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(1))
				Expect(allPlansSet).To(BeFalse())

				services := servicePlanVisibilityRepo.SearchArgsForCall(0)
				Expect(services["organization_guid"]).To(Equal("org-1-guid"))

				visibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(visibilityGuid).To(Equal("visibility-guid-1"))
			})

			It("Does not try to update service plans if they are all public", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(publicService, nil)

				allPlansWereSet, err := actor.UpdateOrgForService("my-public-and-limited-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
				Expect(allPlansWereSet).To(BeTrue())
			})

			It("Does not try to update service plans if the org already did not have visibility", func() {
				serviceBuilder.GetServiceByNameForOrgReturns(privateService, nil)

				allPlansWereSet, err := actor.UpdateOrgForService("my-private-service", "org-1", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
				Expect(allPlansWereSet).To(BeTrue())
			})
		})
	})

	Describe(".UpdateSinglePlanForService", func() {
		It("Returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(models.ServiceOffering{}, errors.New("service was not found"))
			_, err := actor.UpdateSinglePlanForService("not-a-service", "public-service-plan", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("Returns None if the original plan was private", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(privateService, nil)
			originalAccessValue, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(originalAccessValue).To(Equal(actors.None))
		})

		It("Returns All if the original plan was public", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
			originalAccessValue, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(originalAccessValue).To(Equal(actors.All))
		})

		It("Returns an error if the plan cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
			_, err := actor.UpdateSinglePlanForService("my-mixed-service", "not-a-service-plan", true)
			Expect(err.Error()).To(Equal("The plan not-a-service-plan could not be found for service my-mixed-service"))
		})

		Context("when setting a public service plan to public", func() {
			It("Does not try to update the service plan", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", true)
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
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", true)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGuid).To(Equal("private-service-plan-visibility-guid"))
			})

			It("sets a service plan to public", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
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
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "private-service-plan", false)
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
				_, err := actor.UpdateSinglePlanForService("my-mixed-service", "public-service-plan", false)
				Expect(err).ToNot(HaveOccurred())

				servicePlanVisibilityGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
				Expect(servicePlanVisibilityGuid).To(Equal("public-service-plan-visibility-guid"))
			})

			It("sets the plan to private", func() {
				serviceBuilder.GetServiceByNameWithPlansReturns(mixedService, nil)
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
			orgRepo.FindByNameReturns(org1, nil)
		})

		It("returns an error if the service cannot be found", func() {
			serviceBuilder.GetServiceByNameForOrgReturns(models.ServiceOffering{}, errors.New("service was not found"))

			_, err := actor.UpdatePlanAndOrgForService("not-a-service", "public-service-plan", "public-org", true)
			Expect(err.Error()).To(Equal("service was not found"))
		})

		It("returns an error if the org cannot be found", func() {
			orgRepo.FindByNameReturns(models.Organization{}, errors.NewModelNotFoundError("organization", "not-an-org"))
			_, err := actor.UpdatePlanAndOrgForService("a-real-service", "public-service-plan", "not-an-org", true)
			Expect(err).To(HaveOccurred())
		})

		It("returns an error if the plan cannot be found", func() {
			serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)

			_, err := actor.UpdatePlanAndOrgForService("a-real-service", "not-a-plan", "org-1", true)
			Expect(err).To(HaveOccurred())
		})

		Context("when disabling access to a single plan for a single org", func() {
			Context("for a public plan", func() {
				It("returns All", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", false)

					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.All))
				})

				It("does not try and delete the visibility", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", false)

					Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.All))
				})
			})

			Context("for a private plan", func() {
				Context("with no service plan visibilities", func() {
					It("returns None", func() {
						serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", false)

						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.None))
					})
					It("does not try and delete the visibility", func() {
						serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", false)

						Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.None))
					})
				})

				Context("with service plan visibilities", func() {
					BeforeEach(func() {
						servicePlanVisibilityRepo.SearchReturns(
							[]models.ServicePlanVisibilityFields{limitedServicePlanVisibilityFields}, nil)

					})
					It("deletes a service plan visibility", func() {
						serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-1", false)

						servicePlanVisGuid := servicePlanVisibilityRepo.DeleteArgsForCall(0)
						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.Limited))
						Expect(servicePlanVisGuid).To(Equal("limited-service-plan-visibility-guid"))
					})

					It("does not call delete if the specified service plan visibility does not exist", func() {
						serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
						orgRepo.FindByNameReturns(org2, nil)
						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-2", false)

						Expect(servicePlanVisibilityRepo.DeleteCallCount()).To(Equal(0))
						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.Limited))
					})
				})
			})
		})

		Context("when enabling access", func() {
			Context("for a public plan", func() {
				It("returns All", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", true)

					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.All))
				})

				It("does not try and create the visibility", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "public-service-plan", "org-1", true)

					Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(0))
					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.All))
				})
			})

			Context("for a limited plan", func() {
				BeforeEach(func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
				})
				It("returns Limited", func() {
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-1", true)

					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.Limited))
				})

				Context("when the org already has access", func() {
					It("does not try and create the visibility", func() {
						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-1", true)

						Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(0))
						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.Limited))
					})
				})
				Context("when the org does not have access", func() {
					It("creates the visibility", func() {
						orgRepo.FindByNameReturns(org2, nil)

						originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "limited-service-plan", "org-2", true)

						Expect(servicePlanVisibilityRepo.CreateCallCount()).To(Equal(1))
						Expect(err).NotTo(HaveOccurred())
						Expect(originalAccessValue).To(Equal(actors.Limited))
					})
				})
			})

			Context("for a private plan", func() {
				It("returns None", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", true)

					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.None))
				})

				It("creates a service plan visibility", func() {
					serviceBuilder.GetServiceByNameForOrgReturns(mixedService, nil)
					originalAccessValue, err := actor.UpdatePlanAndOrgForService("my-mixed-service", "private-service-plan", "org-1", true)

					servicePlanGuid, orgGuid := servicePlanVisibilityRepo.CreateArgsForCall(0)
					Expect(err).NotTo(HaveOccurred())
					Expect(originalAccessValue).To(Equal(actors.None))
					Expect(servicePlanGuid).To(Equal("private-service-plan-guid"))
					Expect(orgGuid).To(Equal("org-1-guid"))
				})
			})
		})
	})
})
