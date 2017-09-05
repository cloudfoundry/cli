package isolated

import (
	"io/ioutil"
	"path/filepath"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Config", func() {
	var configDir string

	BeforeEach(func() {
		configDir = filepath.Join(homeDir, ".cf")
	})

	Describe("Empty Config File", func() {
		BeforeEach(func() {
			helpers.SetConfigContent(configDir, "")
		})

		It("displays json warning for a refactored command", func() {
			session := helpers.CF("api")
			Eventually(session.Err).Should(Say("Warning: Error read/writing config: unexpected end of JSON input for %s\n", helpers.ConvertPathToRegularExpression(filepath.Join(configDir, "config.json"))))
			Eventually(session).Should(Exit())
		})

		It("displays json warning for an unrefactored command", func() {
			session := helpers.CF("curl", "/v2/info")
			Eventually(session.Err).Should(Say("Warning: Error read/writing config: unexpected end of JSON input for %s\n", helpers.ConvertPathToRegularExpression(filepath.Join(configDir, "config.json"))))
			Eventually(session).Should(Exit())
		})
	})

	Describe("Lingering Config Temp Files", func() {
		Context("when lingering tmp files exist from previous failed attempts to write the config", func() {
			BeforeEach(func() {
				for i := 0; i < 3; i++ {
					tmpFile, err := ioutil.TempFile(configDir, "temp-config")
					Expect(err).ToNot(HaveOccurred())
					tmpFile.Close()
				}
			})

			It("removes those temp files on `logout`", func() {
				Eventually(helpers.CF("logout")).Should(Exit(0))

				oldTempFileNames, err := filepath.Glob(filepath.Join(configDir, "temp-config?*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(oldTempFileNames).To(BeEmpty())
			})

			It("removes those temp files on `login`", func() {
				Eventually(helpers.CF("login")).Should(Exit(1))

				oldTempFileNames, err := filepath.Glob(filepath.Join(configDir, "temp-config?*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(oldTempFileNames).To(BeEmpty())
			})

			It("removes those temp files on `auth`", func() {
				helpers.LoginCF()

				oldTempFileNames, err := filepath.Glob(filepath.Join(configDir, "temp-config?*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(oldTempFileNames).To(BeEmpty())
			})

			It("removes those temp files on `oauth-token`", func() {
				Eventually(helpers.CF("oauth-token")).Should(Exit(1))

				oldTempFileNames, err := filepath.Glob(filepath.Join(configDir, "temp-config?*"))
				Expect(err).ToNot(HaveOccurred())
				Expect(oldTempFileNames).To(BeEmpty())
			})
		})
	})

	Describe("Enable Color", func() {
		Context("when color is enabled", func() {
			It("prints colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "true"}, "help")
				Eventually(session).Should(Say("\x1b\\[1m"))
				Eventually(session).Should(Exit())
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
				Eventually(session).Should(Exit())
			})
		})
	})
})
