package selfcontained_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("cf api", func() {
	var (
		server       *ghttp.Server
		responseBody string
	)

	BeforeEach(func() {
		responseBody = "{}"
	})

	JustBeforeEach(func() {
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/"),
				ghttp.RespondWith(http.StatusOK, responseBody),
			),
		)

		Eventually(helpers.CF("api", server.URL())).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		server.Close()
	})

	It("disables cf-on-k8s in config", func() {
		Expect(loadConfig().CFOnK8s.Enabled).To(BeFalse())
	})

	When("pointed to cf-on-k8s", func() {
		BeforeEach(func() {
			responseBody = `{ "cf_on_k8s": true }`
		})

		It("enables cf-on-k8s in config", func() {
			Expect(loadConfig().CFOnK8s.Enabled).To(BeTrue())
		})
	})
})

func loadConfig() configv3.JSONConfig {
	rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
	Expect(err).NotTo(HaveOccurred())

	var configFile configv3.JSONConfig
	Expect(json.Unmarshal(rawConfig, &configFile)).To(Succeed())

	return configFile
}
