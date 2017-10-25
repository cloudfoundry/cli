package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

type ServiceInstanceGUID struct {
	Resources []struct {
		Metadata struct {
			GUID string `json:"guid"`
		} `json:"metadata"`
	} `json:"resources"`
}

// ManagedServiceInstanceGUID returns the GUID for a managed service instance.
func ManagedServiceInstanceGUID(managedServiceInstanceName string) string {
	session := CF("curl", fmt.Sprintf("/v2/service_instances?q=name:%s", managedServiceInstanceName))
	Eventually(session).Should(Exit(0))

	rawJSON := strings.TrimSpace(string(session.Out.Contents()))

	var serviceInstanceGUID ServiceInstanceGUID
	err := json.Unmarshal([]byte(rawJSON), &serviceInstanceGUID)
	Expect(err).NotTo(HaveOccurred())

	Expect(serviceInstanceGUID.Resources).To(HaveLen(1))
	return serviceInstanceGUID.Resources[0].Metadata.GUID
}

// UserProvidedServiceInstanceGUID returns the GUID for a user provided service instance.
func UserProvidedServiceInstanceGUID(userProvidedServiceInstanceName string) string {
	session := CF("curl", fmt.Sprintf("/v2/user_provided_service_instances?q=name:%s", userProvidedServiceInstanceName))
	Eventually(session).Should(Exit(0))

	rawJSON := strings.TrimSpace(string(session.Out.Contents()))

	var serviceInstanceGUID ServiceInstanceGUID
	err := json.Unmarshal([]byte(rawJSON), &serviceInstanceGUID)
	Expect(err).NotTo(HaveOccurred())

	Expect(serviceInstanceGUID.Resources).To(HaveLen(1))
	return serviceInstanceGUID.Resources[0].Metadata.GUID
}
