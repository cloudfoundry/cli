package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Sidecar", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetProcessSidecars", func() {
		var (
			processSidecars []resources.Sidecar
			warnings        []string
			err             error
		)

		JustBeforeEach(func() {
			processSidecars, warnings, err = client.GetProcessSidecars("some-process-guid")
		})

		When("the process has sidecars", func() {
			BeforeEach(func() {
				response := `{
"resources": [
    {
      "guid": "process-1-guid",
      "name": "auth-sidecar",
      "command": "bundle exec rackup",
      "process_types": ["web", "worker"],
      "memory_in_mb": 300,
      "relationships": {
        "app": {
          "data": {
            "guid": "process-1-guid"
          }
        }
      },
      "created_at": "2017-02-01T01:33:58Z",
      "updated_at": "2017-02-01T01:33:58Z"
    },
    {
      "guid": "process-2-guid",
      "name": "echo-sidecar",
      "command": "start-echo-server",
      "process_types": ["web"],
      "memory_in_mb": 300,
      "relationships": {
        "app": {
          "data": {
            "guid": "process-2-guid"
          }
        }
      },
      "created_at": "2017-02-01T01:33:59Z",
      "updated_at": "2017-02-01T01:33:59Z"
    }
  ]
}
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/sidecars"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the sidecars and all warnings", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(warnings).To(ConsistOf("this is a warning"))
				Expect(len(processSidecars)).To(Equal(2))
				Expect(processSidecars[0]).To(MatchAllFields(Fields{
					"GUID":    Equal("process-1-guid"),
					"Name":    Equal("auth-sidecar"),
					"Command": Equal(types.FilteredString{IsSet: true, Value: "bundle exec rackup"}),
				}))
				Expect(processSidecars[1]).To(MatchAllFields(Fields{
					"GUID":    Equal("process-2-guid"),
					"Name":    Equal("echo-sidecar"),
					"Command": Equal(types.FilteredString{IsSet: true, Value: "start-echo-server"}),
				}))
			})
		})

		When("the process has no sidecars", func() {
			BeforeEach(func() {
				response := `{
					"resources": []
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/sidecars"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("does not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("the process does not exist", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"detail": "Process not found",
							"title": "CF-ResourceNotFound",
							"code": 10010
						}
					]
				}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/sidecars"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)

			})

			It("returns an error and warnings", func() {
				Expect(err).To(MatchError(ccerror.ProcessNotFoundError{}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})

		When("the cloud controller returns errors and warnings", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10008,
							"detail": "The request is semantically invalid: command presence",
							"title": "CF-UnprocessableEntity"
						},
						{
							"code": 10009,
							"detail": "Some CC Error",
							"title": "CF-SomeNewError"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/processes/some-process-guid/sidecars"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10009,
							Detail: "Some CC Error",
							Title:  "CF-SomeNewError",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
