package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP routes in manifest", func() {
	var (
		app       string
		domain    helpers.Domain
		subdomain helpers.Domain
		route1    helpers.Route
		route2    helpers.Route
	)

	BeforeEach(func() {
		app = helpers.NewAppName()
		domain = helpers.NewDomain(organization, helpers.DomainName())
		subdomain = helpers.NewDomain(organization, "sub."+domain.Name)
		route1 = helpers.NewRoute(space, domain.Name, helpers.PrefixedRandomName("r1"), "")
		route2 = helpers.NewRoute(space, subdomain.Name, helpers.PrefixedRandomName("r2"), "")
	})

	Context("when the domain exist", func() {
		BeforeEach(func() {
			domain.Create()
			subdomain.Create()
		})

		Context("when the routes are new", func() {
			It("creates and binds the routes", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": app,
								"routes": []map[string]string{
									{"route": route1.String()},
									{"route": route2.String()},
								},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))

					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", app))
					Eventually(session).Should(Say("\\s+routes:"))
					Eventually(session).Should(Say("(?i)\\+\\s+%s", route1))
					Eventually(session).Should(Say("(?i)\\+\\s+%s", route2))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("app", app)
				Eventually(session).Should(Say("name:\\s+%s", app))
				Eventually(session).Should(Say("routes:\\s+(%s, %s)|(%s, %s)", route1, route2, route2, route1))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when one of the routes exists", func() {
			Context("when the route is in the current space", func() {
				BeforeEach(func() {
					route2.Create()
				})

				It("creates and binds the new route; binds the old route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name": app,
									"routes": []map[string]string{
										{"route": route1.String()},
										{"route": route2.String()},
									},
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))

						Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
						Eventually(session).Should(Say("\\+\\s+name:\\s+%s", app))
						Eventually(session).Should(Say("\\s+routes:"))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route1))
						Eventually(session).Should(Say("(?i)\\+\\s+%s", route2))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", app)
					Eventually(session).Should(Say("name:\\s+%s", app))
					Eventually(session).Should(Say("routes:\\s+(%s, %s)|(%s, %s)", route1, route2, route2, route1))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the route is in a different space", func() {
				BeforeEach(func() {
					otherSpace := helpers.NewSpaceName()
					helpers.CreateSpace(otherSpace)
					route2.Space = otherSpace
					route2.Create()
				})

				It("returns an error", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name": app,
									"routes": []map[string]string{
										{"route": route1.String()},
										{"route": route2.String()},
									},
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session.Err).Should(Say("The route %s is already in use.", route2))
						Eventually(session.Err).Should(Say("TIP: Change the hostname with -n HOSTNAME or use --random-route to generate a new route and then push again."))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		Context("when the route contains a port", func() {
			BeforeEach(func() {
				route1 = helpers.NewTCPRoute(space, domain.Name, 1234)
			})

			It("returns an error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": app,
								"routes": []map[string]string{
									{"route": route1.String()},
									{"route": route2.String()},
								},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session.Err).Should(Say("Port not allowed in HTTP domain %s", domain.Name))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when the domain does not exist", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": app,
							"routes": []map[string]string{
								{"route": route1.String()},
								{"route": route2.String()},
							},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session.Err).Should(Say("The route %s did not match any existing domains.", route1))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
