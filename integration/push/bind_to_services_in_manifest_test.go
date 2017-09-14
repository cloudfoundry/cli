package push

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind app to provided services from manifest", func() {
	var (
		appName                         string
		serviceName                     string
		servicePlan                     string
		managedServiceInstanceName      string
		userProvidedServiceInstanceName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		serviceName = helpers.PrefixedRandomName("SERVICE")
		servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
		managedServiceInstanceName = helpers.PrefixedRandomName("si")
		userProvidedServiceInstanceName = helpers.PrefixedRandomName("usi")
	})

	Context("when the services do not exist", func() {
		It("fails with the service not found message", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name":     appName,
							"path":     dir,
							"services": []string{managedServiceInstanceName, userProvidedServiceInstanceName},
						},
					},
				})

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
				Eventually(session.Err).Should(Say("Service instance %s not found", managedServiceInstanceName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the services do exist", func() {
		var broker helpers.ServiceBroker

		BeforeEach(func() {
			domain := defaultSharedDomain()

			broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, serviceName, servicePlan)
			broker.Push()
			broker.Configure()
			broker.Create()

			Eventually(helpers.CF("enable-service-access", serviceName)).Should(Exit(0))

			Eventually(helpers.CF("create-service", serviceName, servicePlan, managedServiceInstanceName)).Should(Exit(0))

			Eventually(helpers.CF("create-user-provided-service", userProvidedServiceInstanceName)).Should(Exit(0))
		})

		AfterEach(func() {
			broker.Destroy()
		})

		Context("when the app is new", func() {
			It("binds the provided services", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":     appName,
								"path":     dir,
								"services": []string{managedServiceInstanceName, userProvidedServiceInstanceName},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\+\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("services:"))
					Eventually(session).Should(Say("\\+\\s+%s", managedServiceInstanceName))
					Eventually(session).Should(Say("\\+\\s+%s", userProvidedServiceInstanceName))
					Eventually(session).Should(Say("Binding services\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("services")
				Eventually(session).Should(Say("name\\s+service\\s+plan\\s+bound apps\\s+last operation"))
				Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+%s", managedServiceInstanceName, serviceName, servicePlan, appName))
				Eventually(session).Should(Say("%s\\s+user-provided\\s+%s", userProvidedServiceInstanceName, appName))

			})
		})

		Context("when the app already exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":     appName,
								"path":     dir,
								"services": []string{managedServiceInstanceName},
							},
						},
					})
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Exit(0))
				})
			})

			It("binds the unbound services", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":     appName,
								"path":     dir,
								"services": []string{managedServiceInstanceName, userProvidedServiceInstanceName},
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName)
					Eventually(session).Should(Say("Getting app info\\.\\.\\."))
					Eventually(session).Should(Say("Updating app with these attributes\\.\\.\\."))
					Eventually(session).Should(Say("\\s+name:\\s+%s", appName))
					Eventually(session).Should(Say("services:"))
					Eventually(session).Should(Say("(?m)$\\s+%s", managedServiceInstanceName))
					Eventually(session).Should(Say("\\+\\s+%s", userProvidedServiceInstanceName))
					Eventually(session).Should(Say("Binding services\\.\\.\\."))
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("services")
				Eventually(session).Should(Say("name\\s+service\\s+plan\\s+bound apps\\s+last operation"))
				Eventually(session).Should(Say("%s\\s+user-provided\\s+%s", userProvidedServiceInstanceName, appName))
			})
		})
	})
})
