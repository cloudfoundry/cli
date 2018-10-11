package push

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with hostname", func() {
	var (
		appName string
		route   string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	When("the default domain is a HTTP domain", func() {
		When("no host is provided / host defaults to app name", func() {
			BeforeEach(func() {
				route = fmt.Sprintf("%s.%s", strings.ToLower(appName), helpers.DefaultSharedDomain())
			})

			When("the default route does not exist", func() {
				It("creates and maps the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("routes:\\s+%s", route))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("the default route exists and is unmapped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-route", space, helpers.DefaultSharedDomain(), "-n", strings.ToLower(appName))).Should(Exit(0))
				})

				It("maps the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("routes:\\s+%s", route))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the default route is mapped to the application", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")).Should(Exit(0))
					})
				})

				It("does nothing", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					appGUID := helpers.AppGUID(appName)
					session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s/routes", appGUID))
					Eventually(session).Should(Say(`"total_results":\s+1,`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the host is provided", func() {
			var hostname string

			BeforeEach(func() {
				hostname = strings.ToLower(helpers.NewAppName())
				route = fmt.Sprintf("%s.%s", hostname, helpers.DefaultSharedDomain())
			})

			When("the default route does not exist", func() {
				It("creates and maps the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--hostname", hostname, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("routes:\\s+%s", route))
					Eventually(session).Should(Exit(0))

					By("does not remap default route after")
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Consistently(session).ShouldNot(Say("(?i)\\+\\s+.*%s.*", helpers.DefaultSharedDomain()))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("the default route exists and is unmapped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-route", space, helpers.DefaultSharedDomain(), "-n", strings.ToLower(appName))).Should(Exit(0))
				})

				It("creates and maps the route with the provided hostname", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--hostname", hostname, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("routes:\\s+%s", route))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the default route is mapped to the application", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--hostname", hostname, "--no-start")).Should(Exit(0))
					})
				})

				It("does nothing", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--hostname", hostname, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					appGUID := helpers.AppGUID(appName)
					session := helpers.CF("curl", fmt.Sprintf("/v2/apps/%s/routes", appGUID))
					Eventually(session).Should(Say(`"total_results":\s+1,`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	When("using a tcp domain", func() {
		var (
			domain     helpers.Domain
			domainName string
		)

		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionRoutingV3)

			domainName = helpers.NewDomainName()
			domain = helpers.NewDomain(organization, domainName)
			domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
		})

		AfterEach(func() {
			domain.DeleteShared()
		})

		It("returns an error", func() {
			session := helpers.CF(PushCommandName, appName, "--hostname", "I-dont-care", "-d", domainName, "--no-start")
			Eventually(session.Err).Should(Say("The route is invalid: a hostname cannot be used with a TCP domain."))
			Eventually(session).Should(Exit(1))
		})
	})
})
