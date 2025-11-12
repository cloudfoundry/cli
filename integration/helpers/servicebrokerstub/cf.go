package servicebrokerstub

import (
	"encoding/json"

	"code.cloudfoundry.org/cli/v9/integration/helpers"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func (s *ServiceBrokerStub) register(spaceScoped bool) {
	var params []string

	switch s.registered {
	case true:
		params = []string{"update-service-broker"}
	case false:
		params = []string{"create-service-broker"}
	}

	params = append(params, s.Name, s.Username, s.Password, s.URL)

	if spaceScoped {
		params = append(params, "--space-scoped")
	}

	Eventually(helpers.CF(params...)).Should(Exit(0))
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
	config := s.ServiceAccessConfig

	if len(config) == 0 {
		for _, offering := range s.Services {
			config = append(config, ServiceAccessConfig{OfferingName: offering.Name})
		}
	}

	for _, c := range config {
		args := []string{"enable-service-access", c.OfferingName, "-b", s.Name}

		if c.PlanName != "" {
			args = append(args, "-p", c.PlanName)
		}

		if c.OrgName != "" {
			args = append(args, "-o", c.OrgName)
		}

		Eventually(helpers.CF(args...)).Should(Exit(0))
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
