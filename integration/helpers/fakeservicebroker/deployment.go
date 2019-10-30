package fakeservicebroker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
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

func (f *FakeServiceBroker) getCatalogServiceNames() []string {
	if !f.alreadyPushedApp() {
		return nil
	}

	resp, err := http.Get(f.URL("/v2/catalog"))
	Expect(err).ToNot(HaveOccurred())

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred())

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("expected '%s' to be a valid json", string(body)))

	var serviceNames []string
	for _, value := range data["services"].([]interface{}) {
		service := value.(map[string]interface{})
		serviceNames = append(serviceNames, service["name"].(string))
	}

	return serviceNames
}

func generateReusableBrokerName(suffix string) string {
	return fmt.Sprintf("fake-service-broker-%s%02d", suffix, ginkgo.GinkgoParallelNode())
}

func (f *FakeServiceBroker) register() {
	if f.reusable {
		f.deregister()
	}

	Eventually(helpers.CF("create-service-broker", f.name, f.username, f.password, f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))

	Eventually(func() io.Reader {
		session := helpers.CF("service-access", "-b", f.name)
		Eventually(session).Should(Exit(0))

		return session.Out
	}).Should(Say(f.ServiceName()))
}

func (f *FakeServiceBroker) update() {
	Eventually(helpers.CF("update-service-broker", f.name, "username", "password", f.URL())).Should(Exit(0))
	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(f.name)))
}

func (f *FakeServiceBroker) cleanup() {
	f.deregister()
	f.deleteApp()
}

func (f *FakeServiceBroker) deleteApp() {
	if f.reusable {
		return
	}

	Eventually(helpers.CF("delete", f.name, "-f", "-r")).Should(Exit(0))
}

func (f *FakeServiceBroker) deregister() {
	f.purgeAllServiceOfferings(true)

	Eventually(helpers.CF("delete-service-broker", f.name, "-f")).Should(Exit())
}

func (f *FakeServiceBroker) purgeAllServiceOfferings(ignoreFailures bool) {
	for _, service := range f.getCatalogServiceNames() {
		assertion := Eventually(helpers.CF("purge-service-offering", service, "-b", f.name, "-f"))

		if ignoreFailures {
			assertion.Should(Exit())
		} else {
			assertion.Should(Exit(0))
		}
	}
}

func (f *FakeServiceBroker) stopReusing() *FakeServiceBroker {
	f.reusable = false
	return f
}
