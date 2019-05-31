package fakeservicebroker

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func (f *FakeServiceBroker) pushApp() {
	Eventually(helpers.CF(
		"push", f.name,
		"--no-start",
		"-m", "256M",
		"-p", "../../assets/service_broker",
		"--no-route",
	)).Should(Exit(0))

	Eventually(helpers.CF(
		"map-route",
		f.name,
		f.domain,
		"--hostname", f.name,
	)).Should(Exit(0))

	Eventually(helpers.CF("start", f.name)).Should(Exit(0))
}

func (f *FakeServiceBroker) register() {
	Eventually(helpers.CF("create-service-broker", f.name, "username", "password", f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))
}

func (f *FakeServiceBroker) update() {
	Eventually(helpers.CF("update-service-broker", f.name, "username", "password", f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))
}

func (f *FakeServiceBroker) deleteApp() {
	Eventually(helpers.CF("delete", f.name, "-f", "-r")).Should(Exit(0))
}

func (f *FakeServiceBroker) deregister() {
	Eventually(helpers.CF("purge-service-offering", f.ServiceName(), "-b", f.name, "-f")).Should(Exit(0))
	Eventually(helpers.CF("delete-service-broker", f.name, "-f")).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Not(Say(f.name))))
}
