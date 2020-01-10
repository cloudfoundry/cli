package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-route command", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("delete-route", "ROUTES", "Delete a route"))
		})

		It("displays the help information", func() {
			session := helpers.CF("delete-route", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`delete-route - Delete a route\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf delete-route DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] \[-f\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf delete-route example.com\s+# example.com`))
			Eventually(session).Should(Say(`cf delete-route example.com --hostname myhost\s+# myhost.example.com`))
			Eventually(session).Should(Say(`cf delete-route example.com --hostname myhost --path foo\s+# myhost.example.com/foo`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`-f\s+Force deletion without confirmation`))
			Eventually(session).Should(Say(`--hostname, -n\s+Hostname used to identify the HTTP route \(required for shared domains\)`))
			Eventually(session).Should(Say(`--path\s+Path used to identify the HTTP route`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`delete-orphaned-routes, routes, unmap-route`))

			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "delete-route", "some-domain")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			buffer     *Buffer
			orgName    string
			spaceName  string
			domainName string
		)

		BeforeEach(func() {
			buffer = NewBuffer()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			domainName = helpers.NewDomainName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})
		When("the -f flag is not given", func() {
			var (
				domain helpers.Domain
			)

			BeforeEach(func() {
				domain = helpers.NewDomain(orgName, domainName)
				domain.CreatePrivate()
			})

			AfterEach(func() {
				domain.Delete()
			})

			When("the user enters 'y'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())

					Eventually(helpers.CF("create-route", domainName)).Should(Exit(0))
				})

				When("the user attempts to delete a route with a private domain", func() {
					It("it asks for confirmation and deletes the domain", func() {
						session := helpers.CFWithStdin(buffer, "delete-route", domainName)
						Eventually(session).Should(Say("This action impacts all apps using this route."))
						Eventually(session).Should(Say("Deleting this route will make apps unreachable via this route."))
						Eventually(session).Should(Say(`Really delete the route %s\?`, domainName))
						Eventually(session).Should(Say(regexp.QuoteMeta(`Deleting route %s...`), domainName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))
					})
				})
			})

			When("the user enters 'n'", func() {
				BeforeEach(func() {
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("it asks for confirmation and does not delete the domain", func() {
					session := helpers.CFWithStdin(buffer, "delete-route", domainName)
					Eventually(session).Should(Say("This action impacts all apps using this route."))
					Eventually(session).Should(Say("Deleting this route will make apps unreachable via this route."))
					Eventually(session).Should(Say(`Really delete the route %s\?`, domainName))
					Eventually(session).Should(Say(`'%s' has not been deleted`, domainName))
					Consistently(session).ShouldNot(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the -f flag is given", func() {
			BeforeEach(func() {
				domainName = helpers.NewDomainName()
			})

			When("the route does not exist", func() {
				var (
					domain helpers.Domain
				)

				BeforeEach(func() {
					domain = helpers.NewDomain(orgName, domainName)
					domain.CreatePrivate()
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("warns the user that it has already been deleted and runs to completion without failing", func() {
					session := helpers.CF("delete-route", domainName, "-f")
					Eventually(session).Should(Say(`Deleting route %s\.\.\.`, domainName))
					Eventually(session).Should(Say(`Unable to delete\. Route with domain '%s' not found\.`, domainName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the route exist", func() {
				When("the domain is private", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain(orgName, domainName)
						domain.Create()
					})

					AfterEach(func() {
						domain.Delete()
					})

					When("no flags are used", func() {
						It("deletes the route", func() {
							Eventually(helpers.CF("create-route", domainName)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.\.\.`, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s`, domainName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in a hostname", func() {
						It("deletes the route with the hostname", func() {
							hostname := "tiramisu"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\.\.\.`, hostname, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s`, hostname, domainName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in a path", func() {
						It("deletes the route with the path", func() {
							path := "/flan"
							Eventually(helpers.CF("create-route", domainName, "--path", path)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "--path", path, "-f")
							Eventually(session).Should(Say(`Deleting route %s%s\.\.\.`, domainName, path))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s`, domainName, path))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in hostname and path with a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "/recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s\s+%s`, hostname, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in hostname and path without a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "more-recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\/%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s\s+\/%s`, hostname, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in empty hostname and path", func() {
						It("deletes the route with hostname and path", func() {
							hostname := ""
							pathString := "/recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s%s\.\.\.`, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s`, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in path without specifying a hostname", func() {
						It("deletes the route with hostname and path", func() {
							pathString := "/recipes"
							Eventually(helpers.CF("create-route", domainName, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s%s\.\.\.`, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))

							session = helpers.CF("routes")
							Consistently(session).ShouldNot(Say(`%s\s+%s`, domainName, pathString))
							Eventually(session).Should(Exit(0))
						})
					})

				})

				When("the domain is shared", func() {
					var domain helpers.Domain

					BeforeEach(func() {
						domain = helpers.NewDomain("", domainName)
						domain.CreateShared()
					})

					AfterEach(func() {
						domain.DeleteShared()
					})

					When("no flags are used", func() {
						It("fails with a helpful message", func() {
							session := helpers.CF("delete-route", domainName, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.\.\.`, domainName))
							Eventually(session).Should(Say(`Unable to delete\. Route with domain '%s' not found\.`, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("passing in a hostname", func() {
						It("deletes the route with the hostname", func() {
							hostname := "tiramisu"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\.\.\.`, hostname, domainName))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))
						})
					})

					When("passing in hostname and path with a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "/recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))
						})
					})

					When("passing in hostname and path without a leading '/'", func() {
						It("deletes the route with hostname and path", func() {
							hostname := "tiramisu"
							pathString := "more-recipes"
							Eventually(helpers.CF("create-route", domainName, "-n", hostname, "--path", pathString)).Should(Exit(0))

							session := helpers.CF("delete-route", domainName, "-n", hostname, "--path", pathString, "-f")
							Eventually(session).Should(Say(`Deleting route %s\.%s\/%s\.\.\.`, hostname, domainName, pathString))
							Eventually(session).Should(Say(`OK`))
							Eventually(session).Should(Exit(0))

							Expect(string(session.Out.Contents())).NotTo(ContainSubstring("Unable to delete"))
						})
					})
				})
			})
		})

		When("the domain does not exist", func() {
			It("displays error and exits 0", func() {
				session := helpers.CF("delete-route", "some-domain", "-f")
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
				Expect(string(session.Err.Contents())).To(Equal("Domain 'some-domain' not found.\n"))
			})
		})

		When("the domain is not specified", func() {
			It("displays error and exits 1", func() {
				session := helpers.CF("delete-route")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `DOMAIN` was not provided\n"))
				Eventually(session.Err).Should(Say("\n"))
				Eventually(session).Should(Say("NAME:\n"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
