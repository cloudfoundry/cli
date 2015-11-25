package command_registry_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dependency", func() {
	var dependency command_registry.Dependency

	It("populates all fields by calling all the dependency contructors", func() {
		dependency = command_registry.NewDependency()

		Ω(dependency.Ui).ToNot(BeNil())
		Ω(dependency.Config).ToNot(BeNil())
		Ω(dependency.RepoLocator).ToNot(BeNil())
		Ω(dependency.PluginConfig).ToNot(BeNil())
		Ω(dependency.ManifestRepo).ToNot(BeNil())
		Ω(dependency.AppManifest).ToNot(BeNil())
		Ω(dependency.Gateways).ToNot(BeNil())
		Ω(dependency.TeePrinter).ToNot(BeNil())
		Ω(dependency.PluginRepo).ToNot(BeNil())
		Ω(dependency.PluginModels).ToNot(BeNil())
		Ω(dependency.ServiceBuilder).ToNot(BeNil())
		Ω(dependency.BrokerBuilder).ToNot(BeNil())
		Ω(dependency.PlanBuilder).ToNot(BeNil())
		Ω(dependency.ServiceHandler).ToNot(BeNil())
		Ω(dependency.ServicePlanHandler).ToNot(BeNil())
		Ω(dependency.WordGenerator).ToNot(BeNil())
		Ω(dependency.AppZipper).ToNot(BeNil())
		Ω(dependency.AppFiles).ToNot(BeNil())
		Ω(dependency.PushActor).ToNot(BeNil())
		Ω(dependency.ChecksumUtil).ToNot(BeNil())
	})
})
