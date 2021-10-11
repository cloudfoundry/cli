package selfcontained_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("cf api", func() {
	var apiConfig fake.CFAPIConfig

	BeforeEach(func() {
		apiConfig = fake.CFAPIConfig{
			Routes: map[string]fake.Response{
				"GET /": {Code: http.StatusOK, Body: ccv3.Info{}},
			},
		}
		apiServer.SetConfiguration(apiConfig)
	})

	JustBeforeEach(func() {
		Eventually(helpers.CF("api", apiServer.URL())).Should(gexec.Exit(0))
	})

	It("disables cf-on-k8s in config", func() {
		Expect(loadConfig().CFOnK8s.Enabled).To(BeFalse())
	})

	When("pointed to cf-on-k8s", func() {
		BeforeEach(func() {
			apiConfig.Routes["GET /"] = fake.Response{
				Code: http.StatusOK, Body: ccv3.Info{CFOnK8s: true},
			}
			apiServer.SetConfiguration(apiConfig)
		})

		It("enables cf-on-k8s in config", func() {
			Expect(loadConfig().CFOnK8s.Enabled).To(BeTrue())
		})
	})
})
