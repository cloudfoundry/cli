package isolated

import (
	"path/filepath"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {
	Describe("Empty Config File", func() {
		var configDir string
		BeforeEach(func() {
			configDir = filepath.Join(homeDir, ".cf")
			helpers.SetConfigContent(configDir, "")
		})

		It("displays json warning for a refactored command", func() {
			session := helpers.CF("api")
			Eventually(session.Err).Should(Say("Warning: Error read/writing config: unexpected end of JSON input for %s", filepath.Join(configDir, "config.json")))
		})

		It("displays json warning for an unrefactored command", func() {
			session := helpers.CF("curl")
			Eventually(session.Err).Should(Say("Warning: Error read/writing config: unexpected end of JSON input for %s", filepath.Join(configDir, "config.json")))
		})
	})

	Describe("Enable Color", func() {
		Context("when color is enabled", func() {
			It("prints colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "true"}, "help")
				Eventually(session).Should(Say("\x1b\\[1m"))
			})
		})

		Context("when color is disabled", func() {
			It("does not print colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "false"}, "help")
				Consistently(session).ShouldNot(Say("\x1b\\[1m"))
				Eventually(session).Should(Exit(0))
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
