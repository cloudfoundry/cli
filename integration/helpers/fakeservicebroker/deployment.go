package fakeservicebroker

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"fmt"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"net/http"
)

const (
	itsOrg             = "fakeservicebroker"
	itsSpace           = "integration"
	defaultMemoryLimit = "256M"
	defaultBrokerPath  = "../../assets/service_broker"
)

func (f *FakeServiceBroker) pushAppIfNecessary() {
	if !f.reusable {
		f.pushApp()
		return
	}

	if f.alreadyPushedApp() {
		return
	}

	f.pushReusableApp()
}

func (f *FakeServiceBroker) pushApp() {
	Eventually(helpers.CF(
		"push", f.name,
		"-p", defaultBrokerPath,
		"-m", defaultMemoryLimit,
	)).Should(Exit(0))
}

func (f *FakeServiceBroker) pushReusableApp() {
	helpers.CreateOrgAndSpaceUnlessExists(itsOrg, itsSpace)
	helpers.WithRandomHomeDir(func() {
		helpers.SetAPI()
		helpers.LoginCF()
		helpers.TargetOrgAndSpace(itsOrg, itsSpace)

		f.pushApp()
	})
}

func (f *FakeServiceBroker) alreadyPushedApp() bool {
	resp, err := http.Get(f.URL("/config"))
	Expect(err).ToNot(HaveOccurred())
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func generateReusableBrokerName(suffix string) string {
	return fmt.Sprintf("fake-service-broker-%s%02d", suffix, ginkgo.GinkgoParallelNode())
}

func (f *FakeServiceBroker) register() {
	if f.reusable {
		f.deregisterIgnoringFailures()
	}

	Eventually(helpers.CF("create-service-broker", f.name, "username", "password", f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))
}

func (f *FakeServiceBroker) update() {
	Eventually(helpers.CF("update-service-broker", f.name, "username", "password", f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))
}

func (f *FakeServiceBroker) cleanup() {
	f.deregisterIgnoringFailures()
	f.deleteAppIgnoringFailures()
}

func (f *FakeServiceBroker) deleteApp() {
	if f.reusable {
		return
	}

	Eventually(helpers.CF("delete", f.name, "-f", "-r")).Should(Exit(0))
}

func (f *FakeServiceBroker) deleteAppIgnoringFailures() {
	if f.reusable {
		return
	}

	Eventually(helpers.CF("delete", f.name, "-f", "-r")).Should(Exit(0))
}

func (f *FakeServiceBroker) deregister() {
	Eventually(helpers.CF("purge-service-offering", f.ServiceName(), "-b", f.name, "-f")).Should(Exit(0))
	Eventually(helpers.CF("delete-service-broker", f.name, "-f")).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Not(Say(f.name))))
}

func (f *FakeServiceBroker) deregisterIgnoringFailures() {
	Eventually(helpers.CF("purge-service-offering", f.ServiceName(), "-b", f.name, "-f")).Should(Exit())
	Eventually(helpers.CF("delete-service-broker", f.name, "-f")).Should(Exit())
}

func (f *FakeServiceBroker) stopReusing() *FakeServiceBroker {
	f.reusable = false
	return f
}
