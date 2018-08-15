package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Application Instance Status", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetApplicationApplicationInstanceStatuses", func() {
		var (
			appGUID string

			instances  map[int]ApplicationInstanceStatus
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			instances, warnings, executeErr = client.GetApplicationApplicationInstanceStatuses(appGUID)
		})

		When("the app is found", func() {
			BeforeEach(func() {
				response := `{
					"0": {
						"state": "RUNNING",
						"isolation_segment": "some-isolation-segment",
						"stats": {
							"usage": {
								"disk": 66392064,
								"mem": 29880320,
								"cpu": 0.13511219703079957,
								"time": "2014-06-19 22:37:58 +0000"
							},
							"name": "app_name",
							"uris": [
								"app_name.example.com"
							],
							"host": "10.0.0.1",
							"port": 61035,
							"uptime": 65007,
							"mem_quota": 536870912,
							"disk_quota": 1073741824,
							"fds_quota": 16384
						}
					},
					"1": {
						"state": "STARTING",
						"isolation_segment": "some-isolation-segment",
						"stats": {
							"usage": {
								"disk": 66392064,
								"mem": 29880320,
								"cpu": 0.13511219703079957,
								"time": "2014-06-19 22:37:58 +0000"
							},
							"name": "app_name",
							"uris": [
								"app_name.example.com"
							],
							"host": "10.0.0.1",
							"port": 61035,
							"uptime": 65007,
							"mem_quota": 536870912,
							"disk_quota": 1073741824,
							"fds_quota": 16384
						}
					}
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/stats"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the app instances and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(instances).To(Equal(
					map[int]ApplicationInstanceStatus{
						0: {
							CPU:              0.13511219703079957,
							Disk:             66392064,
							DiskQuota:        1073741824,
							ID:               0,
							IsolationSegment: "some-isolation-segment",
							Memory:           29880320,
							MemoryQuota:      536870912,
							State:            constant.ApplicationInstanceRunning,
							Uptime:           65007,
						},
						1: {
							CPU:              0.13511219703079957,
							Disk:             66392064,
							DiskQuota:        1073741824,
							ID:               1,
							IsolationSegment: "some-isolation-segment",
							Memory:           29880320,
							MemoryQuota:      536870912,
							State:            constant.ApplicationInstanceStarting,
							Uptime:           65007,
						},
					},
				))

				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})

		When("the client returns an error", func() {
			BeforeEach(func() {
				response := `{
					"code": 100004,
					"description": "The app could not be found: some-app-guid",
					"error_code": "CF-AppNotFound"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/apps/some-app-guid/stats"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The app could not be found: some-app-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
			})
		})
	})
})
