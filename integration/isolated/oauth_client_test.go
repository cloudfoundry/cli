package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func fileAsString(path string) string {
	configBytes, err := ioutil.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())

	return string(configBytes)
}

func replaceConfig(path string, old string, new string) {
	r := regexp.MustCompile(old)
	newConfig := r.ReplaceAllString(fileAsString(path), new)
	err := ioutil.WriteFile(path, []byte(newConfig), 0600)
	Expect(err).ToNot(HaveOccurred())
}

var _ = Describe("custom oauth client id", func() {
	var configPath string

	BeforeEach(func() {
		configPath = filepath.Join(homeDir, ".cf", "config.json")
	})

	Context("when the config file exists", func() {
		BeforeEach(func() {
			setupCF(ReadOnlyOrg, ReadOnlySpace)
		})

		Context("when the client id and secret keys are set in the config", func() {
			BeforeEach(func() {
				replaceConfig(
					configPath, `"UAAOAuthClient": ".*"`, `"UAAOAuthClient": "cf2"`)
				replaceConfig(
					configPath, `"UAAOAuthClientSecret": ".*"`, `"UAAOAuthClientSecret": "secret2"`)
			})

			Context("oauth-token", func() {
				It("uses the custom client id and secret", func() {
					session := helpers.CF("oauth-token")
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say(
						"Server error, status code: 401, error code: unauthorized"))
				})
			})

			Context("auth", func() {
				It("uses the custom client id and secret", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("auth", username, password)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say(
						"Credentials were rejected, please try again."))
				})
			})

			Context("login", func() {
				It("uses the custom client id and secret", func() {
					username, password := helpers.GetCredentials()
					session := helpers.CF("login", "-u", username, "-p", password)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say(
						"Credentials were rejected, please try again."))
				})
			})
		})

		Context("when the client id in the config is empty", func() {
			BeforeEach(func() {
				replaceConfig(
					configPath, `"UAAOAuthClient": ".*",`, `"UAAOAuthClient": "",`)
			})

			Context("v2 command", func() {
				It("does not write default values for client id and secret", func() {
					session := helpers.CF("oauth-token")
					Eventually(session).Should(Exit(1))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": ""`))
				})
			})

			Context("v3 command", func() {
				It("writes default values for client id and secret", func() {
					session := helpers.CF("tasks", "some-app")
					Eventually(session).Should(Exit(1))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})
		})

		Context("when there are no client id and secret keys in the config", func() {
			BeforeEach(func() {
				replaceConfig(
					configPath, `"UAAOAuthClient": ".*",`, "")
				replaceConfig(
					configPath, `"UAAOAuthClientSecret": ".*",`, "")
			})

			Context("v2 command", func() {
				It("writes default values for client id and secret", func() {
					session := helpers.CF("oauth-token")
					Eventually(session).Should(Exit(0))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})

			Context("v3 command", func() {
				It("writes default values for client id and secret", func() {
					session := helpers.CF("tasks")
					Eventually(session).Should(Exit(1))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})
		})
	})

	Context("when the config file does not exist", func() {
		BeforeEach(func() {
			err := os.Remove(configPath)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("v2 command", func() {
			It("writes default values for client id and secret to the config", func() {
				Expect(configPath).ToNot(BeAnExistingFile())

				session := helpers.CF("help")
				Eventually(session).Should(Exit(0))

				configString := fileAsString(configPath)
				Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
				Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
			})
		})

		Context("v3 command", func() {
			It("writes default values for client id and secret to the config", func() {
				Expect(configPath).ToNot(BeAnExistingFile())

				session := helpers.CF("tasks")
				Eventually(session).Should(Exit(1))

				configString := fileAsString(configPath)
				Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
				Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
			})
		})
	})
})
