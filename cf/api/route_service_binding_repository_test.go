package api_test

import (
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/net"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouteServiceBindingsRepository", func() {
	var (
		ccServer                *ghttp.Server
		configRepo              coreconfig.ReadWriter
		routeServiceBindingRepo api.CloudControllerRouteServiceBindingRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAPIEndpoint(ccServer.URL())

		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		routeServiceBindingRepo = api.NewCloudControllerRouteServiceBindingRepository(configRepo, gateway)
	})

	AfterEach(func() {
		ccServer.Close()
	})

	Describe("Bind", func() {
		var (
			serviceInstanceGUID string
			routeGUID           string
		)

		BeforeEach(func() {
			serviceInstanceGUID = "service-instance-guid"
			routeGUID = "route-guid"
		})

		It("creates the service binding when the service instance is managed", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGUID, routeGUID, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("creates the service binding when the service instance is user-provided", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
					ghttp.RespondWith(http.StatusCreated, nil),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGUID, routeGUID, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("creates the service binding with the provided body wrapped in parameters", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
					ghttp.RespondWith(http.StatusCreated, nil),
					ghttp.VerifyJSON(`{"parameters":{"some":"json"}}`),
				),
			)
			err := routeServiceBindingRepo.Bind(serviceInstanceGUID, routeGUID, true, `{"some":"json"}`)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":61003,"description":"Route does not exist"}`),
					),
				)
			})

			It("returns an HTTPError", func() {
				err := routeServiceBindingRepo.Bind(serviceInstanceGUID, routeGUID, false, "")
				Expect(err).To(HaveOccurred())
				httpErr, ok := err.(errors.HTTPError)
				Expect(ok).To(BeTrue())

				Expect(httpErr.ErrorCode()).To(Equal("61003"))
				Expect(httpErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})

	Describe("Unbind", func() {
		var (
			serviceInstanceGUID string
			routeGUID           string
		)

		BeforeEach(func() {
			serviceInstanceGUID = "service-instance-guid"
			routeGUID = "route-guid"
		})

		It("deletes the service binding when unbinding a managed service instance", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Unbind(serviceInstanceGUID, routeGUID, false)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		It("deletes the service binding when the service instance is user-provided", func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/user_provided_service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
			err := routeServiceBindingRepo.Unbind(serviceInstanceGUID, routeGUID, true)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})

		Context("when an API error occurs", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("/v2/service_instances/%s/routes/%s", serviceInstanceGUID, routeGUID)),
						ghttp.RespondWith(http.StatusBadRequest, `{"code":61003,"description":"Route does not exist"}`),
					),
				)
			})

			It("returns an HTTPError", func() {
				err := routeServiceBindingRepo.Unbind(serviceInstanceGUID, routeGUID, false)
				Expect(err).To(HaveOccurred())
				httpErr, ok := err.(errors.HTTPError)
				Expect(ok).To(BeTrue())

				Expect(httpErr.ErrorCode()).To(Equal("61003"))
				Expect(httpErr.Error()).To(ContainSubstring("Route does not exist"))
			})
		})
	})
})
