package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("add-plugin-repo command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("add-plugin-repo", "--help", "-k")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("add-plugin-repo - Add a new plugin repository"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf add-plugin-repo REPO_NAME URL"))
				Eventually(session.Out).Should(Say("EXAMPLES"))
				Eventually(session.Out).Should(Say("cf add-plugin-repo ExampleRepo https://example\\.com/repo"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("install-plugin, list-plugin-repos"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the command line arguments are invalid", func() {
		Context("when no arguments are provided", func() {
			It("fails with incorrect usage message and displays help", func() {
				session := helpers.CF("add-plugin-repo", "-k")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `REPO_NAME` and `URL` were not provided"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when only one argument is provided", func() {
			It("fails with incorrect usage message and displays help", func() {
				session := helpers.CF("add-plugin-repo", "repo-name", "-k")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `URL` was not provided"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the user provides a url without a protocol scheme", func() {
		It("defaults to 'https://'", func() {
			session := helpers.CF("add-plugin-repo", "some-repo", "example.com/repo", "-k")

			Eventually(session.Err).Should(Say("Could not add repository 'some-repo' from https://example\\.com/repo:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the provided URL is a valid plugin repository", func() {
		var (
			server     *Server
			serverURL  string
			pluginRepo helpers.PluginRepository
		)

		BeforeEach(func() {
			pluginRepo = helpers.PluginRepository{
				Plugins: []helpers.Plugin{},
			}
			server = helpers.NewPluginRepositoryTLSServer(pluginRepo)
			serverURL = server.URL()
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds and exits 0", func() {
			session := helpers.CF("add-plugin-repo", "repo1", serverURL, "-k")

			Eventually(session.Out).Should(Say("%s added as repo1", serverURL))
			Eventually(session).Should(Exit(0))
		})

		Context("when the repo URL is already in use", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("add-plugin-repo", "repo1", serverURL, "-k")).Should(Exit(0))
			})

			It("allows the duplicate repo URL", func() {
				session := helpers.CF("add-plugin-repo", "some-repo", serverURL, "-k")

				Eventually(session.Out).Should(Say("%s added as some-repo", serverURL))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the repo name is already in use", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("add-plugin-repo", "repo1", serverURL, "-k")).Should(Exit(0))
			})

			Context("when the repo name is different only in case sensitivity", func() {
				It("succeeds and exists 0", func() {
					session := helpers.CF("add-plugin-repo", "rEPo1", serverURL, "-k")

					Eventually(session.Out).Should(Say("%s already registered as repo1", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the URL is different", func() {
				It("errors and says the repo name is taken", func() {
					session := helpers.CF("add-plugin-repo", "repo1", "some-other-url", "-k")

					Eventually(session.Err).Should(Say("Plugin repo named 'repo1' already exists, please use another name\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the URL is the same", func() {
				It("succeeds and exits 0", func() {
					session := helpers.CF("add-plugin-repo", "repo1", serverURL, "-k")

					Eventually(session.Out).Should(Say("%s already registered as repo1", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the URL is the same except for a trailing '/'", func() {
				It("succeeds and exits 0", func() {
					session := helpers.CF("add-plugin-repo", "repo1", fmt.Sprintf("%s/", serverURL), "-k")

					Eventually(session.Out).Should(Say("%s already registered as repo1", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the repo URL contains a path", func() {
			BeforeEach(func() {
				jsonBytes, err := json.Marshal(pluginRepo)
				Expect(err).ToNot(HaveOccurred())

				server.SetHandler(
					0,
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/some-path/list"),
						RespondWith(http.StatusOK, jsonBytes),
					),
				)
			})

			Context("when the repo URL ends with /list", func() {
				It("succeeds and exits 0", func() {
					session := helpers.CF("add-plugin-repo", "some-repo", fmt.Sprintf("%s/some-path/list", serverURL), "-k")

					Eventually(session.Out).Should(Say("%s/some-path/list added as some-repo", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the repo URL does not end with /list", func() {
				It("succeeds and exits 0", func() {
					session := helpers.CF("add-plugin-repo", "some-repo", fmt.Sprintf("%s/some-path", serverURL), "-k")

					Eventually(session.Out).Should(Say("%s/some-path added as some-repo", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the repo URL ends with trailing /", func() {
				It("succeeds and exits 0", func() {
					session := helpers.CF("add-plugin-repo", "some-repo", fmt.Sprintf("%s/some-path/", serverURL), "-k")

					Eventually(session.Out).Should(Say("%s/some-path/ added as some-repo", serverURL))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	Context("when the provided URL is NOT a valid plugin repository", func() {
		var server *Server

		BeforeEach(func() {
			server = NewTLSServer()
			// Suppresses ginkgo server logs
			server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when the protocol is unsupported", func() {
			It("reports an appropriate error", func() {
				session := helpers.CF("add-plugin-repo", "repo1", "ftp://example.com/repo", "-k")

				Eventually(session.Err).Should(Say("Could not add repository 'repo1' from ftp://example\\.com/repo: Get ftp://example\\.com/repo/list: unsupported protocol scheme \"ftp\""))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the domain cannot be reached", func() {
			It("reports an appropriate error", func() {
				session := helpers.CF("add-plugin-repo", "repo1", "cfpluginrepothatdoesnotexist.cf-app.com", "-k")

				Eventually(session.Err).Should(Say("Could not add repository 'repo1' from https://cfpluginrepothatdoesnotexist\\.cf-app\\.com: Get https://cfpluginrepothatdoesnotexist\\.cf-app\\.com/list: dial tcp: lookup cfpluginrepothatdoesnotexist\\.cf-app\\.com.*: no such host"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the path cannot be found", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					RespondWith(http.StatusNotFound, "foobar"),
				)
			})

			It("returns an appropriate error", func() {
				session := helpers.CF("add-plugin-repo", "repo1", server.URL(), "-k")

				Eventually(session.Err).Should(Say("Could not add repository 'repo1' from https://127\\.0\\.0\\.1:\\d{1,5}"))
				Eventually(session.Err).Should(Say("HTTP Response: 404"))
				Eventually(session.Err).Should(Say("HTTP Response Body: foobar"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the response is not parseable", func() {
			BeforeEach(func() {
				server.AppendHandlers(RespondWith(http.StatusOK, `{"plugins":[}`))
			})

			It("returns an appropriate error", func() {
				session := helpers.CF("add-plugin-repo", "repo1", server.URL(), "-k")

				Eventually(session.Err).Should(Say("Could not add repository 'repo1' from https://127\\.0\\.0\\.1:\\d{1,5}: invalid character '}' looking for beginning of value"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
