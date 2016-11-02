package ccv3_test

import (
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Info", func() {
	var (
		client *CloudControllerClient
		// statusCode   int
	)

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("Info", func() {
		BeforeEach(func() {
			response := `{
   "links": {
      "self": {
         "href": "https://api.bosh-lite.com/v3"
      },
      "tasks": {
         "href": "https://api.bosh-lite.com/v3/tasks"
      },
      "uaa": {
         "href": "https://uaa.bosh-lite.com"
      }
   }
}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3/"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		It("returns back the CC Information", func() {
			info, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.UAA()).To(Equal("https://uaa.bosh-lite.com"))
		})

		It("sets the http endpoint and warns user", func() {
			_, warnings, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(warnings).To(ContainElement("this is a warning"))
		})
	})
	Context("when the uaa endpoint does not exist", func() {
		BeforeEach(func() {
			response := `{
   "links": {
      "self": {
         "href": "https://api.bosh-lite.com/v3"
      },
      "tasks": {
         "href": "https://api.bosh-lite.com/v3/tasks"
      }
   }
}`
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v3/"),
					RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
				),
			)
		})

		It("returns an empty endpoint", func() {
			info, _, err := client.Info()
			Expect(err).NotTo(HaveOccurred())
			Expect(info.UAA()).To(BeEmpty())
			//What else do we add to test here?
		})
	})
})
