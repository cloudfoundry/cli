package plugin

import (
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("repo-plugins command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("repo-plugins", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("repo-plugins - List all available plugins in specified repository or in all added repositories"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf repo-plugins [-r REPO_NAME]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-r\\s+Name of a registered repository"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("add-plugin-repo, remove-plugin-repo, install-plugin"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when no repo exists", func() {
		Context("and no repo is specified", func() {
			It("should only output flavor text", func() {
				session := helpers.CF("repo-plugins")
				// TODO: enforce "only"
				Eventually(session).Should(Say("Getting plugins from all repositories ..."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("and a repo is specified", func() {
			It("errors and tells the user that the repo does not exist and gives a tip", func() {
				session := helpers.CF("repo-plugins", "-r", "nonexistent_repo")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("nonexistent_repo does not exist as an available plugin repo."))
				Eventually(session).Should(Say("Tip: use `add-plugin-repo` command to add repos."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when repos exist", func() {
		var server *Server

		BeforeEach(func() {
			server = NewServer()

			response := `{
					"plugins": [
						{
							"name": "plugin-1",
							"description": "useful plugin for useful things",
							"version": "1.0.0",
							"binaries": [{"platform":"osx","url":"http://some-url","checksum":"somechecksum"},{"platform":"win64","url":"http://another-url","checksum":"anotherchecksum"},{"platform":"linux64","url":"http://last-url","checksum":"lastchecksum"}]
						},
						{
							"name": "plugin-2",
							"description": "amazing plugin",
							"version": "2.0.0"
						}
					]
				}`
			server.RouteToHandler(http.MethodGet, "/list", func(w http.ResponseWriter, req *http.Request) {
				_, err := w.Write([]byte(response))
				Expect(err).ToNot(HaveOccurred())
			})

			session := helpers.CF("add-plugin-repo", "workingrepo1", server.URL())
			Eventually(session).Should(Exit(0))
			session = helpers.CF("add-plugin-repo", "workingrepo2", server.URL())
			Eventually(session).Should(Exit(0))
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when a repo connection doesn't fail", func() {
			Context("when a repo returns a valid response", func() {
				Context("and no repo is specified", func() {
					It("lists the repos with their respective plugins", func() {
						// TODO: fix spacing ( ...)
						session := helpers.CF("repo-plugins")
						Eventually(session).Should(Say("Getting plugins from all repositories ..."))

						Eventually(session).Should(Say("Repository: workingrepo1"))
						Eventually(session).Should(Say("name\\s+version\\s+description"))
						Eventually(session).Should(Say("plugin-1\\s+1.0.0\\s+useful plugin for useful things"))
						Eventually(session).Should(Say("plugin-2\\s+2.0.0\\s+amazing plugin"))

						Eventually(session).Should(Say("Repository: workingrepo2"))
						Eventually(session).Should(Say("name\\s+version\\s+description"))
						Eventually(session).Should(Say("plugin-1\\s+1.0.0\\s+useful plugin for useful things"))
						Eventually(session).Should(Say("plugin-2\\s+2.0.0\\s+amazing plugin"))

						Eventually(session).Should(Exit(0))
					})
				})

				Context("and a repo is specified", func() {
					It("lists the repo name and its relevant plugins", func() {
						session := helpers.CF("repo-plugins", "-r", "workingrepo1")
						Eventually(session).Should(Say("Getting plugins from repository 'workingrepo1'"))
						Eventually(session).Should(Say("Repository: workingrepo1"))
						Eventually(session).Should(Say("name\\s+version\\s+description"))
						Eventually(session).Should(Say("plugin-1\\s+1.0.0\\s+useful plugin for useful things"))
						Eventually(session).Should(Say("plugin-2\\s+2.0.0\\s+amazing plugin"))
						Consistently(session).ShouldNot(Say("workingrepo2"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the repo returns invalid response", func() {
				var (
					invalidServer   *Server
					invalidResponse string
				)

				BeforeEach(func() {
					invalidServer = NewServer()
					first := true

					invalidServer.RouteToHandler(http.MethodGet, "/list",
						func(w http.ResponseWriter, req *http.Request) {
							if first {
								first = false
								_, err := w.Write([]byte(`{}`))
								Expect(err).ToNot(HaveOccurred())
							} else {
								_, err := w.Write([]byte(invalidResponse))
								Expect(err).ToNot(HaveOccurred())
							}
						})

					session := helpers.CF("add-plugin-repo", "invalidrepo", invalidServer.URL())
					Eventually(session).Should(Exit(0))
				})

				AfterEach(func() {
					invalidServer.Close()
				})

				Context("when a repo returns an invalid json response", func() {
					BeforeEach(func() {
						invalidResponse = `{"blah": "blahblah"}`
					})

					Context("and no repo is specified", func() {
						It("should log an error for this repo at the end of output", func() {
							session := helpers.CF("repo-plugins")
							// TODO: fix spacing ( ...)
							Eventually(session).Should(Say("Getting plugins from all repositories ..."))
							Eventually(session).Should(Say("Repository: workingrepo1"))
							Eventually(session).Should(Say("Repository: workingrepo2"))
							// TODO: stderr for new code
							Eventually(session).Should(Say("Invalid data from 'invalidrepo' - plugin data does not exist"))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("and a repo is specified", func() {
						It("should log an error for this repo", func() {
							session := helpers.CF("repo-plugins", "-r", "invalidrepo")
							// TODO: fix spacing ( ...)
							Eventually(session).Should(Say("Getting plugins from repository 'invalidrepo' ..."))
							// TODO: stderr for new code
							Eventually(session).Should(Say("Invalid data from 'invalidrepo' - plugin data does not exist"))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when a repo returns an invalid html response", func() {
					BeforeEach(func() {
						invalidResponse = `<html><body>404 not found</body></html>`
					})

					Context("and no repo is specified", func() {
						It("should log an error for this repo at the end of output", func() {
							session := helpers.CF("repo-plugins")
							// TODO: fix spacing ( ...)
							Eventually(session).Should(Say("Getting plugins from all repositories ..."))
							Eventually(session).Should(Say("Repository: workingrepo1"))
							Eventually(session).Should(Say("Repository: workingrepo2"))
							// TODO: stderr for new code
							Eventually(session).Should(Say("Invalid json data from 'invalidrepo' - invalid character '<' looking for beginning of value"))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("and a repo is specified", func() {
						It("should log an error for this repo at the end of output", func() {
							session := helpers.CF("repo-plugins", "-r", "invalidrepo")
							// TODO: fix spacing ( ...)
							Eventually(session).Should(Say("Getting plugins from repository 'invalidrepo' ..."))
							// TODO: stderr for new code
							Eventually(session).Should(Say("Invalid json data from 'invalidrepo' - invalid character '<' looking for beginning of value"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})

		Context("when a repo connection fails", func() {
			BeforeEach(func() {
				server.Close()
			})

			Context("and no repo is specified", func() {
				It("should log an error for this repo at the end of output", func() {
					session := helpers.CF("repo-plugins")
					Eventually(session).Should(Say("Error requesting from 'workingrepo1' - Get .* getsockopt: connection refused"))
					Eventually(session).Should(Say("Error requesting from 'workingrepo2' - Get .* getsockopt: connection refused"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("and a repo is specified", func() {
				It("should log an error for this repo at the end of output", func() {
					session := helpers.CF("repo-plugins", "-r", "workingrepo1")
					Eventually(session).Should(Say("Error requesting from 'workingrepo1' - Get .* getsockopt: connection refused"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
