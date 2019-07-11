package v2action_test

import (
	"encoding/json"
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("User-Provided Service Instance Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("Updating", func() {
		const serviceInstanceGUID = "service-instance-guid"

		DescribeTable("updating service instance data",
			func(expectation string, instance UserProvidedServiceInstance) {
				fakeCloudControllerClient.UpdateUserProvidedServiceInstanceReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					nil,
				)

				warnings, err := actor.UpdateUserProvidedServiceInstance(serviceInstanceGUID, instance)
				Expect(err).NotTo(HaveOccurred())

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(fakeCloudControllerClient.UpdateUserProvidedServiceInstanceCallCount()).To(Equal(1))
				guid, inst := fakeCloudControllerClient.UpdateUserProvidedServiceInstanceArgsForCall(0)
				Expect(guid).To(Equal(serviceInstanceGUID))

				body, err := json.Marshal(inst)
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(MatchJSON(expectation))
			},
			Entry(
				"setting credentials",
				`{"credentials": {"username": "super-secret-password"}}`,
				UserProvidedServiceInstance{}.WithCredentials(map[string]interface{}{"username": "super-secret-password"}),
			),
			Entry(
				"removing credentials",
				`{"credentials": {}}`,
				UserProvidedServiceInstance{}.WithCredentials(nil),
			),
			Entry(
				"setting route service URL",
				`{"route_service_url": "fake-route-url"}`,
				UserProvidedServiceInstance{}.WithRouteServiceURL("fake-route-url"),
			),
			Entry(
				"removing route service URL",
				`{"route_service_url": ""}`,
				UserProvidedServiceInstance{}.WithRouteServiceURL(""),
			),
			Entry(
				"setting syslog drain URL",
				`{"syslog_drain_url": "fake-syslog-drain-url"}`,
				UserProvidedServiceInstance{}.WithSyslogDrainURL("fake-syslog-drain-url"),
			),
			Entry(
				"removing syslog drain URL",
				`{"syslog_drain_url": ""}`,
				UserProvidedServiceInstance{}.WithSyslogDrainURL(""),
			),
			Entry(
				"setting tags",
				`{"tags": ["foo", "bar"]}`,
				UserProvidedServiceInstance{}.WithTags([]string{"foo", "bar"}),
			),
			Entry(
				"removing tags",
				`{"tags": []}`,
				UserProvidedServiceInstance{}.WithTags(nil),
			),
			Entry(
				"updating everything",
				`{
				   "credentials": {"username": "super-secret-password"},
				   "route_service_url": "fake-route-url",
				   "syslog_drain_url": "fake-syslog-drain-url",
				   "tags": ["foo", "bar"]
				}`,
				UserProvidedServiceInstance{}.
					WithCredentials(map[string]interface{}{"username": "super-secret-password"}).
					WithRouteServiceURL("fake-route-url").
					WithSyslogDrainURL("fake-syslog-drain-url").
					WithTags([]string{"foo", "bar"}),
			),
		)

		When("the update fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateUserProvidedServiceInstanceReturns(
					ccv2.Warnings{"warning-1", "warning-2"},
					errors.New("update failed horribly!!!"),
				)
			})

			It("returns the error and all the warnings", func() {
				warnings, err := actor.UpdateUserProvidedServiceInstance(serviceInstanceGUID, UserProvidedServiceInstance{})
				Expect(err).To(MatchError("update failed horribly!!!"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})
})
