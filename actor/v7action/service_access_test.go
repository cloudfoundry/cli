package v7action_test

import (
	"errors"

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

	Describe("GetServiceAccess", func() {
		//Filtering
		//Displaying

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

		When("the client fails to return the plans", func() {
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

		When("the client fails to return the offerings", func() {
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

		When("the client fails to return the visibility for the plan", func() {
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
