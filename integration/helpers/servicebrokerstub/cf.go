package servicebrokerstub

import (
	"encoding/json"
	"io"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceBrokerStub) registerServiceBroker() {
	Eventually(helpers.CF("create-service-broker", s.Name, "username", "password", s.URL)).Should(Exit(0))

	Eventually(helpers.CF("service-brokers")).Should(And(Exit(0), Say(s.Name)))

	Eventually(func() io.Reader {
		session := helpers.CF("service-access", "-b", s.Name)
		Eventually(session).Should(Exit(0))

		return session.Out
	}).Should(Say(s.FirstServiceOfferingName()))
}

func (s *ServiceBrokerStub) register() {
	Eventually(helpers.CF("create-service-broker", s.Name, s.Username, s.Password, s.URL)).Should(Exit(0))
}

func (s *ServiceBrokerStub) deregister() {
	s.purgeServiceOfferings(true)
	Eventually(helpers.CF("delete-service-broker", "-f", s.Name)).Should(Exit(0))
}

func (s *ServiceBrokerStub) registerViaV2() {
	broker := map[string]interface{}{
		"name":          s.Name,
		"broker_url":    s.URL,
		"auth_username": s.Username,
		"auth_password": s.Password,
	}
	data, err := json.Marshal(broker)
	Expect(err).NotTo(HaveOccurred())

	Eventually(helpers.CF("curl", "-X", "POST", "/v2/service_brokers", "-d", string(data))).Should(Exit(0))
}

func (s *ServiceBrokerStub) enableServiceAccess() {
	for _, offering := range s.Services {
		session := helpers.CF("enable-service-access", offering.Name)
		Eventually(session).Should(Exit(0))
	}
}

func (s *ServiceBrokerStub) purgeServiceOfferings(ignoreFailures bool) {
	for _, service := range s.Services {
		session := helpers.CF("purge-service-offering", service.Name, "-b", s.Name, "-f")
		session.Wait()

		if !ignoreFailures {
			Expect(session).To(Exit(0))
		}
	}
}
