package isolated

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("api command", func() {
	Context("no arguments", func() {
		Context("when the api is set", func() {
			Context("when the user is not logged in", func() {
				It("outputs the current api", func() {
					session := helpers.CF("api")

					Eventually(session).Should(Say("api endpoint:\\s+%s", apiURL))
					Eventually(session).Should(Say("api version:\\s+\\d+\\.\\d+\\.\\d+"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the user is logged in", func() {
				var target, apiVersion, org, space string

				BeforeEach(func() {
					target = "https://api.fake.com"
					apiVersion = "2.59.0"
					org = "the-org"
					space = "the-space"

					userConfig := configv3.Config{
						ConfigFile: configv3.CFConfig{
							Target:      target,
							APIVersion:  apiVersion,
							AccessToken: "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiI3YzZkMDA2MjA2OTI0NmViYWI0ZjBmZjY3NGQ3Zjk4OSIsInN1YiI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ3MzI4NDU3NywicmV2X3NpZyI6IjZiMjdkYTZjIiwiaWF0IjoxNDczMjg0NTc3LCJleHAiOjE0NzMyODUxNzcsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2YiLCJvcGVuaWQiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiLCJzY2ltIiwiY2xvdWRfY29udHJvbGxlciIsInVhYSIsInBhc3N3b3JkIiwiZG9wcGxlciJdfQ.OcH_w9yIKJkEcTZMThIs-qJAHk3G0JwNjG-aomVH9hKye4ciFO6IMQMLKmCBrrAQVc7ST1SZZwq7gv12Dq__6Jp-hai0a2_ADJK-Vc9YXyNZKgYTWIeVNGM1JGdHgFSrBR2Lz7IIrH9HqeN8plrKV5HzU8uI9LL4lyOCjbXJ9cM",
							TargetedOrganization: configv3.Organization{
								Name: org,
							},
							TargetedSpace: configv3.Space{
								Name: space,
							},
						},
					}
					err := configv3.WriteConfig(&userConfig)
					Expect(err).ToNot(HaveOccurred())
				})

				It("outputs the user's target information", func() {
					session := helpers.CF("api")
					Eventually(session).Should(Say("api endpoint:\\s+%s", target))
					Eventually(session).Should(Say("api version:\\s+%s", apiVersion))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the api is not set", func() {
			BeforeEach(func() {
				os.RemoveAll(filepath.Join(homeDir, ".cf"))
			})

			It("outputs that nothing is set", func() {
				session := helpers.CF("api")
				Eventually(session).Should(Say("No api endpoint set. Use 'cf api' to set an endpoint"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("--unset is passed", func() {
			BeforeEach(func() {

				userConfig := configv3.Config{
					ConfigFile: configv3.CFConfig{
						ConfigVersion: 3,
						Target:        "https://api.fake.com",
						APIVersion:    "2.59.0",
						AccessToken:   "bearer tokenstuff",
						TargetedOrganization: configv3.Organization{
							Name: "the-org",
						},
						TargetedSpace: configv3.Space{
							Name: "the-space",
						},
					},
				}
				err := configv3.WriteConfig(&userConfig)
				Expect(err).ToNot(HaveOccurred())
			})

			It("clears the targetted context", func() {
				session := helpers.CF("api", "--unset")

				Eventually(session).Should(Say("Unsetting api endpoint..."))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
				Expect(err).NotTo(HaveOccurred())

				var configFile configv3.CFConfig
				err = json.Unmarshal(rawConfig, &configFile)
				Expect(err).NotTo(HaveOccurred())

				Expect(configFile.ConfigVersion).To(Equal(3))
				Expect(configFile.Target).To(BeEmpty())
				Expect(configFile.APIVersion).To(BeEmpty())
				Expect(configFile.AuthorizationEndpoint).To(BeEmpty())
				Expect(configFile.DopplerEndpoint).To(BeEmpty())
				Expect(configFile.UAAEndpoint).To(BeEmpty())
				Expect(configFile.AccessToken).To(BeEmpty())
				Expect(configFile.RefreshToken).To(BeEmpty())
				Expect(configFile.TargetedOrganization.GUID).To(BeEmpty())
				Expect(configFile.TargetedOrganization.Name).To(BeEmpty())
				Expect(configFile.TargetedSpace.GUID).To(BeEmpty())
				Expect(configFile.TargetedSpace.Name).To(BeEmpty())
				Expect(configFile.TargetedSpace.AllowSSH).To(BeFalse())
				Expect(configFile.SkipSSLValidation).To(BeFalse())
			})
		})
	})

	Context("when Skip SSL Validation is required", func() {
		Context("api has SSL", func() {
			BeforeEach(func() {
				if skipSSLValidation == "" {
					Skip("SKIP_SSL_VALIDATION is not enabled")
				}
			})

			It("warns about skip SSL", func() {
				session := helpers.CF("api", apiURL)
				Eventually(session).Should(Say("Setting api endpoint to %s...", apiURL))
				Eventually(session.Err).Should(Say("x509: certificate has expired or is not yet valid|SSL Certificate Error x509: certificate is valid for|Invalid SSL Cert for %s", apiURL))
				Eventually(session.Err).Should(Say("TIP: Use 'cf api --skip-ssl-validation' to continue with an insecure API endpoint"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})

			It("sets the API endpoint", func() {
				session := helpers.CF("api", apiURL, "--skip-ssl-validation")
				Eventually(session).Should(Say("Setting api endpoint to %s...", apiURL))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("api endpoint:\\s+%s", apiURL))
				Eventually(session).Should(Say("api version:\\s+\\d+\\.\\d+\\.\\d+"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("api does not have SSL", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				serverAPIURL := server.URL()[7:]

				response := `{
					"name":"",
					"build":"",
					"support":"http://support.cloudfoundry.com",
					"version":0,
					"description":"",
					"authorization_endpoint":"https://login.APISERVER",
					"token_endpoint":"https://uaa.APISERVER",
					"min_cli_version":null,
					"min_recommended_cli_version":null,
					"api_version":"2.59.0",
					"app_ssh_endpoint":"ssh.APISERVER",
					"app_ssh_host_key_fingerprint":"a6:d1:08:0b:b0:cb:9b:5f:c4:ba:44:2a:97:26:19:8a",
					"app_ssh_oauth_client":"ssh-proxy",
					"logging_endpoint":"wss://loggregator.APISERVER",
					"doppler_logging_endpoint":"wss://doppler.APISERVER"
				}`
				response = strings.Replace(response, "APISERVER", serverAPIURL, -1)
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/info"),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)
			})

			AfterEach(func() {
				server.Close()
			})

			It("falls back to http and gives a warning", func() {
				session := helpers.CF("api", server.URL(), "--skip-ssl-validation")
				Eventually(session).Should(Say("Setting api endpoint to %s...", server.URL()))
				Eventually(session).Should(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(0))
			})
		})

		It("sets SSL Disabled in the config file to true", func() {
			command := exec.Command("cf", "api", apiURL, "--skip-ssl-validation")
			session, err := Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(Exit(0))

			rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
			Expect(err).NotTo(HaveOccurred())

			var configFile configv3.CFConfig
			err = json.Unmarshal(rawConfig, &configFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(configFile.SkipSSLValidation).To(BeTrue())
		})
	})

	Context("when skip-ssl-validation is not required", func() {
		BeforeEach(func() {
			if skipSSLValidation != "" {
				Skip("SKIP_SSL_VALIDATION is enabled")
			}
		})

		It("logs in without any warnings", func() {
			session := helpers.CF("api", apiURL)
			Eventually(session).Should(Say("Setting api endpoint to %s...", apiURL))
			Consistently(session).ShouldNot(Say("Warning: Insecure http API endpoint detected: secure https API endpoints are recommended"))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say("Not logged in. Use 'cf login' to log in."))
			Eventually(session).Should(Exit(0))
		})

		It("sets SSL Disabled in the config file to false", func() {
			session := helpers.CF("api", apiURL, skipSSLValidation)
			Eventually(session).Should(Exit(0))

			rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
			Expect(err).NotTo(HaveOccurred())

			var configFile configv3.CFConfig
			err = json.Unmarshal(rawConfig, &configFile)
			Expect(err).NotTo(HaveOccurred())

			Expect(configFile.SkipSSLValidation).To(BeTrue())
		})
	})

	It("sets the config file", func() {
		session := helpers.CF("api", apiURL, skipSSLValidation)
		Eventually(session).Should(Exit(0))

		rawConfig, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
		Expect(err).NotTo(HaveOccurred())

		var configFile configv3.CFConfig
		err = json.Unmarshal(rawConfig, &configFile)
		Expect(err).NotTo(HaveOccurred())

		Expect(configFile.ConfigVersion).To(Equal(3))
		Expect(configFile.Target).To(Equal(apiURL))
		Expect(configFile.APIVersion).To(MatchRegexp("\\d+\\.\\d+\\.\\d+"))
		Expect(configFile.AuthorizationEndpoint).ToNot(BeEmpty())
		Expect(configFile.DopplerEndpoint).To(MatchRegexp("^wss://"))
		Expect(configFile.RoutingEndpoint).NotTo(BeEmpty())
		Expect(configFile.UAAEndpoint).To(BeEmpty())
		Expect(configFile.AccessToken).To(BeEmpty())
		Expect(configFile.RefreshToken).To(BeEmpty())
		Expect(configFile.TargetedOrganization.GUID).To(BeEmpty())
		Expect(configFile.TargetedOrganization.Name).To(BeEmpty())
		Expect(configFile.TargetedSpace.GUID).To(BeEmpty())
		Expect(configFile.TargetedSpace.Name).To(BeEmpty())
		Expect(configFile.TargetedSpace.AllowSSH).To(BeFalse())
	})

	It("handles API endpoints with trailing slash", func() {
		session := helpers.CF("api", apiURL+"/", skipSSLValidation)
		Eventually(session).Should(Exit(0))

		helpers.LoginCF()

		session = helpers.CF("orgs")
		Eventually(session).Should(Exit(0))
	})
})
