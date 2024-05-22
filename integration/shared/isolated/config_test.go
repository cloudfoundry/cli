package isolated

import (
	"os"
	"path/filepath"
	"time"

	helpers "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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
		When("lingering tmp files exist from previous failed attempts to write the config", func() {
			BeforeEach(func() {
				for i := 0; i < 3; i++ {
					tmpFile, err := os.CreateTemp(configDir, "temp-config")
					Expect(err).ToNot(HaveOccurred())
					tmpFile.Close()
					oldTime := time.Now().Add(-time.Minute * 10)
					err = os.Chtimes(tmpFile.Name(), oldTime, oldTime)
					Expect(err).ToNot(HaveOccurred())
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
		When("color is enabled", func() {
			It("prints colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "true"}, "help")
				Eventually(session).Should(Say("\x1b\\[1m"))
				Eventually(session).Should(Exit())
			})
		})

		When("color is disabled", func() {
			It("does not print colors", func() {
				session := helpers.CFWithEnv(map[string]string{"CF_COLOR": "false"}, "help")
				Consistently(session).ShouldNot(Say("\x1b\\[1m"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
