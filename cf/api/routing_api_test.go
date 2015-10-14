package api_test

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("RoutingApi", func() {

	var (
		repo             api.RoutingApiRepository
		configRepo       core_config.Repository
		routingApiServer *ghttp.Server
	)

	BeforeEach(func() {
		configRepo = testconfig.NewRepositoryWithDefaults()
		gateway := net.NewRoutingApiGateway(configRepo, time.Now, &testterm.FakeUI{})

		repo = api.NewRoutingApiRepository(configRepo, gateway)
	})

	AfterEach(func() {
		routingApiServer.Close()
	})

	Describe("ListRouterGroups", func() {

		Context("when routing api return router groups", func() {
			BeforeEach(func() {
				routingApiServer = ghttp.NewServer()
				routingApiServer.RouteToHandler("GET", "/v1/router_groups",
					func(w http.ResponseWriter, req *http.Request) {
						responseBody := []byte(`[
			{
				  "guid": "bad25cff-9332-48a6-8603-b619858e7992",
					"name": "default-tcp",
					"type": "tcp"
			}]`)
						w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write(responseBody)
					})
				configRepo.SetRoutingApiEndpoint(routingApiServer.URL())
			})

			It("lists routing groups", func() {
				cb := func(grp models.RouterGroup) bool {
					Expect(grp).To(Equal(models.RouterGroup{
						Guid: "bad25cff-9332-48a6-8603-b619858e7992",
						Name: "default-tcp",
						Type: "tcp",
					}))
					return true
				}
				err := repo.ListRouterGroups(cb)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when routing api returns an empty response ", func() {
			BeforeEach(func() {
				routingApiServer = ghttp.NewServer()
				routingApiServer.RouteToHandler("GET", "/v1/router_groups",
					func(w http.ResponseWriter, req *http.Request) {
						responseBody := []byte("[]")
						w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						w.Write(responseBody)
					})
				configRepo.SetRoutingApiEndpoint(routingApiServer.URL())
			})

			It("doesn't list any router groups", func() {
				cb := func(grp models.RouterGroup) bool {
					Fail(fmt.Sprintf("Not expected to receive callback for router group:%#v", grp))
					return false
				}
				err := repo.ListRouterGroups(cb)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when routing api returns an error ", func() {
			BeforeEach(func() {
				routingApiServer = ghttp.NewServer()
				routingApiServer.RouteToHandler("GET", "/v1/router_groups",
					func(w http.ResponseWriter, req *http.Request) {
						w.WriteHeader(http.StatusUnauthorized)
						w.Write([]byte(`{"name":"UnauthorizedError","message":"token is expired"}`))
					})
				configRepo.SetRoutingApiEndpoint(routingApiServer.URL())
			})

			It("doesn't list any router groups and displays error message", func() {
				cb := func(grp models.RouterGroup) bool {
					Fail(fmt.Sprintf("Not expected to receive callback for router group:%#v", grp))
					return false
				}

				err := repo.ListRouterGroups(cb)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("token is expired"))
				Expect(err.(errors.HttpError).ErrorCode()).To(ContainSubstring("UnauthorizedError"))
				Expect(err.(errors.HttpError).StatusCode()).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})
