package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-route command", func() {
	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CF("create-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`create-route - Create a url route in a space for later use\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`Create an HTTP route:`))
			Eventually(session).Should(Say(`cf create-route SPACE DOMAIN \[--hostname HOSTNAME\] \[--path PATH\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`Create a TCP route:`))
			Eventually(session).Should(Say(`cf create-route SPACE DOMAIN \(--port PORT \| --random-port\)\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf create-route my-space example.com\s+# example.com`))
			Eventually(session).Should(Say(`cf create-route my-space example.com --hostname myapp\s+# myapp.example.com`))
			Eventually(session).Should(Say(`cf create-route my-space example.com --hostname myapp --path foo\s+# myapp.example.com/foo`))
			Eventually(session).Should(Say(`cf create-route my-space example.com --port 5000\s+# example.com:5000\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname for the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path for the HTTP route`))
			Eventually(session).Should(Say(`--port\s+Port for the TCP route`))
			Eventually(session).Should(Say(`--random-port\s+Create a random port for the TCP route\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`check-route, domains, map-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	Context("Flag Errors", func() {
		When("--hostname and --port are provided", func() {
			It("fails with a message about being unable to mix --port with the HTTP route options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--hostname", "some-host", "--port", "1122")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --hostname, --port`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("--hostname and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--hostname", "some-host", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --hostname, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("--path and --port are provided", func() {
			It("fails with a message about being unable to mix --port with the HTTP route options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--path", "/some-path", "--port", "1111")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --path, --port`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("--path and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--path", "/some-path", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --path, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("both --port and --random-port are provided", func() {
			It("fails with a message about being unable to mix --random-port with any other options", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--port", "1121", "--random-port")
				Eventually(session.Err).Should(Say(`Incorrect Usage: The following arguments cannot be used together: --port, --random-port`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the provided port is not valid / parseable", func() {
			It("fails with an appropriate error", func() {
				session := helpers.CF("create-route", "some-space", "some-domain", "--port", "ABC")
				Eventually(session.Err).Should(Say(`Incorrect Usage: invalid argument for flag '--port' \(expected int > 0\)`))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "create-route", "some-space", "some-domain")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the space does not exist", func() {
			It("displays 'space not found' and exits 1", func() {
				badSpaceName := fmt.Sprintf("%s-1", spaceName)
				session := helpers.CF("create-route", badSpaceName, "some-domain")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Space '%s' not found\.`, badSpaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `DOMAIN` were not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route", spaceName, "some-domain")
				Eventually(session).Should(Say(`FAILED`))
				Eventually(session.Err).Should(Say(`Domain some-domain not found`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("create-route", spaceName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space and domain exist", func() {
			var (
				userName   string
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.NewDomainName()
				userName, _ = helpers.GetCredentials()
			})

			When("the route already exists", func() {
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
					Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say(`Route %s already exists\.`, domainName))
					Eventually(session).Should(Say(`OK`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the route already exists in a different space", func() {
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
					Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("The app cannot be mapped to route %s because the route exists in a different space.", domainName))
					Eventually(session).Should(Say(`FAILED`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the route does not already exist", func() {
				When("the domain is private", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.Create()
						Eventually(helpers.CF("create-route", spaceName, domainName)).Should(Exit(0))
					})

					AfterEach(func() {
						domain.Delete()
					})

					When("no flags are used", func() {
						It("creates the route", func() {
							session := helpers.CF("create-route", spaceName, domainName)
							Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the path is provided but the hostname is not", func() {
						var path string

						BeforeEach(func() {
							path = helpers.PrefixedRandomName("path")
						})

						It("creates the route", func() {
							session := helpers.CF("create-route", spaceName, domainName, "--path", path)
							Eventually(session).Should(Say(`Creating route %s/%s for org %s / space %s as %s\.\.\.`, domainName, path, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the domain is a shared HTTP domain", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.CreateShared()
					})

					AfterEach(func() {
						domain.DeleteShared()
					})

					When("no flags are used", func() {
						When("the domain already has some routes", func() {
							var hostName string

							BeforeEach(func() {
								hostName = helpers.PrefixedRandomName("my-host")
								Eventually(helpers.CF("create-route", spaceName, domainName, "--hostname", hostName)).Should(Exit(0))
							})

							It("fails with error message informing users to provide a port or random-port", func() {
								session := helpers.CF("create-route", spaceName, domainName)
								Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
								Eventually(session).Should(Say(`FAILED`))
								Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
								Eventually(session).Should(Exit(1))
							})
						})

						It("fails with an error message and exits 1", func() {
							session := helpers.CF("create-route", spaceName, domainName)
							Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
							Eventually(session).Should(Exit(1))
						})
					})

					When("TCP flag options are provided", func() {
						It("fails with an error message and exits 1", func() {
							port := "90230"
							session := helpers.CF("create-route", spaceName, domainName, "--port", port)
							Eventually(session).Should(Say(`Creating route %s:%s for org %s / space %s as %s\.\.\.`, domainName, port, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`Port not allowed in HTTP domain %s`, domainName))
							Eventually(session).Should(Exit(1))
						})

						It("fails with an error message and exits 1", func() {
							session := helpers.CF("create-route", spaceName, domainName, "--random-port")
							Eventually(session).Should(Say(`Creating route %s for org %s / space %s as %s\.\.\.`, domainName, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say(`Port not allowed in HTTP domain %s`, domainName))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the hostname is provided", func() {
						var hostName string

						BeforeEach(func() {
							hostName = helpers.PrefixedRandomName("my-host")
						})

						When("no path is provided", func() {
							It("creates the route", func() {
								session := helpers.CF("create-route", spaceName, domainName, "--hostname", hostName)
								Eventually(session).Should(Say(`Creating route %s.%s for org %s / space %s as %s\.\.\.`, hostName, domainName, orgName, spaceName, userName))
								Eventually(session).Should(Exit(0))
							})
						})

						When("a path is provided", func() {
							It("creates the route", func() {
								path := fmt.Sprintf("/%s", helpers.PrefixedRandomName("path"))
								session := helpers.CF("create-route", spaceName, domainName, "--hostname", hostName, "--path", path)
								Eventually(session).Should(Say(`Creating route %s.%s%s for org %s / space %s as %s\.\.\.`, hostName, domainName, path, orgName, spaceName, userName))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("the hostname is not provided", func() {
						var path string

						BeforeEach(func() {
							path = helpers.PrefixedRandomName("path")
						})

						When("the path is provided", func() {
							It("fails with an error message and exits 1", func() {
								session := helpers.CF("create-route", spaceName, domainName, "--path", path)
								Eventually(session).Should(Say(`Creating route %s/%s for org %s / space %s as %s\.\.\.`, domainName, path, orgName, spaceName, userName))
								Eventually(session.Err).Should(Say(`The route is invalid: host is required for shared-domains`))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})
