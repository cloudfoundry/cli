package push

import (
	"fmt"
	"path"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("random route", func() {
	const randomRouteRegexp = `(?m)routes:\s+%s-[\w]+-[\w]+(-[\w]+)?\.%s$`

	var (
		appName string
	)

	BeforeEach(func() {
		appName = fmt.Sprintf("app%d", time.Now().Nanosecond())
	})

	When("passed the --random-route flag", func() {
		When("the app does not already exist", func() {
			It("generates a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-manifest", "--no-start")
					Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Exit(0))
				})

				appSession := helpers.CF("app", appName)
				Eventually(appSession).Should(Say(`name:\s+%s`, appName))
				Eventually(appSession).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
				Eventually(appSession).Should(Exit(0))
			})
		})

		When("the app exists and has an existing route", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-manifest", "--no-start")
					Eventually(session).Should(Exit(0))
				})
			})

			It("does not generate a random route for the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--random-route", "--no-manifest", "--no-start")
					Eventually(session).Should(Say(`(?m)routes:\s+%s\.%s$`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})

	When("passed the random-route manifest property as 'true' and no --random-route flag", func() {
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
				Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})

		It("generates a random route for the app", func() {
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
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

	When("passed the random-route manifest property as 'false' and with --random-route flag", func() {
		var manifest map[string]interface{}

		BeforeEach(func() {
			manifest = map[string]interface{}{
				"applications": []map[string]interface{}{
					{
						"name":         appName,
						"random-route": false,
					},
				},
			}

			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(path.Join(dir, "manifest.yml"), manifest)
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start", "--random-route")
				Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
				Eventually(session).Should(Exit(0))
			})
		})

		It("generates a random route for the app", func() {
			session := helpers.CF("app", appName)
			Eventually(session).Should(Say(`name:\s+%s`, appName))
			Eventually(session).Should(Say(randomRouteRegexp, appName, helpers.DefaultSharedDomain()))
			Eventually(session).Should(Exit(0))
		})
	})
})
