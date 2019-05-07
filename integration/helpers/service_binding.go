package helpers

import (
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// PollLastOperationUntilSuccess polls the last operation performed on a service instance bound to a given app until
// success. An expectation will fail if the last operation does not succeed or polling takes over 5 minutes.
func PollLastOperationUntilSuccess(client *ccv2.Client, appName string, serviceInstanceName string) {
	apps, _, err := client.GetApplications(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{appName},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(apps).To(HaveLen(1))

	serviceInstances, _, err := client.GetServiceInstances(ccv2.Filter{
		Type:     constant.NameFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceInstanceName},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(serviceInstances).To(HaveLen(1))

	bindings, _, err := client.GetServiceBindings(ccv2.Filter{
		Type:     constant.AppGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{apps[0].GUID},
	}, ccv2.Filter{
		Type:     constant.ServiceInstanceGUIDFilter,
		Operator: constant.EqualOperator,
		Values:   []string{serviceInstances[0].GUID},
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(bindings).To(HaveLen(1))
	binding := bindings[0]

	startTime := time.Now()
	for binding.LastOperation.State == constant.LastOperationInProgress {
		if time.Now().After(startTime.Add(5 * time.Minute)) {
			Fail("Service Binding in progress for more than 5 minutes - failing")
		}
		time.Sleep(time.Second)
		bindings, _, err := client.GetServiceBindings(ccv2.Filter{
			Type:     constant.AppGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{apps[0].GUID},
		}, ccv2.Filter{
			Type:     constant.ServiceInstanceGUIDFilter,
			Operator: constant.EqualOperator,
			Values:   []string{serviceInstances[0].GUID},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(bindings).To(HaveLen(1))
		binding = bindings[0]
	}
	Expect(binding.LastOperation.State).To(Equal(constant.LastOperationSucceeded))
}
