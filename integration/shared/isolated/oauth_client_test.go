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

	When("the config file exists", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
		})

		When("the client id and secret keys are set in the config", func() {
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
					Expect(session.Err).To(Say(`Credentials were rejected, please try again\.`))
				})
			})

			Context("auth", func() {
				It("uses the custom client id and secret", func() {
					helpers.SkipIfClientCredentialsTestMode()
					username, password := helpers.GetCredentials()
					env := map[string]string{
						"CF_USERNAME": username,
						"CF_PASSWORD": password,
					}
					session := helpers.CFWithEnv(env, "auth")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say(
						"Credentials were rejected, please try again."))
				})
			})

			Context("login", func() {
				It("uses the custom client id and secret", func() {
					helpers.SkipIfClientCredentialsTestMode()
					username, password := helpers.GetCredentials()
					session := helpers.CF("login", "-u", username, "-p", password)
					Eventually(session).Should(Exit(1))
					Expect(session).To(Say(
						"Credentials were rejected, please try again."))
				})
			})
		})

		When("the client id in the config is empty", func() {
			BeforeEach(func() {
				replaceConfig(
					configPath, `"UAAOAuthClient": ".*",`, `"UAAOAuthClient": "",`)
			})

			Context("v6 command", func() {
				It("replaces the empty client id with the default values for client id and secret", func() {
					// when authenticated as a client, client credentials must be present
					helpers.SkipIfClientCredentialsTestMode()
					session := helpers.CF("oauth-token")
					Eventually(session).Should(Exit(0))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})

			Context("v7 command", func() {
				It("writes default values for client id and secret", func() {
					session := helpers.CF("tasks", "some-app")
					Eventually(session).Should(Exit(1))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})
		})

		When("there are no client id and secret keys in the config", func() {
			BeforeEach(func() {
				replaceConfig(
					configPath, `"UAAOAuthClient": ".*",`, "")
				replaceConfig(
					configPath, `"UAAOAuthClientSecret": ".*",`, "")
			})

			Context("v6 command", func() {
				It("writes default values for client id and secret", func() {
					helpers.SkipIfClientCredentialsTestMode()
					session := helpers.CF("oauth-token")
					Eventually(session).Should(Exit(0))

					configString := fileAsString(configPath)
					Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
					Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
				})
			})

			Context("v7 command", func() {
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

	When("the config file does not exist", func() {
		BeforeEach(func() {
			err := os.Remove(configPath)
			Expect(err).ToNot(HaveOccurred())
		})

		Context("v6 command", func() {
			It("writes default values for client id and secret to the config", func() {
				Expect(configPath).ToNot(BeAnExistingFile())

				session := helpers.CF("help")
				Eventually(session).Should(Exit(0))

				configString := fileAsString(configPath)
				Expect(configString).To(ContainSubstring(`"UAAOAuthClient": "cf"`))
				Expect(configString).To(ContainSubstring(`"UAAOAuthClientSecret": ""`))
			})
		})

		Context("v7 command", func() {
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
