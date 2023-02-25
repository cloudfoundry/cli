package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("marketplace", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("Marketplace", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetServicePlansWithOfferingsReturns(
				[]ccv3.ServiceOfferingWithPlans{
					{
						GUID:              "offering-guid-1",
						Name:              "offering-1",
						Description:       "about offering 1",
						ServiceBrokerName: "service-broker-1",
						Plans: []resources.ServicePlan{
							{
								GUID: "plan-guid-1",
								Name: "plan-1",
							},
						},
					},
					{
						GUID:              "offering-guid-2",
						Name:              "offering-2",
						Description:       "about offering 2",
						ServiceBrokerName: "service-broker-2",
						Plans: []resources.ServicePlan{
							{
								GUID: "plan-guid-2",
								Name: "plan-2",
							},
							{
								GUID: "plan-guid-3",
								Name: "plan-3",
							},
						},
					},
				},
				ccv3.Warnings{"foo", "bar"},
				nil,
			)
		})

		It("calls the client", func() {
			_, _, _ = actor.Marketplace(MarketplaceFilter{})

			Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
			queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
			Expect(queries).To(ConsistOf(
				ccv3.Query{Key: ccv3.AvailableFilter, Values: []string{"true"}},
				ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
			))
		})

		It("returns a list of service offerings and plans", func() {
			offerings, warnings, err := actor.Marketplace(MarketplaceFilter{})
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ConsistOf("foo", "bar"))
			Expect(offerings).To(Equal([]ServiceOfferingWithPlans{
				{
					GUID:              "offering-guid-1",
					Name:              "offering-1",
					Description:       "about offering 1",
					ServiceBrokerName: "service-broker-1",
					Plans: []resources.ServicePlan{
						{
							GUID: "plan-guid-1",
							Name: "plan-1",
						},
					},
				},
				{
					GUID:              "offering-guid-2",
					Name:              "offering-2",
					Description:       "about offering 2",
					ServiceBrokerName: "service-broker-2",
					Plans: []resources.ServicePlan{
						{
							GUID: "plan-guid-2",
							Name: "plan-2",
						},
						{
							GUID: "plan-guid-3",
							Name: "plan-3",
						},
					},
				},
			}))
		})

		When("a space GUID is specified", func() {
			It("adds the GUID to the query", func() {
				_, _, _ = actor.Marketplace(MarketplaceFilter{SpaceGUID: "space-guid"})

				Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
				Expect(queries).To(ContainElement(ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}}))
			})
		})

		When("a service offering name is specified", func() {
			It("adds the service offering name to the query", func() {
				_, _, _ = actor.Marketplace(MarketplaceFilter{ServiceOfferingName: "my-service-offering"})

				Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
				Expect(queries).To(ContainElement(ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{"my-service-offering"}}))
			})
		})

		When("a service broker name is specified", func() {
			It("adds the service broker name to the query", func() {
				_, _, _ = actor.Marketplace(MarketplaceFilter{ServiceBrokerName: "my-service-broker"})

				Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
				Expect(queries).To(ContainElement(ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{"my-service-broker"}}))
			})
		})

		When("the show unavailable filter is specified", func() {
			It("does not add the `available` query parameter", func() {
				_, _, _ = actor.Marketplace(MarketplaceFilter{ShowUnavailable: true})

				Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
				Expect(queries).NotTo(ContainElement(ccv3.Query{Key: ccv3.AvailableFilter, Values: []string{"true"}}))
			})
		})

		When("all filters are specified", func() {
			It("adds all the filters to the query", func() {
				_, _, _ = actor.Marketplace(MarketplaceFilter{
					SpaceGUID:           "space-guid",
					ServiceBrokerName:   "my-service-broker",
					ServiceOfferingName: "my-service-offering",
				})

				Expect(fakeCloudControllerClient.GetServicePlansWithOfferingsCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetServicePlansWithOfferingsArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{Key: ccv3.ServiceBrokerNamesFilter, Values: []string{"my-service-broker"}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}},
					ccv3.Query{Key: ccv3.ServiceOfferingNamesFilter, Values: []string{"my-service-offering"}},
					ccv3.Query{Key: ccv3.AvailableFilter, Values: []string{"true"}},
					ccv3.Query{Key: ccv3.PerPage, Values: []string{ccv3.MaxPerPage}},
				))
			})
		})

		When("the client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetServicePlansWithOfferingsReturns(
					[]ccv3.ServiceOfferingWithPlans{{}},
					ccv3.Warnings{"foo", "bar"},
					errors.New("bang"),
				)
			})

			It("fails", func() {
				offerings, warnings, err := actor.Marketplace(MarketplaceFilter{})
				Expect(err).To(MatchError("bang"))
				Expect(warnings).To(ConsistOf("foo", "bar"))
				Expect(offerings).To(BeNil())
			})
		})
	})
})
