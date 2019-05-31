package fakeservicebroker

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const (
	defaultMemoryLimit = "256M"
	defaultBrokerPath  = "../../assets/service_broker"
)

func (f *FakeServiceBroker) pushApp() {
	if helpers.V7 {
		// cf start on V7 currently does not stage the app before starting after a push --no-start
		Eventually(helpers.CF(
			"push", f.name,
			"-p", defaultBrokerPath,
			"-m", defaultMemoryLimit,
			"--no-route",
		)).Should(Exit(0))
	} else {
		Eventually(helpers.CF(
			"push", f.name,
			"--no-start",
			"-m", defaultMemoryLimit,
			"-p", defaultBrokerPath,
			"--no-route",
		)).Should(Exit(0))
	}

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
