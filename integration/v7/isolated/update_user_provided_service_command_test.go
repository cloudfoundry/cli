package isolated

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-user-provided-service command", func() {
	expectOKMessage := func(session *Session, serviceName, orgName, spaceName, userName string) {
		Expect(session.Out).To(Say("Updating user provided service %s in org %s / space %s as %s...", serviceName, orgName, spaceName, userName))
		Expect(session.Out).To(Say("OK"))
		Expect(session.Out).To(Say("TIP: Use 'cf restage' for any bound apps to ensure your env variable changes take effect"))
	}

	expectHelpMessage := func(session *Session) {
		Expect(session).To(Say(`NAME:`))
		Expect(session).To(Say(`\s+update-user-provided-service - Update user-provided service instance`))
		Expect(session).To(Say(`USAGE:`))
		Expect(session).To(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE \[-p CREDENTIALS\] \[-l SYSLOG_DRAIN_URL\] \[-r ROUTE_SERVICE_URL\] \[-t TAGS\]`))
		Expect(session).To(Say(`\s+Pass comma separated credential parameter names to enable interactive mode:`))
		Expect(session).To(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p "comma, separated, parameter, names"`))
		Expect(session).To(Say(`\s+Pass credential parameters as JSON to create a service non-interactively:`))
		Expect(session).To(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p '{"key1":"value1","key2":"value2"}'`))
		Expect(session).To(Say(`\s+Specify a path to a file containing JSON:`))
		Expect(session).To(Say(`\s+cf update-user-provided-service SERVICE_INSTANCE -p PATH_TO_FILE`))
		Expect(session).To(Say(`EXAMPLES:`))
		Expect(session).To(Say(`\s+cf update-user-provided-service my-db-mine -p '{"username":"admin", "password":"pa55woRD"}'`))
		Expect(session).To(Say(`\s+cf update-user-provided-service my-db-mine -p /path/to/credentials.json`))
		Expect(session).To(Say(`\s+cf create-user-provided-service my-db-mine -t "list, of, tags"`))
		Expect(session).To(Say(`\s+cf update-user-provided-service my-drain-service -l syslog://example.com`))
		Expect(session).To(Say(`\s+cf update-user-provided-service my-route-service -r https://example.com`))
		Expect(session).To(Say(`ALIAS:`))
		Expect(session).To(Say(`\s+uups`))
		Expect(session).To(Say(`OPTIONS:`))
		Expect(session).To(Say(`\s+-l\s+URL to which logs for bound applications will be streamed`))
		Expect(session).To(Say(`\s+-p\s+Credentials, provided inline or in a file, to be exposed in the VCAP_SERVICES environment variable for bound applications. Provided credentials will override existing credentials.`))
		Expect(session).To(Say(`\s+-r\s+URL to which requests for bound routes will be forwarded. Scheme for this URL must be https`))
		Expect(session).To(Say(`\s+-t\s+User provided tags`))
		Expect(session).To(Say(`SEE ALSO:`))
		Expect(session).To(Say(`\s+rename-service, services, update-service`))
	}

	randomUserProvidedServiceName := func() string {
		return helpers.PrefixedRandomName("ups")
	}

	createUserProvidedService := func(name string) {
		Eventually(helpers.CF("create-user-provided-service", name)).Should(Exit(0))
	}

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("update-user-provided-service", "--help")
				Eventually(session).Should(Exit(0))

				expectHelpMessage(session)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "update-user-provided-service", "foo")
		})
	})

	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			userName  string
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the user-provided service instance name is not provided", func() {
			It("displays the help message and exits 1", func() {
				session := helpers.CF("update-user-provided-service")
				Eventually(session).Should(Exit(1))

				expectHelpMessage(session)
			})
		})

		When("there are unknown additional arguments", func() {
			It("displays the help message and exits 1", func() {
				session := helpers.CF("update-user-provided-service", "service-name", "additional", "invalid", "arguments")
				Eventually(session).Should(Exit(1))

				expectHelpMessage(session)
			})
		})

		When("the user-provided service instance does not exist", func() {
			It("displays an informative error and exits 1", func() {
				session := helpers.CF("update-user-provided-service", "nonexistent-service", "-l", "syslog://example.com")
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say("Service instance 'nonexistent-service' not found"))
				Expect(session.Out).To(Say("FAILED"))
			})
		})

		When("the user-provided service exists", func() {
			var serviceName string

			BeforeEach(func() {
				serviceName = randomUserProvidedServiceName()
				createUserProvidedService(serviceName)
			})

			When("no flags are provided", func() {
				It("displays an informative message and exits 0", func() {
					session := helpers.CF("update-user-provided-service", serviceName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(Say("No flags specified. No changes were made."))
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

					Eventually(session).Should(Exit(0))
					expectOKMessage(session, serviceName, orgName, spaceName, userName)

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`tags:\s+tag1,\s*tag2`))
					Expect(session).To(Say(`route service url:\s+https://example.com`))
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
					Eventually(session).Should(Exit(0))

					Eventually(session).Should(Say(`tags:\s+tag1,\s*tag2`))
					Eventually(session).Should(Say(`route service url:\s+https://example.com`))
				})

				It("can unset previous values by providing empty strings as flag values", func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `""`,
						`-p`, `""`,
						`-r`, `""`,
						`-t`, `""`,
					)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Exit(0))

					Expect(session).NotTo(Say(`tags:\s+tag1,\s*tag2`))
					Expect(session).NotTo(Say(`route service url:\s+https://example.com`))
				})

				It("does not unset previous values for flags that are not provided", func() {
					session := helpers.CF(
						`update-user-provided-service`, serviceName,
						`-l`, `""`,
						`-p`, `""`,
					)
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service", serviceName)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`tags:\s+tag1,\s*tag2`))
					Expect(session).To(Say(`route service url:\s+https://example.com`))
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
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say("username: "))
					Expect(session).To(Say("password: "))
					Expect(session).NotTo(Say("fake-username"), "credentials should not be echoed to the user")
					Expect(session).NotTo(Say("fake-password"), "credentials should not be echoed to the user")
					expectOKMessage(session, serviceName, orgName, spaceName, userName)
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
					Eventually(session).Should(Exit(0))
					expectOKMessage(session, serviceName, orgName, spaceName, userName)
				})
			})
		})
	})
})
