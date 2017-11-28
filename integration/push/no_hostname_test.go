package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing with no-hostname", func() {
	var (
		appName    string
		domainName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		domainName = helpers.DomainName()
	})

	Context("when pushing with no manifest", func() {
		Context("when using a private domain", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domain = helpers.NewDomain(organization, domainName)
				domain.Create()
			})

			AfterEach(func() {
				domain.Delete()
			})

			It("creates a route with no hostname", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-hostname", "-d", domainName, "--no-start")
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("\\+\\s+%s", domainName))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", appName)
				Eventually(session).Should(Say("name:\\s+%s", appName))
				Eventually(session).Should(Say("(?m)routes:\\s+%s", domainName))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when using a shared domain", func() {
			Context("when using an HTTP domain", func() {
				It("returns an invalid route error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-hostname", "--no-start")
						Eventually(session.Err).Should(Say("The route is invalid: a hostname is required for shared domains."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when using a TCP domain", func() {
				var domain helpers.Domain

				BeforeEach(func() {
					domain = helpers.NewDomain(organization, domainName)
					domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
				})

				AfterEach(func() {
					domain.DeleteShared()
				})

				It("creates a TCP route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-hostname", "-d", domainName, "--no-start")
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
						Eventually(session).Should(Say("\\+\\s+%s:", domainName))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", appName)
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("(?m)routes:\\s+%s:\\d+", domainName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
