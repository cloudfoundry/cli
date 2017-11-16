package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("TCP random route", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = "short-app-name" // used on purpose to fit route length requirement
	})

	Context("when passed the --random-route flag", func() {
		Context("when also passed a tcp domain", func() {
			var tcpDomain helpers.Domain

			BeforeEach(func() {
				domainName := helpers.DomainName("tcp-domain")
				tcpDomain = helpers.NewDomain(organization, domainName)
				tcpDomain.CreateWithRouterGroup(helpers.DefaultTCPRouterGroup)
			})

			AfterEach(func() {
				tcpDomain.DeleteShared()
			})

			It("creates a new route with the provided domain", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", tcpDomain.Name, "--no-start")
					Eventually(session).Should(Say("\\+\\s+%s:\\?\\?\\?\\?", tcpDomain.Name))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app already exists with a tcp route", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", tcpDomain.Name, "--no-start")).Should(Exit(0))
					})
				})

				It("does not create any new routes", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", tcpDomain.Name, "--no-start")
						Consistently(session).ShouldNot(Say("\\+\\s+%s:", tcpDomain.Name))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
