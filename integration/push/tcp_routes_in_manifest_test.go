package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("TCP routes in manifest", func() {
	var (
		app    string
		domain helpers.Domain
		route1 helpers.Route
		route2 helpers.Route
	)

	BeforeEach(func() {
		app = helpers.NewAppName()
		domain = helpers.NewDomain(organization, helpers.DomainName())
		route1 = helpers.NewTCPRoute(space, domain.Name, 1024)
		route2 = helpers.NewTCPRoute(space, domain.Name, 1025)
	})

	Context("when the domain exists", func() {
		BeforeEach(func() {
			domain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelNode()))
		})

		AfterEach(func() {
			domain.DeleteShared()
		})

		Context("when the routes are new", func() {
			It("creates and maps the routes", func() {
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

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
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

		Context("when a route already exist", func() {
			Context("when the routes exist in the current space", func() {
				BeforeEach(func() {
					route2.Create()
				})

				It("creates and maps the new route; maps the old route", func() {
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

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
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

			Context("when the routes exist in another space", func() {
				var otherSpace string

				BeforeEach(func() {
					otherSpace = helpers.NewSpaceName()
					helpers.CreateSpace(otherSpace)
					route2.Space = otherSpace
					route2.Create()
				})

				AfterEach(func() {
					helpers.QuickDeleteSpace(otherSpace)
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

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say("Getting app info\\.\\.\\."))
						Eventually(session.Err).Should(Say("The app cannot be mapped to route %s because the route is not in this space. Apps must be mapped to routes in the same space.", route2))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		Context("when a host is provided", func() {
			BeforeEach(func() {
				route1.Host = "some-host"
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

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session.Err).Should(Say("Host and path not allowed in route with TCP domain %s", route1.Domain))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when an path is provided", func() {
			It("returns an error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": app,
								"routes": []map[string]string{
									{"route": route1.String() + "/some-path"},
								},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
					Eventually(session.Err).Should(Say("Host and path not allowed in route with TCP domain %s", route1.Domain))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when the domains don't exist", func() {
		It("returns an error", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": app,
							"routes": []map[string]string{
								{"route": route1.String()},
							},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session.Err).Should(Say("The route %s did not match any existing domains.", route1))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
