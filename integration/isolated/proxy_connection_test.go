package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("proxy", func() {
	var proxyURL string

	BeforeEach(func() {
		proxyURL = "http://127.0.0.1:9999"
	})

	Context("V2 Legacy", func() {
		It("handles a proxy", func() {
			Skip("Figure out how to test not using API command")
			helpers.SkipIfExperimental("error messages have changed in refactored code")
			session := helpers.CFWithEnv(map[string]string{"https_proxy": proxyURL}, "api", apiURL)
			Eventually(session).Should(Say("Error performing request: Get %s/v2/info: http: error connecting to proxy %s", apiURL, proxyURL))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("V2", func() {
		It("handles a proxy", func() {
			session := helpers.CFWithEnv(map[string]string{"https_proxy": proxyURL}, "api", apiURL)
			Eventually(session.Err).Should(Say("Get %s/v2/info: http: error connecting to proxy %s", apiURL, proxyURL))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("V3", func() {
		It("handles a proxy", func() {
			session := helpers.CFWithEnv(map[string]string{"https_proxy": proxyURL}, "run-task", "app", "echo")
			Eventually(session.Err).Should(Say("Get %s: http: error connecting to proxy %s", apiURL, proxyURL))
			Eventually(session).Should(Exit(1))
		})
	})
})
