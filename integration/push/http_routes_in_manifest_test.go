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
		app    string
		domain helpers.Domain
		route1 helpers.Route
		route2 helpers.Route
	)

	BeforeEach(func() {
		app = helpers.NewAppName()
		domain = helpers.NewDomain(organization, helpers.DomainName())
		route1 = helpers.NewRoute(space, domain.Name, helpers.PrefixedRandomName("r1"), "")
		route2 = helpers.NewRoute(space, domain.Name, helpers.PrefixedRandomName("r2"), "")
	})

	Context("when the domain exist", func() {
		BeforeEach(func() {
			domain.Create()
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
	})

	Context("when the domain does not exist", func() {
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
				Eventually(session.Err).Should(Say("Domain %s not found", domain.Name))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
