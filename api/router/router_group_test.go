package router_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
	"code.cloudfoundry.org/cli/api/router/wrapper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Router Groups", func() {
	Describe("GetRouterGroupsByName", func() {
		var (
			client          *Client
			fakeConfig      Config
			routerGroup     RouterGroup
			executeErr      error
			routerGroupName string
		)

		JustBeforeEach(func() {
			fakeConfig = NewTestConfig()
			fakeConfig.Wrappers = append([]ConnectionWrapper{wrapper.NewErrorWrapper()}, fakeConfig.Wrappers...)
			client = NewTestRouterClient(fakeConfig)
			routerGroup, executeErr = client.GetRouterGroupByName(routerGroupName)
		})

		When("the request fails", func() {
			BeforeEach(func() {
				response := `{"name":"ResourceNotFoundError","message":"Router Group 'not-a-real-router-group' not found"}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/routing/v1/router_groups"),
						VerifyHeaderKV("Content-Type", "application/json"),
						RespondWith(http.StatusNotFound, response),
					))
			})

			It("returns the error", func() {
				Expect(executeErr).To(HaveOccurred())
				expectedErr := routererror.ResourceNotFoundError{
					Message: "Router Group 'not-a-real-router-group' not found",
				}
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(routerGroup).To(Equal(RouterGroup{}))
			})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `[
					{
						"guid":"some-router-group-guid-1",
						"name":"some-router-group",
						"type":"tcp",
						"reservable_ports":"1024-1123"
					}
				]`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/routing/v1/router_groups", "name=some-router-group"),
						VerifyHeaderKV("Content-Type", "application/json"),
						RespondWith(http.StatusOK, response),
					))
				routerGroupName = "some-router-group"
			})

			It("returns the list of router groups and no errors", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(routerGroup).To(Equal(RouterGroup{
					GUID:            "some-router-group-guid-1",
					Name:            "some-router-group",
					Type:            "tcp",
					ReservablePorts: "1024-1123",
				}))
			})
		})
	})
})
