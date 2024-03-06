package push

import (
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("HTTP routes in manifest", func() {
	var (
		app           string
		domain        helpers.Domain
		subdomain     helpers.Domain
		route1        helpers.Route
		route2        helpers.Route
		routeWithPath helpers.Route
		routeStar     helpers.Route
	)

	BeforeEach(func() {
		app = helpers.NewAppName()
		domain = helpers.NewDomain(organization, helpers.NewDomainName())
		subdomain = helpers.NewDomain(organization, "sub."+domain.Name)
		route1 = helpers.NewRoute(space, domain.Name, helpers.PrefixedRandomName("r1"), "")
		route2 = helpers.NewRoute(space, subdomain.Name, helpers.PrefixedRandomName("r2"), "")
		routeWithPath = helpers.NewRoute(space, domain.Name, helpers.PrefixedRandomName("r3"), helpers.PrefixedRandomName("p1"))
		routeStar = helpers.NewRoute(space, subdomain.Name, "*", "")
	})

	When("the domain exist", func() {
		BeforeEach(func() {
			domain.Create()
			subdomain.Create()
		})

		When("the routes don't have a path", func() {
			When("the routes are new", func() {
				It("creates and maps the routes", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name": app,
									"routes": []map[string]string{
										{"route": route1.String()},
										{"route": route2.String()},
										{"route": routeStar.String()},
									},
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say(`Getting app info\.\.\.`))

						Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
						Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
						Eventually(session).Should(Say(`\s+routes:`))
						Eventually(session).Should(Say(`(?i)\+\s+%s`, regexp.QuoteMeta(routeStar.String())))
						Eventually(session).Should(Say(`(?i)\+\s+%s`, route1))
						Eventually(session).Should(Say(`(?i)\+\s+%s`, route2))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", app)
					Eventually(session).Should(Say(`name:\s+%s`, app))
					Eventually(session).Should(Say(`routes:\s+(%s, %s)|(%s, %s)`, route1, route2, route2, route1))
					Eventually(session).Should(Exit(0))
				})

				When("--random-route is also specified", func() {
					It("creates and maps the routes", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name":         app,
										"random-route": true,
										"routes": []map[string]string{
											{"route": route1.String()},
											{"route": route2.String()},
										},
									},
								},
							})

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
							Eventually(session).Should(Say(`Getting app info\.\.\.`))

							Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
							Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
							Eventually(session).Should(Say(`\s+routes:`))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, route1))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, route2))
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("app", app)
						Eventually(session).Should(Say(`name:\s+%s`, app))
						Eventually(session).Should(Say(`routes:\s+(%s, %s)|(%s, %s)`, route1, route2, route2, route1))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("one of the routes exists", func() {
				When("the route is in the current space", func() {
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
							Eventually(session).Should(Say(`Getting app info\.\.\.`))

							Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
							Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
							Eventually(session).Should(Say(`\s+routes:`))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, route1))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, route2))
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("app", app)
						Eventually(session).Should(Say(`name:\s+%s`, app))
						Eventually(session).Should(Say(`routes:\s+(%s, %s)|(%s, %s)`, route1, route2, route2, route1))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the route is in a different space", func() {
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

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
							Eventually(session).Should(Say(`Getting app info\.\.\.`))
							Eventually(session.Err).Should(Say("The app cannot be mapped to route %s because the route exists in a different space.", route2))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			When("the route contains a port", func() {
				BeforeEach(func() {
					route1.Port = 1234
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
						Eventually(session).Should(Say(`Getting app info\.\.\.`))
						Eventually(session.Err).Should(Say("Port not allowed in HTTP domain %s", domain.Name))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the route has a path", func() {

			When("the route is new", func() {
				It("creates and maps the route", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name": app,
									"routes": []map[string]string{
										{"route": routeWithPath.String()},
									},
								},
							},
						})
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
						Eventually(session).Should(Say(`Getting app info\.\.\.`))

						Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
						Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
						Eventually(session).Should(Say(`\s+routes:`))
						Eventually(session).Should(Say(`(?i)\+\s+%s`, routeWithPath))
						Eventually(session).Should(Exit(0))
					})

					session := helpers.CF("app", app)
					Eventually(session).Should(Say(`name:\s+%s`, app))
					Eventually(session).Should(Say(`routes:\s+%s`, routeWithPath))
					Eventually(session).Should(Exit(0))
				})

				// This test refers to this dev consideration [https://www.pivotaltracker.com/story/show/126256733/comments/181122853]
				When("a route with the same hostname and domain exists", func() {
					var routeWithoutPath helpers.Route

					BeforeEach(func() {
						routeWithoutPath = routeWithPath
						routeWithoutPath.Path = ""

						routeWithoutPath.Create()
					})

					It("creates and maps the route", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name": app,
										"routes": []map[string]string{
											{"route": routeWithPath.String()},
										},
									},
								},
							})
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
							Eventually(session).Should(Say(`Getting app info\.\.\.`))

							Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
							Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
							Eventually(session).Should(Say(`\s+routes:`))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, routeWithPath))
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("app", app)
						Eventually(session).Should(Say(`name:\s+%s`, app))
						Eventually(session).Should(Say(`routes:\s+%s`, routeWithPath))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the route exists", func() {
				When("the route exists in the same space", func() {
					BeforeEach(func() {
						routeWithPath.Create()
					})

					It("creates and maps the new route; maps the old route", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name": app,
										"routes": []map[string]string{
											{"route": routeWithPath.String()},
										},
									},
								},
							})

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
							Eventually(session).Should(Say(`Getting app info\.\.\.`))

							Eventually(session).Should(Say(`Creating app with these attributes\.\.\.`))
							Eventually(session).Should(Say(`\+\s+name:\s+%s`, app))
							Eventually(session).Should(Say(`\s+routes:`))
							Eventually(session).Should(Say(`(?i)\+\s+%s`, routeWithPath))
							Eventually(session).Should(Exit(0))
						})

						session := helpers.CF("app", app)
						Eventually(session).Should(Say(`name:\s+%s`, app))
						Eventually(session).Should(Say(`routes:\s+%s`, routeWithPath))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the route exists in another space", func() {
					BeforeEach(func() {
						otherSpace := helpers.NewSpaceName()
						helpers.CreateSpace(otherSpace)
						routeWithPath.Space = otherSpace
						routeWithPath.Create()
					})

					It("returns an error", func() {
						helpers.WithHelloWorldApp(func(dir string) {
							helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
								"applications": []map[string]interface{}{
									{
										"name": app,
										"routes": []map[string]string{
											{"route": routeWithPath.String()},
										},
									},
								},
							})

							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, "--no-start")
							Eventually(session).Should(Say(`Getting app info\.\.\.`))
							Eventually(session.Err).Should(Say("The app cannot be mapped to route %s because the route exists in a different space.", routeWithPath))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})
	})

	When("the domain does not exist", func() {
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
				Eventually(session).Should(Say(`Getting app info\.\.\.`))
				Eventually(session.Err).Should(Say("The route %s did not match any existing domains.", route1))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
