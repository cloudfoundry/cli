package isolated

import (
	helpers "code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {
	Describe("Enable Color", func() {
		Context("when color is enabled", func() {
			It("prints colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "true"}, "help")
				Eventually(session).Should(Say("\x1b\\[38;1m"))
			})
		})

		Context("when color is disabled", func() {
			It("does not print colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "false"}, "help")
				Eventually(session).Should(Exit(0))
				Expect(session).NotTo(Say("\x1b\\[38;1m"))
			})
		})
	})

	Describe("Dial Timeout", func() {
		Context("when the dial timeout is set", func() {
			BeforeEach(func() {
				config, err := configv3.LoadConfig()
				Expect(err).ToNot(HaveOccurred())

				config.ConfigFile.Target = "http://1.2.3.4"

				err = configv3.WriteConfig(config)
				Expect(err).ToNot(HaveOccurred())
			})

			It("times out connection attempts after the dial timeout has passed", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_DIAL_TIMEOUT": "1"}, "unbind-service", "banana", "pants")
				Eventually(session.Err).Should(Say("dial tcp 1.2.3.4:80: i/o timeout"))
			})
		})
	})
})
