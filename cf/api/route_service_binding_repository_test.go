package api_test

import (
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"

	"github.com/cloudfoundry/cli/testhelpers/cloud_controller_gateway"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteServiceBindingsRepository", func() {
	var (
		ccServer                *ghttp.Server
		configRepo              core_config.ReadWriter
		routeServiceBindingRepo api.CloudControllerRouteServiceBindingRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(ccServer.URL())

		gateway := cloud_controller_gateway.NewTestCloudControllerGateway(configRepo)
		routeServiceBindingRepo = api.NewCloudControllerRouteServiceBindingRepository(configRepo, gateway)
	})

	AfterEach(func() {
		ccServer.Close()
	})

	Describe("Bind", func() {
		var (
			serviceInstanceGuid string
			routeGuid           string
		)

		BeforeEach(func() {
			serviceInstanceGuid = "service-instance-guid"
			routeGuid = "route-guid"
		})

		It("creates the service binding when the service instance is managed", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGuid, routeGuid, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("creates the service binding when the service instance is user-provided", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
					ghttp.RespondWith(http.StatusCreated, nil),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGuid, routeGuid, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("creates the service binding with the provided body wrapped in parameters", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
					ghttp.RespondWith(http.StatusCreated, nil),
					ghttp.VerifyJSON(`{"parameters":{"some":"json"}}`),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGuid, routeGuid, true, `{"some":"json"}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":61003,"description":"Route does not exist"}`),
					),
				)
			})

			It("returns an HttpError", func() {
				err := routeServiceBindingRepo.Bind(serviceInstanceGuid, routeGuid, false, "")
				Expect(err).To(HaveOccurred())
				httpErr, ok := err.(errors.HttpError)
				Expect(ok).To(BeTrue())

				Expect(httpErr.ErrorCode()).To(Equal("61003"))
				Expect(httpErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})

	Describe("Unbind", func() {
		var (
			serviceInstanceGuid string
			routeGuid           string
		)

		BeforeEach(func() {
			serviceInstanceGuid = "service-instance-guid"
			routeGuid = "route-guid"
		})

		It("deletes the service binding when unbinding a managed service instance", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Unbind(serviceInstanceGuid, routeGuid, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("deletes the service binding when the service instance is user-provided", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Unbind(serviceInstanceGuid, routeGuid, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGuid, routeGuid)),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":61003,"description":"Route does not exist"}`),
					),
				)
			})

			It("returns an HttpError", func() {
				err := routeServiceBindingRepo.Unbind(serviceInstanceGuid, routeGuid, false)
				Expect(err).To(HaveOccurred())
				httpErr, ok := err.(errors.HttpError)
				Expect(ok).To(BeTrue())

				Expect(httpErr.ErrorCode()).To(Equal("61003"))
				Expect(httpErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})
})
