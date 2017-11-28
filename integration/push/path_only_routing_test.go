package push

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with route path", func() {
	var (
		appName string
		route   string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the default domain is a HTTP domain", func() {
		Context("when the route path is provided", func() {
			var routePath string

			BeforeEach(func() {
				routePath = "some-path"
				route = fmt.Sprintf("%s.%s/%s", strings.ToLower(appName), defaultSharedDomain(), routePath)
			})

			Context("when the default route with path does not exist", func() {
				It("creates and maps the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--route-path", routePath, "--no-start")
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

			Context("the default route with path exists and is unmapped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-route", space, defaultSharedDomain(), "-n", strings.ToLower(appName), "--path", routePath)).Should(Exit(0))
				})

				It("maps the route with the provided path", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--route-path", routePath, "--no-start")
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

			Context("when the default route with path exists and is mapped to the application", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--route-path", routePath, "--no-start")).Should(Exit(0))
					})
				})

				It("does nothing", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--route-path", routePath, "--no-start")
						Eventually(session).Should(Say("routes:"))
						Eventually(session).Should(Say("(?i)\\s+%s", route))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("routes:\\s+%s", route))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when using a tcp domain", func() {
			var (
				domain     helpers.Domain
				domainName string
			)

			BeforeEach(func() {
				domainName = helpers.DomainName()
				domain = helpers.NewDomain(organization, domainName)
				domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
			})

			AfterEach(func() {
				domain.DeleteShared()
			})

			It("returns an error", func() {
				session := helpers.CF(PushCommandName, appName, "--route-path", "/potatoes", "-d", domainName, "--no-start")
				Eventually(session.Err).Should(Say("The route is invalid: a route path cannot be used with a TCP domain."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
