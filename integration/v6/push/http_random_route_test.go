package push

import (
	"path"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP random route", func() {

	const randomRouteRegexp = `\+\s+%s-[\w]+-[\w]+-[a-z]{2}\.%s`

	var (
		appName string
	)

	BeforeEach(func() {
		appName = "short-app-name" // used on purpose to fit route length requirement
	})

	When("passed the --random-route flag", func() {
		When("the app does not already exist", func() {
			It("generates a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-start")
					Eventually(session).Should(Say("routes:"))
					Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Exit(0))
				})

				appSession := helpers.CF("app", appName)
				Eventually(appSession).Should(Say(`name:\s+%s`, appName))
				Eventually(appSession).Should(Say(`routes:\s+%s-[\w]+-[\w]+-[a-z]{2}\.%s`, appName, helpers.DefaultSharedDomain()))
				Eventually(appSession).Should(Exit(0))
			})
		})

		When("the app exists and has an existing route", func() {
			It("does not generate a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-start")
					Eventually(session).ShouldNot(Say(`\+\s+%s-[\w]+`, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("also passed an http domain", func() {
				var domain helpers.Domain

				BeforeEach(func() {
					domain = helpers.NewDomain(organization, helpers.NewDomainName("some-domain"))
					domain.Create()
				})

				AfterEach(func() {
					domain.Delete()
				})

				It("does not create a new route with the provided domain name", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", domain.Name, "--no-start")
						Eventually(session).ShouldNot(Say(randomRouteRegexp, appName, domain.Name))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("also passed an http domain", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domain = helpers.NewDomain(organization, helpers.NewDomainName("some-domain"))
				domain.Create()
			})

			AfterEach(func() {
				domain.Delete()
			})

			It("creates a new route with the provided domain", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "-d", domain.Name, "--no-start")
					Eventually(session).Should(Say(randomRouteRegexp, appName, domain.Name))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	When("passed the random-route manifest property", func() {
		var manifest map[string]interface{}

		BeforeEach(func() {
			manifest = map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":         appName,
						"random-route": true,
					},
				},
			}

			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(path.Join(dir, "manifest.yml"), manifest)
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})

		It("generates a random route for the app", func() {
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(`routes:\s+%s-[\w]+-[\w]+-[a-z]{2}\.%s`, appName, helpers.DefaultSharedDomain()))
			Eventually(session).Should(Exit(0))
		})

		When("the app has an existing route", func() {
			It("does not generate a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(path.Join(dir, "manifest.yml"), manifest)

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
					Eventually(session).ShouldNot(Say(`\+\s+%s-[\w]+`, appName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
