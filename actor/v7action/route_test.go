package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("create route with private domain", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateRoute("org-name", "space-name", "domain-name", "hostname")
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]ccv3.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{
						{Name: "org-name", GUID: "org-guid"},
					},
					ccv3.Warnings{"get-orgs-warning"},
					nil,
				)

				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{
						{Name: "space-name", GUID: "space-guid"},
					},
					ccv3.Warnings{"get-spaces-warning"},
					nil,
				)

				fakeCloudControllerClient.CreateRouteReturns(
					ccv3.Route{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname"},
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					nil)
			})

			It("returns the route and prints warnings", func() {
				Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-orgs-warning", "get-domains-warning", "get-spaces-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

				Expect(passedRoute).To(Equal(
					ccv3.Route{
						SpaceGUID:  "space-guid",
						DomainGUID: "domain-guid",
						Host:       "hostname",
					},
				))
			})
		})

		When("the API call to get the domain returns an error", func() {
			When("the cc client returns an RouteNotUniqueError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]ccv3.Domain{
							{Name: "domain-name", GUID: "domain-guid"},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)

					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{
							{Name: "org-name", GUID: "org-guid"},
						},
						ccv3.Warnings{"get-orgs-warning"},
						nil,
					)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{
							{Name: "space-name", GUID: "space-guid"},
						},
						ccv3.Warnings{"get-spaces-warning"},
						nil,
					)

					fakeCloudControllerClient.CreateRouteReturns(
						ccv3.Route{},
						ccv3.Warnings{"create-route-warning"},
						ccerror.RouteNotUniqueError{
							UnprocessableEntityError: ccerror.UnprocessableEntityError{Message: "some cool error"},
						},
					)
				})

				It("returns the RouteAlreadyExistsError and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.RouteAlreadyExistsError{
						Err: ccerror.RouteNotUniqueError{
							UnprocessableEntityError: ccerror.UnprocessableEntityError{Message: "some cool error"},
						},
					}))
					Expect(warnings).To(ConsistOf("get-domains-warning",
						"get-orgs-warning",
						"get-spaces-warning",
						"create-route-warning"))
				})
			})

			When("the cc client returns a different error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]ccv3.Domain{},
						ccv3.Warnings{"domain-warning-1", "domain-warning-2"},
						errors.New("api-domains-error"),
					)
				})

				It("it returns an error and prints warnings", func() {
					Expect(warnings).To(ConsistOf("domain-warning-1", "domain-warning-2"))
					Expect(executeErr).To(MatchError("api-domains-error"))

					Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(0))

				})
			})
		})

	})
})
