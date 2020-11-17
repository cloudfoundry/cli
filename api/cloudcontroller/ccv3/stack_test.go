package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Stacks", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetStacks", func() {
		var (
			query Query

			stacks     []Stack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			stacks, warnings, executeErr = client.GetStacks(query)
		})

		When("stacks exist", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/stacks?names=some-stack-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      	"name": "stack-name-1",
      	"guid": "stack-guid-1",
      	"description": "stack desc 1"
    },
    {
      	"name": "stack-name-2",
      	"guid": "stack-guid-2",
		"description": "stack desc 2"
    }
  ]
}`, server.URL())
				response2 := `{
	"pagination": {
		"next": null
	},
	"resources": [
	  {
		"name": "stack-name-3",
		  "guid": "stack-guid-3",
		"description": "stack desc 3"
		}
	]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/stacks", "names=some-stack-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/stacks", "names=some-stack-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-stack-name"},
				}
			})

			It("returns the queried stacks and all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())

				Expect(stacks).To(ConsistOf(
					Stack{Name: "stack-name-1", GUID: "stack-guid-1", Description: "stack desc 1"},
					Stack{Name: "stack-name-2", GUID: "stack-guid-2", Description: "stack desc 2"},
					Stack{Name: "stack-name-3", GUID: "stack-guid-3", Description: "stack desc 3"},
				))
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
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
      "code": 10010,
      "detail": "stack not found",
      "title": "CF-stackNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/stacks"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.MultiError{
					ResponseCode: http.StatusTeapot,
					Errors: []ccerror.V3Error{
						{
							Code:   10008,
							Detail: "The request is semantically invalid: command presence",
							Title:  "CF-UnprocessableEntity",
						},
						{
							Code:   10010,
							Detail: "stack not found",
							Title:  "CF-stackNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})
})
