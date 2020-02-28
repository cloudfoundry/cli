package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("service access actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("GetServiceAccess", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetServicePlansReturns(fakeServicePlans(), ccv3.Warnings{"plans warning"}, nil)
			fakeCloudControllerClient.GetServiceOfferingsReturns(fakeServiceOfferings(), ccv3.Warnings{"offerings warning"}, nil)

			visibility1 := ccv3.ServicePlanVisibility{
				Organizations: []ccv3.VisibilityDetail{{Name: "org-3"}},
			}
			visibility2 := ccv3.ServicePlanVisibility{
				Organizations: []ccv3.VisibilityDetail{{Name: "org-1"}, {Name: "org-2"}},
			}
			fakeCloudControllerClient.GetServicePlanVisibilityReturnsOnCall(0, visibility1, ccv3.Warnings{"visibility1 1 warning"}, nil)
			fakeCloudControllerClient.GetServicePlanVisibilityReturnsOnCall(1, visibility2, ccv3.Warnings{"visibility1 2 warning"}, nil)
		})

		It("produces a slice of ServicePlanAccess objects", func() {
			access, warnings, err := actor.GetServiceAccess("", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("plans warning", "offerings warning", "visibility1 1 warning", "visibility1 2 warning"))
			Expect(access).To(Equal([]ServicePlanAccess{
				{
					BrokerName:          "land-broker",
					ServiceOfferingName: "yellow",
					ServicePlanName:     "orange",
					VisibilityType:      "organization",
					VisibilityDetails:   []string{"org-1", "org-2"},
				},
				{
					BrokerName:          "land-broker",
					ServiceOfferingName: "yellow",
					ServicePlanName:     "yellow",
					VisibilityType:      "organization",
					VisibilityDetails:   []string{"org-3"},
				},
				{
					BrokerName:          "sea-broker",
					ServiceOfferingName: "magenta",
					ServicePlanName:     "red",
					VisibilityType:      "public",
					VisibilityDetails:   nil,
				},
				{
					BrokerName:          "sea-broker",
					ServiceOfferingName: "magenta",
					ServicePlanName:     "violet",
					VisibilityType:      "public",
					VisibilityDetails:   nil,
				},
				{
					BrokerName:          "sky-broker",
					ServiceOfferingName: "cyan",
					ServicePlanName:     "blue",
					VisibilityType:      "admin",
					VisibilityDetails:   nil,
				},
				{
					BrokerName:          "sky-broker",
					ServiceOfferingName: "cyan",
					ServicePlanName:     "green",
					VisibilityType:      "space",
					VisibilityDetails:   nil,
				},
				{
					BrokerName:          "sky-broker",
					ServiceOfferingName: "key",
					ServicePlanName:     "indigo",
					VisibilityType:      "space",
					VisibilityDetails:   nil,
				},
			}))
		})

		When("there are no service offerings", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{}, ccv3.Warnings{"offerings warning"}, nil)
			})

			It("returns an empty slice", func() {
				result, warnings, err := actor.GetServiceAccess("", "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElement("offerings warning"))
				Expect(result).To(BeEmpty())
			})
		})

		When("filtering on organization", func() {
			const (
				guid    = "fake-org-guid"
				name    = "fake-org-name"
				warning = "fake get org warning"
			)

			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{{GUID: guid}}, ccv3.Warnings{warning}, nil)
			})

			It("passes the organization in the plan filter", func() {
				_, warnings, err := actor.GetServiceAccess("", "", name)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElement(warning))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{name},
				}))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.OrganizationGUIDFilter,
					Values: []string{guid},
				}))
			})

			When("the organization is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{},
						ccv3.Warnings{"org warning"},
						nil,
					)
				})

				It("returns an error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess("", "", "fake-org")
					Expect(warnings).To(ContainElement("org warning"))
					Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{Name: "fake-org"}))
				})
			})
		})

		When("filtering on service offering", func() {
			const (
				guid    = "fake-service-offering-guid"
				name    = "fake-service-offering-name"
				warning = "fake get service offering warning"
			)

			BeforeEach(func() {
				fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{{GUID: guid}}, ccv3.Warnings{warning}, nil)
			})

			It("passes the service offering in the filters", func() {
				_, warnings, err := actor.GetServiceAccess("", name, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElement(warning))

				Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{name},
				}))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.ServiceOfferingNamesFilter,
					Values: []string{name},
				}))
			})

			When("the service offering is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{}, ccv3.Warnings{warning}, nil)
				})

				It("returns an error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess("", name, "")
					Expect(err).To(MatchError(actionerror.ServiceNotFoundError{Name: name}))
					Expect(warnings).To(ContainElement(warning))
				})
			})
		})

		When("filtering on service broker", func() {
			const name = "fake-service-broker-name"

			It("passes the service broker in the filters", func() {
				_, _, err := actor.GetServiceAccess(name, "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.ServiceBrokerNamesFilter,
					Values: []string{name},
				}))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.ServiceBrokerNamesFilter,
					Values: []string{name},
				}))
			})

			When("the service broker filter returns no service offerings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{}, ccv3.Warnings{"warning"}, nil)
				})

				It("returns an error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess(name, "", "")
					Expect(err).To(MatchError(actionerror.ServiceNotFoundError{Broker: name}))
					Expect(warnings).To(ContainElement("warning"))
				})
			})
		})

		When("combining filters", func() {
			const (
				orgGUID         = "fake-org-guid"
				orgName         = "fake-org-name"
				orgWarning      = "fake get org warning"
				offeringGUID    = "fake-service-offering-guid"
				offeringName    = "fake-service-offering-name"
				offeringWarning = "fake get service offering warning"
				brokerName      = "fake-service-broker-name"
			)

			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{{GUID: orgGUID}}, ccv3.Warnings{orgWarning}, nil)
				fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{{GUID: offeringGUID}}, ccv3.Warnings{offeringWarning}, nil)
			})

			It("passes all the filters", func() {
				_, warnings, err := actor.GetServiceAccess(brokerName, offeringName, orgName)
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElements(orgWarning, offeringWarning))

				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationsArgsForCall(0)).To(ContainElement(ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{orgName},
				}))

				Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ContainElements(
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{offeringName},
					},
					ccv3.Query{
						Key:    ccv3.ServiceBrokerNamesFilter,
						Values: []string{brokerName},
					},
				))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ContainElements(
					ccv3.Query{
						Key:    ccv3.OrganizationGUIDFilter,
						Values: []string{orgGUID},
					},
					ccv3.Query{
						Key:    ccv3.ServiceOfferingNamesFilter,
						Values: []string{offeringName},
					},
					ccv3.Query{
						Key:    ccv3.ServiceBrokerNamesFilter,
						Values: []string{brokerName},
					},
				))
			})
		})

		When("the client fails to get resources", func() {
			Context("service plans", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						nil,
						ccv3.Warnings{"plans warning"},
						errors.New("fake plans error"),
					)
				})

				It("returns the error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess("", "", "")
					Expect(warnings).To(ContainElement("plans warning"))
					Expect(err).To(MatchError("fake plans error"))
				})
			})

			Context("service offerings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns(
						nil,
						ccv3.Warnings{"offerings warning"},
						errors.New("fake offerings error"),
					)
				})

				It("returns the error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess("", "", "")
					Expect(warnings).To(ContainElement("offerings warning"))
					Expect(err).To(MatchError("fake offerings error"))
				})
			})

			Context("service plan visibility", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlanVisibilityReturnsOnCall(
						0,
						ccv3.ServicePlanVisibility{},
						ccv3.Warnings{"visibility warning"},
						errors.New("fake visibility error"),
					)
				})

				It("returns the error and warnings", func() {
					_, warnings, err := actor.GetServiceAccess("", "", "")
					Expect(warnings).To(ContainElement("visibility warning"))
					Expect(err).To(MatchError("fake visibility error"))
				})
			})
		})
	})

	Describe("EnableServiceAccess", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]ccv3.Organization{{GUID: "org-guid"}},
				ccv3.Warnings{"org warning"},
				nil,
			)

			fakeCloudControllerClient.GetServiceOfferingsReturns(
				[]ccv3.ServiceOffering{{GUID: "fake-offering-guid"}},
				ccv3.Warnings{"some warning"},
				nil,
			)

			fakeCloudControllerClient.GetServicePlansReturns(
				[]ccv3.ServicePlan{
					{GUID: "fake-plan-guid-1"},
					{GUID: "fake-plan-guid-2"},
				},
				ccv3.Warnings{"other warning"},
				nil,
			)

			fakeCloudControllerClient.UpdateServicePlanVisibilityReturns(
				ccv3.ServicePlanVisibility{},
				ccv3.Warnings{"post warning"},
				nil,
			)
		})

		It("sets visibility to public", func() {
			skipped, warnings, err := actor.EnableServiceAccess("fake-offering", "", "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("some warning", "other warning", "post warning", "post warning"))
			Expect(skipped).To(BeEmpty())

			Expect(fakeCloudControllerClient.UpdateServicePlanVisibilityCallCount()).To(Equal(2))

			planGUID, actualVisibility := fakeCloudControllerClient.UpdateServicePlanVisibilityArgsForCall(0)
			Expect(planGUID).To(Equal("fake-plan-guid-1"))
			Expect(actualVisibility).To(Equal(ccv3.ServicePlanVisibility{Type: "public"}))

			planGUID, actualVisibility = fakeCloudControllerClient.UpdateServicePlanVisibilityArgsForCall(1)
			Expect(planGUID).To(Equal("fake-plan-guid-2"))
			Expect(actualVisibility).To(Equal(ccv3.ServicePlanVisibility{Type: "public"}))
		})

		Describe("fetching service offering", func() {
			It("filters by service offering name", func() {
				_, _, err := actor.EnableServiceAccess("fake-offering", "", "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.NameFilter,
					Values: []string{"fake-offering"},
				}))
			})

			When("the service broker name is specified", func() {
				It("also filters by service broker name", func() {
					_, _, err := actor.EnableServiceAccess("fake-offering", "", "", "fake-broker-name")
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeCloudControllerClient.GetServiceOfferingsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServiceOfferingsArgsForCall(0)).To(ContainElement(ccv3.Query{
						Key:    ccv3.ServiceBrokerNamesFilter,
						Values: []string{"fake-broker-name"},
					}))
				})
			})

			When("the service offering does not exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{}, ccv3.Warnings{"a warning"}, nil)
				})

				It("returns an error", func() {
					_, warnings, err := actor.EnableServiceAccess("no-such-offering", "", "", "")
					Expect(warnings).To(ContainElement("a warning"))
					Expect(err).To(MatchError(actionerror.ServiceNotFoundError{Name: "no-such-offering"}))
				})
			})

			When("the service offering name is ambiguous", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceOfferingsReturns([]ccv3.ServiceOffering{{}, {}}, ccv3.Warnings{"another warning"}, nil)
				})

				It("returns an error", func() {
					_, warnings, err := actor.EnableServiceAccess("duplicate-offering", "", "", "")
					Expect(warnings).To(ContainElement("another warning"))
					Expect(err).To(MatchError(actionerror.DuplicateServiceError{Name: "duplicate-offering"}))
				})
			})
		})

		Describe("fetching service plans", func() {
			It("gets all plans for the service offering", func() {
				_, warnings, err := actor.EnableServiceAccess("fake-offering", "", "", "fake-broker-name")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ContainElements("some warning", "other warning"))

				Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.ServiceOfferingGUIDsFilter,
					Values: []string{"fake-offering-guid"},
				}))
			})

			When("a plan name is specified", func() {
				It("filters by plan name and service offering GUID", func() {
					_, warnings, err := actor.EnableServiceAccess("fake-offering", "fake-plan-name", "", "fake-broker-name")
					Expect(err).NotTo(HaveOccurred())
					Expect(warnings).To(ContainElements("some warning", "other warning"))

					Expect(fakeCloudControllerClient.GetServicePlansCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetServicePlansArgsForCall(0)).To(ConsistOf(
						ccv3.Query{
							Key:    ccv3.ServiceOfferingGUIDsFilter,
							Values: []string{"fake-offering-guid"},
						},
						ccv3.Query{
							Key:    ccv3.NameFilter,
							Values: []string{"fake-plan-name"},
						},
					))
				})
			})

			When("no plans were found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{}, ccv3.Warnings{"other warning"}, nil)
				})

				It("fails", func() {
					_, warnings, err := actor.EnableServiceAccess("fake-offering", "fake-plan-name", "", "fake-broker-name")
					Expect(err).To(MatchError(actionerror.ServicePlanNotFoundError{
						OfferingName: "fake-offering",
						PlanName:     "fake-plan-name",
					}))
					Expect(warnings).To(ContainElements("some warning", "other warning"))
				})
			})

			When("fetching the plans fail", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServicePlansReturns([]ccv3.ServicePlan{}, ccv3.Warnings{"other warning"}, errors.New("fetch plans error"))
				})

				It("fails", func() {
					_, warnings, err := actor.EnableServiceAccess("", "", "", "")
					Expect(err).To(MatchError("fetch plans error"))
					Expect(warnings).To(ContainElements("some warning", "other warning"))
				})
			})
		})

		Context("with org", func() {
			It("sets visibility to orgs", func() {
				_, warnings, err := actor.EnableServiceAccess("fake-offering", "", "fake-org-name", "")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("some warning", "other warning", "post warning", "post warning", "org warning"))

				Expect(fakeCloudControllerClient.UpdateServicePlanVisibilityCallCount()).To(Equal(2))

				planGUID, actualVisibility := fakeCloudControllerClient.UpdateServicePlanVisibilityArgsForCall(0)
				Expect(planGUID).To(Equal("fake-plan-guid-1"))
				Expect(actualVisibility).To(Equal(ccv3.ServicePlanVisibility{
					Type:          "organization",
					Organizations: []ccv3.VisibilityDetail{{GUID: "org-guid"}}}))

				planGUID, actualVisibility = fakeCloudControllerClient.UpdateServicePlanVisibilityArgsForCall(1)
				Expect(planGUID).To(Equal("fake-plan-guid-2"))
				Expect(actualVisibility).To(Equal(ccv3.ServicePlanVisibility{
					Type:          "organization",
					Organizations: []ccv3.VisibilityDetail{{GUID: "org-guid"}}}))
			})

			When("the plan is public", func() {
				It("skips the plan", func() {
					fakeCloudControllerClient.GetServicePlansReturns(
						[]ccv3.ServicePlan{
							{Name: "fake-plan-1", GUID: "fake-plan-guid-1", VisibilityType: "public"},
							{Name: "fake-plan-2", GUID: "fake-plan-guid-2", VisibilityType: "public"},
							{Name: "fake-plan-3", GUID: "fake-plan-guid-3", VisibilityType: "organization"},
							{Name: "fake-plan-4", GUID: "fake-plan-guid-4", VisibilityType: "admin"},
						},
						ccv3.Warnings{"other warning"},
						nil,
					)

					skippedPlans, _, err := actor.EnableServiceAccess("fake-offering", "", "fake-org-name", "")

					Expect(err).NotTo(HaveOccurred())

					Expect(fakeCloudControllerClient.UpdateServicePlanVisibilityCallCount()).To(Equal(2))
					Expect(skippedPlans).To(ConsistOf("fake-plan-1", "fake-plan-2"))
				})
			})

			When("the org does not exist", func() {
				It("returns an error", func() {
					fakeCloudControllerClient.GetOrganizationsReturns([]ccv3.Organization{}, ccv3.Warnings{"org warning"}, nil)

					_, warnings, err := actor.EnableServiceAccess("fake-offering", "", "fake-org-name", "")
					Expect(err).To(MatchError(actionerror.OrganizationNotFoundError{
						Name: "fake-org-name",
					}))
					Expect(warnings).To(ContainElement("org warning"))
				})
			})
		})

		When("setting visibility fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateServicePlanVisibilityReturns(
					ccv3.ServicePlanVisibility{},
					ccv3.Warnings{"post warning"},
					errors.New("post error"),
				)
			})

			It("returns error and stops setting visibility for the remaining plans", func() {
				_, warnings, err := actor.EnableServiceAccess("fake-offering", "", "", "")

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("post error"))

				Expect(fakeCloudControllerClient.UpdateServicePlanVisibilityCallCount()).To(Equal(1))
				Expect(warnings).To(ConsistOf("some warning", "other warning", "post warning"))
			})
		})

		When("the plan has visibility type 'space'", func() {
			It("returns the appropriate error", func() {
				fakeCloudControllerClient.GetServicePlansReturns(
					[]ccv3.ServicePlan{
						{GUID: "fake-plan-guid-1", VisibilityType: "space"},
					},
					ccv3.Warnings{"other warning"},
					nil,
				)
				_, warnings, err := actor.EnableServiceAccess("fake-offering", "", "", "")
				Expect(err).To(MatchError(actionerror.ServicePlanVisibilityTypeError{}))
				Expect(warnings).To(ContainElement("other warning"))
			})
		})
	})
})

func fakeServicePlans() []ccv3.ServicePlan {
	return []ccv3.ServicePlan{
		{
			Name:                "violet",
			ServiceOfferingGUID: "magenta-offering-guid",
			VisibilityType:      "public",
		},
		{
			Name:                "green",
			ServiceOfferingGUID: "cyan-offering-guid",
			VisibilityType:      "space",
		},
		{
			Name:                "indigo",
			ServiceOfferingGUID: "key-offering-guid",
			VisibilityType:      "space",
		},
		{
			Name:                "red",
			ServiceOfferingGUID: "magenta-offering-guid",
			VisibilityType:      "public",
		},
		{
			Name:                "yellow",
			ServiceOfferingGUID: "yellow-offering-guid",
			VisibilityType:      "organization",
		},
		{
			Name:                "orange",
			ServiceOfferingGUID: "yellow-offering-guid",
			VisibilityType:      "organization",
		},
		{
			Name:                "blue",
			ServiceOfferingGUID: "cyan-offering-guid",
			VisibilityType:      "admin",
		},
	}
}

func fakeServiceOfferings() []ccv3.ServiceOffering {
	return []ccv3.ServiceOffering{
		{
			GUID:              "cyan-offering-guid",
			Name:              "cyan",
			ServiceBrokerName: "sky-broker",
		},
		{
			GUID:              "magenta-offering-guid",
			Name:              "magenta",
			ServiceBrokerName: "sea-broker",
		},
		{
			GUID:              "yellow-offering-guid",
			Name:              "yellow",
			ServiceBrokerName: "land-broker",
		},
		{
			GUID:              "key-offering-guid",
			Name:              "key",
			ServiceBrokerName: "sky-broker",
		},
	}
}
