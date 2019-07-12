package isolated

import (
	"os"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-user-provided-service command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		helpers.SkipIfVersionLessThan(ccversion.MinVersionTagsOnUserProvidedServices)
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("update-user-provided-service", "--help")
				eventuallyExpectHelpMessage(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "update-user-provided-service", "foo")
		})
	})

	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		const userName = "admin"

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

		When("the user-provided service instance name is not provided", func() {
			It("displays the help message and exits 1", func() {
				session := helpers.CF("update-user-provided-service")
				eventuallyExpectHelpMessage(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("there are unknown additional arguments", func() {
			It("displays the help message and exits 1", func() {
				session := helpers.CF("update-user-provided-service", "service-name", "additional", "invalid", "arguments")
				eventuallyExpectHelpMessage(session)
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user-provided service instance does not exist", func() {
			It("displays an informative error and exits 1", func() {
				session := helpers.CF("update-user-provided-service", "non-existent-service")
				Eventually(session.Err).Should(Say("Service instance non-existent-service not found"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user-provided service exists", func() {
			var serviceName string

			BeforeEach(func() {
				serviceName = randomUserProvidedServiceName()
				createUserProvidedService(serviceName)
			})

			AfterEach(func() {
				deleteUserProvidedService(serviceName)
			})

			When("no flags are provided", func() {
				It("displays an informative message and exits 0", func() {
					session := helpers.CF("update-user-provided-service", serviceName)
					Eventually(session.Out).Should(Say("No flags specified. No changes were made."))
					Eventually(session).Should(Exit(0))
				})
			})

			When("flags are provided", func() {
				It("displays success message, exits 0, and updates the service", func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `syslog://example.com`,
						`-p`, `{"some": "credentials"}`,
						`-r`, `https://example.com`,
						`-t`, `"tag1,tag2"`,
					)

					eventuallyExpectOKMessage(session, serviceName, orgName, spaceName, userName)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Say(`tags:\s+tag1,\s*tag2`))
					Eventually(session).Should(Say(`route service url:\s+https://example.com`))
				})
			})

			Context("when the user-provided service already has values", func() {
				BeforeEach(func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `syslog://example.com`,
						`-p`, `{"some": "credentials"}`,
						`-r`, `https://example.com`,
						`-t`, `"tag1,tag2"`,
					)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Say(`tags:\s+tag1,\s*tag2`))
					Eventually(session).Should(Say(`route service url:\s+https://example.com`))
				})

				It("can unset previous values provideding empty strings as flag values", func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `""`,
						`-p`, `""`,
						`-r`, `""`,
						`-t`, `""`,
					)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Consistently(session).ShouldNot(Say(`tags:\s+tag1,\s*tag2`))
					Consistently(session).ShouldNot(Say(`route service url:\s+https://example.com`))
				})

				It("does not unset previous values for flags that are not provided", func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `""`,
						`-p`, `""`,
					)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Say(`tags:\s+tag1,\s*tag2`))
					Eventually(session).Should(Say(`route service url:\s+https://example.com`))
				})
			})

			When("requesting interactive credentials", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("fake-username\nfake-password\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("requests the credentials at a prompt", func() {
					session := helpers.CFWithStdin(buffer, "update-user-provided-service", serviceName, "-p", `"username,password"`)

					Eventually(session).Should(Say("username: "))
					Eventually(session).Should(Say("password: "))
					Consistently(session).ShouldNot(Say("fake-username"), "credentials should not be echoed to the user")
					Consistently(session).ShouldNot(Say("fake-password"), "credentials should not be echoed to the user")
					eventuallyExpectOKMessage(session, serviceName, orgName, spaceName, userName)
					Eventually(session).Should(Exit(0))
				})
			})

			When("reading JSON credentials from a file", func() {
				var path string

				BeforeEach(func() {
					path = helpers.TempFileWithContent(`{"some": "credentials"}`)
				})

				AfterEach(func() {
					Expect(os.Remove(path)).To(Succeed())
				})

				It("accepts a file path", func() {
					session := helpers.CF("update-user-provided-service", serviceName, "-p", path)

					By("checking that it does not interpret the file name as request for an interactive credential prompt")
					Consistently(session.Out.Contents()).ShouldNot(ContainSubstring(path))

					By("succeeding")
					eventuallyExpectOKMessage(session, serviceName, orgName, spaceName, userName)
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})

func eventuallyExpectHelpMessage(session *Session) {
	Eventually(session).Should(Say(`NAME:`))
	Eventually(session).Should(Say(`\s+update-user-provided-service - Update user-provided service instance`))
	Eventually(session).Should(Say(`USAGE:`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE \[-p CREDENTIALS\] \[-l SYSLOG_DRAIN_URL\] \[-r ROUTE_SERVICE_URL\] \[-t TAGS\]`))
	Eventually(session).Should(Say(`\s+Pass comma separated credential parameter names to enable interactive mode:`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"`))
	Eventually(session).Should(Say(`\s+Pass credential parameters as JSON to create a service non-interactively:`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p '{"key1":"value1","key2":"value2"}'`))
	Eventually(session).Should(Say(`\s+Specify a path to a file containing JSON:`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`))
	Eventually(session).Should(Say(`EXAMPLES:`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service my-db-mine -p '{"username":"admin", "password":"pa55woRD"}'`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service my-db-mine -p /path/to/credentials.json`))
	Eventually(session).Should(Say(`\s+cf create-user-provided-service my-db-mine -t "list, of, tags"`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service my-drain-service -l syslog://example.com`))
	Eventually(session).Should(Say(`\s+cf update-user-provided-service my-route-service -r https://example.com`))
	Eventually(session).Should(Say(`ALIAS:`))
	Eventually(session).Should(Say(`\s+uups`))
	Eventually(session).Should(Say(`OPTIONS:`))
	Eventually(session).Should(Say(`\s+-l\s+URL to which logs for bound applications will be streamed`))
	Eventually(session).Should(Say(`\s+-p\s+Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications. Provided credentials will override existing credentials.`))
	Eventually(session).Should(Say(`\s+-r\s+URL to which requests for bound routes will be forwarded. Scheme for this URL must be https`))
	Eventually(session).Should(Say(`\s+-t\s+User provided tags`))
	Eventually(session).Should(Say(`SEE ALSO:`))
	Eventually(session).Should(Say(`\s+rename-service, services, update-service`))
}

func eventuallyExpectOKMessage(session *Session, serviceName, orgName, spaceName, userName string) {
	Eventually(session.Out).Should(Say("Updating user provided service %s in org %s / space %s as %s...", serviceName, orgName, spaceName, userName))
	Eventually(session.Out).Should(Say("OK"))
	Eventually(session.Out).Should(Say("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect"))
}

func randomUserProvidedServiceName() string {
	return helpers.PrefixedRandomName("ups")
}

func createUserProvidedService(name string) {
	Eventually(helpers.CF("create-user-provided-service", name)).Should(Exit(0))
}

func deleteUserProvidedService(name string) {
	Eventually(helpers.CF("delete-service", name)).Should(Exit(0))
}
