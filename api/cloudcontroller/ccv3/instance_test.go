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

	Describe("DeleteInstanceByApplicationProcessTypeAndIndex", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.DeleteApplicationProcessInstance("some-app-guid", "some-process-type", 666)
		})

		Context("when the cloud controller returns an error", func() {
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

			It("returns the error", func() {
				Expect(executeErr).To(MatchError(ccerror.ProcessNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when the delete is successful", func() {
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
		Context("when the process exists", func() {
			BeforeEach(func() {
				response := `{
					"resources": [
						{
							"state": "RUNNING",
							"usage": {
								"cpu": 0.01,
								"mem": 1000000,
								"disk": 2000000
							},
							"mem_quota": 2000000,
							"disk_quota": 4000000,
							"index": 0,
							"uptime": 123
						},
						{
							"state": "RUNNING",
							"usage": {
								"cpu": 0.02,
								"mem": 8000000,
								"disk": 16000000
							},
							"mem_quota": 16000000,
							"disk_quota": 32000000,
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
				processes, warnings, err := client.GetProcessInstances("some-process-guid")
				Expect(err).ToNot(HaveOccurred())

				Expect(processes).To(ConsistOf(
					ProcessInstance{
						State:       constant.ProcessInstanceRunning,
						CPU:         0.01,
						MemoryUsage: 1000000,
						DiskUsage:   2000000,
						MemoryQuota: 2000000,
						DiskQuota:   4000000,
						Index:       0,
						Uptime:      123,
					},
					ProcessInstance{
						State:       constant.ProcessInstanceRunning,
						CPU:         0.02,
						MemoryUsage: 8000000,
						DiskUsage:   16000000,
						MemoryQuota: 16000000,
						DiskQuota:   32000000,
						Index:       1,
						Uptime:      456,
					},
				))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when cloud controller returns an error", func() {
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
						RespondWith(http.StatusNotFound, response),
					),
				)
			})

			It("returns the error", func() {
				_, _, err := client.GetProcessInstances("some-process-guid")
				Expect(err).To(MatchError(ccerror.ProcessNotFoundError{}))
			})
		})
	})

})
