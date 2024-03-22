package isolated

import (
	"regexp"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-route-service command", func() {
	const command = "unbind-route-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("unbind-route-service - Unbind a service instance from an HTTP route"),
			Say("USAGE:"),
			Say(regexp.QuoteMeta("cf unbind-route-service DOMAIN [--hostname HOSTNAME] [--path PATH] SERVICE_INSTANCE [-f]")),
			Say("EXAMPLES:"),
			Say("cf unbind-route-service example.com --hostname myapp --path foo myratelimiter"),
			Say("ALIAS:"),
			Say("urs"),
			Say("OPTIONS:"),
			Say(`-f\s+Force unbinding without confirmation`),
			Say(`--hostname, -n\s+Hostname used in combination with DOMAIN to specify the route to unbind`),
			Say(`--path\s+Path used in combination with HOSTNAME and DOMAIN to specify the route to unbind`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say("SEE ALSO:"),
			Say("delete-service, routes, services"),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `DOMAIN` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "foo", "bar")
		})
	})

	When("targeting a space", func() {
		var (
			orgName   string
			spaceName string
			username  string
		)

		routeBindingCount := func(serviceInstanceName string) int {
			var receiver struct {
				Resources []resources.RouteBinding `json:"resources"`
			}
			helpers.Curl(&receiver, "/v3/service_route_bindings?service_instance_names=%s", serviceInstanceName)
			return len(receiver.Resources)
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("user-provided route service", func() {
			var (
				routeServiceURL     string
				serviceInstanceName string
				domain              string
				hostname            string
				path                string
			)

			BeforeEach(func() {
				routeServiceURL = helpers.RandomURL()
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName, "-r", routeServiceURL)).Should(Exit(0))

				domain = helpers.DefaultSharedDomain()
				hostname = helpers.NewHostName()
				path = helpers.PrefixedRandomName("path")
				Eventually(helpers.CF("create-route", domain, "--hostname", hostname, "--path", path)).Should(Exit(0))

				Eventually(helpers.CF("bind-route-service", domain, "--hostname", hostname, "--path", path, serviceInstanceName, "--wait")).Should(Exit(0))
			})

			It("deletes a route binding", func() {
				session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding route %s.%s/%s from service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(routeBindingCount(serviceInstanceName)).To(BeZero())
			})
		})

		Context("managed route service with synchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				domain              string
				hostname            string
				path                string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithRouteService().EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				domain = helpers.DefaultSharedDomain()
				hostname = helpers.NewHostName()
				path = helpers.PrefixedRandomName("path")
				Eventually(helpers.CF("create-route", domain, "--hostname", hostname, "--path", path)).Should(Exit(0))

				Eventually(helpers.CF("bind-route-service", domain, "--hostname", hostname, "--path", path, serviceInstanceName, "--wait")).Should(Exit(0))
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("deletes a route binding", func() {
				session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding route %s.%s/%s from service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(routeBindingCount(serviceInstanceName)).To(BeZero())
			})
		})

		Context("managed route service with asynchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				domain              string
				hostname            string
				path                string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithRouteService().EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				domain = helpers.DefaultSharedDomain()
				hostname = helpers.NewHostName()
				path = helpers.PrefixedRandomName("path")
				Eventually(helpers.CF("create-route", domain, "--hostname", hostname, "--path", path)).Should(Exit(0))

				Eventually(helpers.CF("bind-route-service", domain, "--hostname", hostname, "--path", path, serviceInstanceName, "--wait")).Should(Exit(0))

				broker.WithAsyncDelay(time.Second).Configure()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("starts to delete a route binding", func() {
				session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding route %s.%s/%s from service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Unbinding in progress\.\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(routeBindingCount(serviceInstanceName)).NotTo(BeZero())
			})

			When("--wait flag specified", func() {
				It("waits for completion", func() {
					session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName, "--wait")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Unbinding route %s.%s/%s from service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					Expect(routeBindingCount(serviceInstanceName)).To(BeZero())
				})
			})
		})

		Context("route binding does not exists", func() {
			var (
				routeServiceURL     string
				serviceInstanceName string
				domain              string
				hostname            string
				path                string
			)

			BeforeEach(func() {
				routeServiceURL = helpers.RandomURL()
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName, "-r", routeServiceURL)).Should(Exit(0))

				domain = helpers.DefaultSharedDomain()
				hostname = helpers.NewHostName()
				path = helpers.PrefixedRandomName("path")
				Eventually(helpers.CF("create-route", domain, "--hostname", hostname, "--path", path)).Should(Exit(0))

				session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))
			})

			It("says OK", func() {
				session := helpers.CF(command, "-f", domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Unbinding route %s.%s/%s from service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`Route %s.%s/%s was not bound to service instance %s\.\n`, hostname, domain, path, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})
	})
})
