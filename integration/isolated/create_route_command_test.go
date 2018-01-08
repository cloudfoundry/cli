package isolated

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("create-route command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-route", "--help")
			Eventually(session.Out).Should(Say(`NAME:`))
			Eventually(session.Out).Should(Say(`create-route - Create a url route in a space for later use\n`))
			Eventually(session.Out).Should(Say(`\n`))

			Eventually(session.Out).Should(Say(`USAGE:`))
			Eventually(session.Out).Should(Say(`Create an HTTP route:`))
			Eventually(session.Out).Should(Say(`cf create-route SPACE DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session.Out).Should(Say(`\n`))

			Eventually(session.Out).Should(Say(`Create a TCP route:`))
			Eventually(session.Out).Should(Say(`cf create-route SPACE DOMAIN \(--port PORT \| --random-port\)\n`))
			Eventually(session.Out).Should(Say(`\n`))

			Eventually(session.Out).Should(Say(`EXAMPLES:`))
			Eventually(session.Out).Should(Say(`cf create-route my-space example.com\s+# example.com`))
			Eventually(session.Out).Should(Say(`cf create-route my-space example.com --hostname myapp\s+# myapp.example.com`))
			Eventually(session.Out).Should(Say(`cf create-route my-space example.com --hostname myapp --path foo\s+# myapp.example.com/foo`))
			Eventually(session.Out).Should(Say(`cf create-route my-space example.com --port 5000\s+# example.com:5000\n`))
			Eventually(session.Out).Should(Say(`\n`))

			Eventually(session.Out).Should(Say(`OPTIONS:`))
			Eventually(session.Out).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session.Out).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session.Out).Should(Say(`--port\s+Port for the TCP route`))
			Eventually(session.Out).Should(Say(`--random-port\s+Create a random port for the TCP route\n`))
			Eventually(session.Out).Should(Say(`\n`))

			Eventually(session.Out).Should(Say(`SEE ALSO:`))
			Eventually(session.Out).Should(Say(`check-route, domains, map-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	Context("Flag Errors", func() {
		Context("when --hostname and --port are provided", func() {
			It("fails with a message about being unable to mix --port with the HTTP route options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--hostname", "some-host", "--port", "1122")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --hostname, --port`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when --hostname and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--hostname", "some-host", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --hostname, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when --path and --port are provided", func() {
			It("fails with a message about being unable to mix --port with the HTTP route options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--path", "/some-path", "--port", "1111")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --path, --port`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when --path and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--path", "/some-path", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --path, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when both --port and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--port", "1121", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --port, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the provided port is not valid / parseable", func() {
			It("fails with an appropriate error", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--port", "ABC")
				Eventually(session.Err).Should(Say(`Incorrect Usage: invalid argument for flag '--port' \(expected int > 0\)`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("create-route", "some-space", "some-domain")
				Eventually(session.Out).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("create-route", "some-space", "some-domain")
				Eventually(session.Out).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Not logged in\. Use 'cf login' to log in\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when no organization is targeted", func() {
			BeforeEach(func() {
				helpers.ClearTarget()
			})

			It("fails with 'no organization targeted' message and exits 1", func() {
				session := helpers.CF("create-route", "some-space", "some-domain")
				Eventually(session.Out).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org\.`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the server's API version is too low", func() {
		var server *Server

		BeforeEach(func() {
			server = NewTLSServer()
			server.AppendHandlers(
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, `{"api_version":"2.34.0"}`),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/v2/info"),
					RespondWith(http.StatusOK, fmt.Sprintf(`{"api_version":"2.34.0", "authorization_endpoint": "%s"}`, server.URL())),
				),
				CombineHandlers(
					VerifyRequest(http.MethodGet, "/login"),
					RespondWith(http.StatusOK, `{}`),
				),
			)
			Eventually(helpers.CF("api", server.URL(), "--skip-ssl-validation")).Should(Exit(0))
		})

		AfterEach(func() {
			server.Close()
		})

		Context("for HTTP routes", func() {
			Context("when specifying --path", func() {
				It("reports an error with a minimum-version message", func() {
					session := helpers.CF("create-route", "some-space", "example.com", "--path", "/foo")
					Eventually(session.Out).Should(Say(`FAILED`))
					Eventually(session.Err).Should(Say(`Option '--path' requires CF API version 2\.36\.0 or higher. Your target is 2\.34\.0\.`))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("for TCP routes", func() {
			Context("when specifying --port", func() {
				It("reports an error with a minimum-version message", func() {
					session := helpers.CF("create-route", "some-space", "example.com", "--port", "1025")
					Eventually(session.Out).Should(Say(`FAILED`))
					Eventually(session.Err).Should(Say(`Option '--port' requires CF API version 2\.53\.0 or higher\. Your target is 2\.34\.0\.`))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when specifying --random-port", func() {
				It("reports an error with a minimum-version message", func() {
					session := helpers.CF("create-route", "some-space", "example.com", "--random-port")
					Eventually(session.Out).Should(Say(`FAILED`))
					Eventually(session.Err).Should(Say(`Option '--random-port' requires CF API version 2\.53\.0 or higher\. Your target is 2\.34\.0\.`))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the space does not exist", func() {
			It("displays 'space not found' and exits 1", func() {
				badSpaceName := fmt.Sprintf("%s-1", spaceName)
				session := helpers.CF("create-route", badSpaceName, "some-domain")
				Eventually(session.Out).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Space '%s' not found\.`, badSpaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `DOMAIN` were not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session.Out).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route", spaceName, "some-domain")
				Eventually(session.Out).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain some-domain not found`))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route", spaceName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session.Out).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space and domain exist", func() {
			var (
				userName   string
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.DomainName()
				userName, _ = helpers.GetCredentials()
			})

			Context("when the route already exists", func() {
				var domain helpers.Domain

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					domain.Create()
					Eventually(helpers.CF("create-route", spaceName, domainName)).Should(Exit(0))
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("warns the user that it has already been created and runs to completion without failing", func() {
					session := helpers.CF("create-route", spaceName, domainName)
					Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say(`Route %s already exists\.`, domainName))
					Eventually(session.Out).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the route already exists in a different space", func() {
				var domain helpers.Domain

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					domain.Create()
					differentSpaceName := helpers.NewSpaceName()
					helpers.CreateSpace(differentSpaceName)
					Eventually(helpers.CF("create-route", differentSpaceName, domainName)).Should(Exit(0))
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("warns the user that the route is already in use and then fails", func() {
					session := helpers.CF("create-route", spaceName, domainName)
					Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("The app cannot be mapped to route %s because the route is not in this space. Apps must be mapped to routes in the same space.", domainName))
					Eventually(session.Out).Should(Say(`FAILED`))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the route does not already exist", func() {
				Context("when the domain is private", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.Create()
						Eventually(helpers.CF("create-route", spaceName, domainName)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					Context("when no flags are used", func() {
						It("creates the route", func() {
							session := helpers.CF("create-route", spaceName, domainName)
							Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the path is provided but the hostname is not", func() {
						var path string

						BeforeEach(func() {
							path = helpers.PrefixedRandomName("path")
						})

						It("creates the route", func() {
							session := helpers.CF("create-route", spaceName, domainName, "--path", path)
							Eventually(session.Out).Should(Say(`Creating route %s/%s for org %s / space %s as %s\.\.\.`, domainName, path, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the domain is a shared HTTP domain", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.CreateShared()
					})

					AfterEach(func() {
						domain.DeleteShared()
					})

					Context("when no flags are used", func() {
						Context("when the domain already has some routes", func() {
							var hostName string

							BeforeEach(func() {
								hostName = helpers.PrefixedRandomName("my-host")
								Eventually(helpers.CF("create-route", spaceName, domainName, "--hostname", hostName)).Should(Exit(0))
							})

							It("fails with error message informing users to provide a port or random-port", func() {
								session := helpers.CF("create-route", spaceName, domainName)
								Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say(`FAILED`))
								Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
								Eventually(session).Should(Exit(1))
							})
						})

						It("fails with an error message and exits 1", func() {
							session := helpers.CF("create-route", spaceName, domainName)
							Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("when TCP flag options are provided", func() {
						It("fails with an error message and exits 1", func() {
							port := "90230"
							session := helpers.CF("create-route", spaceName, domainName, "--port", port)
							Eventually(session.Out).Should(Say(`Creating route %s:%s for org %s / space %s as %s\.\.\.`, domainName, port, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`Port not allowed in HTTP domain %s`, domainName))
							Eventually(session).Should(Exit(1))
						})

						It("fails with an error message and exits 1", func() {
							session := helpers.CF("create-route", spaceName, domainName, "--random-port")
							Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`Port not allowed in HTTP domain %s`, domainName))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the hostname is provided", func() {
						var hostName string

						BeforeEach(func() {
							hostName = helpers.PrefixedRandomName("my-host")
						})

						Context("when no path is provided", func() {
							It("creates the route", func() {
								session := helpers.CF("create-route", spaceName, domainName, "--hostname", hostName)
								Eventually(session.Out).Should(Say(`Creating route %s.%s for org %s / space %s as %s\.\.\.`, hostName, domainName, orgName, spaceName, userName))
								Eventually(session).Should(Exit(0))
							})
						})

						Context("when a path is provided", func() {
							It("creates the route", func() {
								path := fmt.Sprintf("/%s", helpers.PrefixedRandomName("path"))
								session := helpers.CF("create-route", spaceName, domainName, "--hostname", hostName, "--path", path)
								Eventually(session.Out).Should(Say(`Creating route %s.%s%s for org %s / space %s as %s\.\.\.`, hostName, domainName, path, orgName, spaceName, userName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the hostname is not provided", func() {
						var path string

						BeforeEach(func() {
							path = helpers.PrefixedRandomName("path")
						})

						Context("when the path is provided", func() {
							It("fails with an error message and exits 1", func() {
								session := helpers.CF("create-route", spaceName, domainName, "-v", "--path", path)
								Eventually(session.Out).Should(Say(`Creating route %s/%s for org %s / space %s as %s\.\.\.`, domainName, path, orgName, spaceName, userName))
								Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})

				Context("when the domain is a shared TCP domain", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
					})

					AfterEach(func() {
						domain.DeleteShared()
					})

					Context("when HTTP flag options are provided", func() {
						It("fails with an error message and exits 1", func() {
							hostName := helpers.PrefixedRandomName("host-")
							path := helpers.PrefixedRandomName("path-")
							session := helpers.CF("create-route", spaceName, domainName, "--hostname", hostName, "--path", path)
							Eventually(session.Out).Should(Say(`Creating route %s.%s/%s for org %s / space %s as %s\.\.\.`, hostName, domainName, path, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`The route is invalid: For TCP routes you must specify a port or request a random one.`))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("when a port is provided", func() {
						It("creates the route", func() {
							port := "1025"
							session := helpers.CF("create-route", spaceName, domainName, "--port", port)
							Eventually(session.Out).Should(Say(`Creating route %s:%s for org %s / space %s as %s\.\.\.`, domainName, port, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})

					Context("when --random-port is provided", func() {
						It("creates the route", func() {
							session := helpers.CF("create-route", spaceName, domainName, "--random-port")
							Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Out).Should(Say(`Route %s:\d+ has been created\.`, domainName))
							Eventually(session).Should(Exit(0))
						})

						Context("when there are other routes in the domain we want to create the route with", func() {
							BeforeEach(func() {
								session := helpers.CF("create-route", spaceName, domainName, "--random-port")
								Eventually(session.Out).Should(Say(`Route %s:\d+ has been created\.`, domainName))
								Eventually(session).Should(Exit(0))
							})

							It("should determine that the random route does not already exist, and create it", func() {
								session := helpers.CF("create-route", spaceName, domainName, "--random-port")
								Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say(`Route %s:\d+ has been created\.`, domainName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when no options are provided", func() {
						It("fails with error message informing users to provide a port or random-port", func() {
							session := helpers.CF("create-route", spaceName, domainName)
							Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Out).Should(Say(`FAILED`))
							Eventually(session.Err).Should(Say(`The route is invalid: For TCP routes you must specify a port or request a random one.`))
							Eventually(session).Should(Exit(1))
						})

						Context("when the domain already has some routes", func() {
							BeforeEach(func() {
								Eventually(helpers.CF("create-route", spaceName, domainName, "--random-port")).Should(Exit(0))
							})

							It("fails with error message informing users to provide a port or random-port", func() {
								session := helpers.CF("create-route", spaceName, domainName)
								Eventually(session.Out).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say(`FAILED`))
								Eventually(session.Err).Should(Say(`The route is invalid: For TCP routes you must specify a port or request a random one.`))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})
