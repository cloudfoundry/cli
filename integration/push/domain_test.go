package push

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with different domain values", func() {
	var (
		appName    string
		domainName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		domainName = helpers.DomainName("http-domain")
	})

	Context("when the domain flag is not provided", func() {
		It("creates a route with the first shared domain", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
					PushCommandName, appName, "--no-start",
				)
				Eventually(session).Should(Say("\\s+routes:\\s+%s.%s", strings.ToLower(appName), defaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when only the domain flag is provided", func() {
		Context("When the domain does not exist", func() {
			It("creates a route that has the specified domain", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName, "--no-start",
						"-d", domainName,
					)
					Eventually(session.Err).Should(Say("Domain %s not found", domainName))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when domain is an HTTP domain", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domain = helpers.NewDomain(organization, domainName)
				domain.Create()
			})

			AfterEach(func() {
				domain.Delete()
			})

			It("creates a route that has the specified domain", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName, "--no-start",
						"-d", domainName,
					)
					Eventually(session).Should(Say("\\s+routes:\\s+%s.%s", strings.ToLower(appName), domainName))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when domain is a TCP domain", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domainName = helpers.DomainName("tcp-domain")
				domain = helpers.NewDomain(organization, domainName)
				domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
			})

			AfterEach(func() {
				domain.DeleteShared()
			})

			It("creates a new route with the specified domain and a random port each time the app is pushed", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName, "--no-start",
						"-d", domainName,
					)
					Eventually(session).Should(Say("\\+\\s+%s:\\?\\?\\?\\?", domainName))
					Eventually(session).Should(Say("\\s+routes:\\s+%s:\\d+", domainName))
					// instead of checking that the port is different each time we push
					// the same app, we check that the push does not fail
					Eventually(session).Should(Exit(0))

					session = helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir},
						PushCommandName, appName,
						"-d", domainName,
					)
					Eventually(session).Should(Say("\\+\\s+%s:\\?\\?\\?\\?", domainName))
					Eventually(session).Should(Say("\\s+routes:\\s+%s:\\d+", domainName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
