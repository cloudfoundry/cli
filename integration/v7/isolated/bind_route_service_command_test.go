package isolated

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-route-service command", func() {
	const command = "bind-route-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Bind a service instance to an HTTP route\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf bind-route-service DOMAIN \[--hostname HOSTNAME\] \[--path PATH\] SERVICE_INSTANCE \[-c PARAMETERS_AS_JSON\]\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf bind-route-service example.com --hostname myapp --path foo myratelimiter\n`),
			Say(`\s+cf bind-route-service example.com myratelimiter -c file.json\n`),
			Say(`\s+cf bind-route-service example.com myratelimiter -c '{"valid":"json"}'\n`),
			Say(`\n`),
			Say(`\s+In Windows PowerShell use double-quoted, escaped JSON: "\{\\"valid\\":\\"json\\"\}"\n`),
			Say(`\s+In Windows Command Line use single-quoted, escaped JSON: '\{\\"valid\\":\\"json\\"\}'\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+brs\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided inline or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.\n`),
			Say(`\s+--hostname, -n\s+Hostname used in combination with DOMAIN to specify the route to bind\n`),
			Say(`\s+--path\s+Path used in combination with HOSTNAME and DOMAIN to specify the route to bind\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+routes, services\n`),
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

		When("-c is provided with invalid JSON", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-c", `{"not":json"}`)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("-c is provided with invalid JSON file", func() {
			It("displays a warning, the help text, and exits 1", func() {
				filename := helpers.TempFileWithContent(`{"not":json"}`)
				defer os.Remove(filename)

				session := helpers.CF(command, "-c", filename)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
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

		routeBindingStateForSI := func(serviceInstanceName string) string {
			var receiver struct {
				Resources []resources.RouteBinding `json:"resources"`
			}
			helpers.Curl(&receiver, "/v3/service_route_bindings?service_instance_names=%s", serviceInstanceName)
			Expect(receiver.Resources).To(HaveLen(1))

			return string(receiver.Resources[0].LastOperation.State)
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
			})

			It("creates a route binding", func() {
				session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(routeBindingStateForSI(serviceInstanceName)).To(Equal("succeeded"))
			})

			When("parameters are specified", func() {
				It("fails with an error returned by the CC", func() {
					session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName, "-c", `{"foo":"bar"}`)
					Eventually(session).Should(Exit(1))

					Expect(session.Out).To(SatisfyAll(
						Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
						Say(`FAILED\n`),
					))

					Expect(session.Err).To(Say(`Binding parameters are not supported for user-provided service instances\n`))
				})
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
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("creates a route binding", func() {
				session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(routeBindingStateForSI(serviceInstanceName)).To(Equal("succeeded"))
			})

			When("parameters are specified", func() {
				It("sends the parameters to the broker", func() {
					session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName, "-c", `{"foo":"bar"}`)
					Eventually(session).Should(Exit(0))

					var receiver struct {
						Resources []resources.RouteBinding `json:"resources"`
					}
					helpers.Curl(&receiver, "/v3/service_route_bindings?service_instance_names=%s", serviceInstanceName)
					Expect(receiver.Resources).To(HaveLen(1))

					var parametersReceiver map[string]interface{}
					helpers.Curl(&parametersReceiver, `/v3/service_route_bindings/%s/parameters`, receiver.Resources[0].GUID)
					Expect(parametersReceiver).To(Equal(map[string]interface{}{"foo": "bar"}))
				})
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

				broker.WithAsyncDelay(time.Second).Configure()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("starts to create a route binding", func() {
				session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Binding in progress\.\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(routeBindingStateForSI(serviceInstanceName)).To(Equal("in progress"))
			})

			When("--wait flag specified", func() {
				It("waits for completion", func() {
					session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName, "--wait")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					Expect(routeBindingStateForSI(serviceInstanceName)).To(Equal("succeeded"))
				})
			})
		})

		Context("route binding already exists", func() {
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

				session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))
			})

			It("says OK", func() {
				session := helpers.CF(command, domain, "--hostname", hostname, "--path", path, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding route %s.%s/%s to service instance %s in org %s / space %s as %s\.\.\.\n`, hostname, domain, path, serviceInstanceName, orgName, spaceName, username),
					Say(`Route %s.%s/%s is already bound to service instance %s\.\n`, hostname, domain, path, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})
	})
})
