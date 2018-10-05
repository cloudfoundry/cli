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
		proxyURL = "127.0.0.1:9999"
	})

	Context("V2", func() {
		It("errors when proxy is not setup properly", func() {
			session := helpers.CFWithEnv(map[string]string{"https_proxy": proxyURL}, "api", apiURL)
			Eventually(session.Err).Should(Say("%s/v2/info.*proxy.*%s", apiURL, proxyURL))
			Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("V3", func() {
		It("errors when proxy is not setup properly", func() {
			session := helpers.CFWithEnv(map[string]string{"https_proxy": proxyURL}, "run-task", "app", "echo")
			Eventually(session.Err).Should(Say("%s.*proxy.*%s", apiURL, proxyURL))
			Eventually(session.Err).Should(Say("TIP: If you are behind a firewall and require an HTTP proxy, verify the https_proxy environment variable is correctly set. Else, check your network connection."))
			Eventually(session).Should(Exit(1))
		})
	})
})
