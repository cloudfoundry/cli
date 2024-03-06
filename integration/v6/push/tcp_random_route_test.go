package push

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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

	When("passed the --random-route flag", func() {
		When("also passed a tcp domain", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domainName := helpers.NewDomainName("tcp-domain")
				domain = helpers.NewDomain(organization, domainName)
				domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
			})

			AfterEach(func() {
				domain.DeleteShared()
			})

			It("creates a new route with the provided domain", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", domain.Name, "--no-start")
					Eventually(session).Should(Say(`\+\s+%s:\?\?\?\?`, domain.Name))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app already exists with a tcp route", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", domain.Name, "--no-start")).Should(Exit(0))
					})
				})

				It("does not create any new routes", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", domain.Name, "--no-start")
						Consistently(session).ShouldNot(Say(`\+\s+%s:`, domain.Name))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
