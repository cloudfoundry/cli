package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("ProcessInstance", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("DeleteApplicationProcessInstance", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteApplicationProcessInstance("some-app-guid", "some-process-type", 666)
		})

		When("the cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Process not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/apps/some-app-guid/processes/some-process-type/instances/666"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ProcessNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("the delete is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/apps/some-app-guid/processes/some-process-type/instances/666"),
						RespondWith(http.StatusNoContent, "", http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("GetProcessInstances", func() {
		var (
			processes  []ProcessInstance
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			processes, warnings, executeErr = client.GetProcessInstances("some-process-guid")
		})

		When("the process exists", func() {
			BeforeEach(func() {
				response := `{
					"resources": [
						{
						  "type": "web",
							"state": "RUNNING",
							"usage": {
								"cpu": 0.01,
								"mem": 1000000,
								"disk": 2000000
							},
							"mem_quota": 2000000,
							"disk_quota": 4000000,
							"isolation_segment": "example_iso_segment",
							"index": 0,
							"uptime": 123,
							"details": "some details"
						},
						{
						  "type": "web",
							"state": "RUNNING",
							"usage": {
								"cpu": 0.02,
								"mem": 8000000,
								"disk": 16000000
							},
							"mem_quota": 16000000,
							"disk_quota": 32000000,
							"isolation_segment": "example_iso_segment",
							"index": 1,
							"uptime": 456
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/stats"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns a list of instances for the given process and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(processes).To(ConsistOf(
					ProcessInstance{
						CPU:              0.01,
						Details:          "some details",
						DiskQuota:        4000000,
						DiskUsage:        2000000,
						Index:            0,
						IsolationSegment: "example_iso_segment",
						MemoryQuota:      2000000,
						MemoryUsage:      1000000,
						State:            constant.ProcessInstanceRunning,
						Type:             "web",
						Uptime:           123,
					},
					ProcessInstance{
						CPU:              0.02,
						DiskQuota:        32000000,
						DiskUsage:        16000000,
						Index:            1,
						IsolationSegment: "example_iso_segment",
						MemoryQuota:      16000000,
						MemoryUsage:      8000000,
						State:            constant.ProcessInstanceRunning,
						Type:             "web",
						Uptime:           456,
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		When("cloud controller returns an error", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Process not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/stats"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.ProcessNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})
})
