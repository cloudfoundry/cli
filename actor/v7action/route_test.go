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

	Describe("CreateRoute", func() {
		var (
			warnings   Warnings
			executeErr error
			path       string
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateRoute("org-name", "space-name", "domain-name", "hostname", path)
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
					ccv3.Route{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", Path: "path-name"},
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					nil)
			})

			When("the input path starts with '/'", func() {
				BeforeEach(func() {
					path = "/path-name"
				})
				It("returns the route with '/<path>' and prints warnings", func() {
					Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-orgs-warning", "get-domains-warning", "get-spaces-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
					passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

					Expect(passedRoute).To(Equal(
						ccv3.Route{
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "hostname",
							Path:       "/path-name",
						},
					))
				})
			})

			When("the input path does not start with '/'", func() {
				BeforeEach(func() {
					path = "path-name"
				})

				It("returns the route with '/<path>' and prints warnings", func() {
					Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-orgs-warning", "get-domains-warning", "get-spaces-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
					passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

					Expect(passedRoute).To(Equal(
						ccv3.Route{
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "hostname",
							Path:       "/path-name",
						},
					))
				})
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

	Describe("GetRoutesBySpace", func() {
		var (
			routes     []Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetDomainsReturns(
				[]ccv3.Domain{
					{Name: "domain1-name", GUID: "domain1-guid"},
					{Name: "domain2-name", GUID: "domain2-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					{Name: "space-name", GUID: "space-guid"},
				},
				ccv3.Warnings{"get-spaces-warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]ccv3.Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname"},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path"},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetRoutesBySpace("space-guid")
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routes).To(Equal([]Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname", DomainName: "domain1-name", SpaceName: "space-name"},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path", DomainName: "domain2-name", SpaceName: "space-name"},
				}))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-domains-warning", "get-spaces-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSpacesArgsForCall(0)

				Expect(query).To(Equal(
					[]ccv3.Query{
						{
							Key:    "guids",
							Values: []string{"space-guid"},
						},
					},
				))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)

				Expect(query).To(Equal(
					[]ccv3.Query{
						{
							Key:    "space_guids",
							Values: []string{"space-guid"},
						},
					},
				))

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetDomainsArgsForCall(0)

				Expect(query).To(Equal(
					[]ccv3.Query{
						{
							Key:    "guids",
							Values: []string{"domain1-guid", "domain2-guid"},
						},
					},
				))
			})
		})

		When("getting routes fails", func() {
			var err = errors.New("failed to get route")

			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					nil,
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					err)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2"))
			})
		})

		When("getting spaces fails", func() {
			var err = errors.New("failed to get spaces")

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.Warnings{"get-spaces-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning"))
			})
		})

		When("getting domains fails", func() {
			var err = errors.New("failed to get domains")

			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"get-domains-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning", "get-domains-warning"))
			})
		})
	})
})
