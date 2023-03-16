package helpers

import (
	"fmt"
	"strings"

	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

func ServiceInstanceGUID(serviceInstanceName string) string {
	session := CF("curl", fmt.Sprintf("/v3/service_instances?names=%s", serviceInstanceName))
	Eventually(session).Should(Exit(0))

	rawJSON := strings.TrimSpace(string(session.Out.Contents()))

	var serviceInstanceDetails struct {
		GUIDs []string `jsonry:"resources.guid"`
	}
	err := jsonry.Unmarshal([]byte(rawJSON), &serviceInstanceDetails)
	Expect(err).NotTo(HaveOccurred())

	Expect(serviceInstanceDetails.GUIDs).NotTo(BeEmpty(), fmt.Sprintf("service instance %s not found", serviceInstanceName))
	Expect(serviceInstanceDetails.GUIDs[0]).To(HaveLen(36), fmt.Sprintf("invalid GUID: '%s'", serviceInstanceDetails.GUIDs[0]))
	return serviceInstanceDetails.GUIDs[0]
}

// CreateManagedServiceInstance also waits for completion
func CreateManagedServiceInstance(offering, plan, name string, additional ...string) {
	create := []string{"create-service", offering, plan, name}
	session := CF(append(create, additional...)...)
	Eventually(session).Should(Exit(0))

	Eventually(func() *Buffer {
		return CF("service", name).Wait().Out
	}).Should(Say(`status:\s+create succeeded`))
}
