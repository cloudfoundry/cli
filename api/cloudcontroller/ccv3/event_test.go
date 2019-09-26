package ccv3_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"fmt"
	"net/http"

	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Task", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetEvents", func() {
		var (
			events     []Event
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			events, warnings, executeErr = client.GetEvents(
				Query{Key: TargetGUIDFilter, Values: []string{"some-target-guid"}},
				Query{Key: OrderBy, Values: []string{"-created_at"}},
				Query{Key: PerPage, Values: []string{"1"}})
		})

		var response string
		BeforeEach(func() {
			response = fmt.Sprintf(`{
  "pagination": {
    "total_results": 3,
    "total_pages": 2,
    "next": {
      "href": "%s/v3/audit_events?page=2&per_page=2"
    },
    "previous": null
  },
  "resources": [
    {
      "guid": "some-event-guid",
      "created_at": "2016-06-08T16:41:23Z",
      "updated_at": "2016-06-08T16:41:26Z",
      "type": "audit.app.update",
      "actor": {
        "guid": "d144abe3-3d7b-40d4-b63f-2584798d3ee5",
        "type": "user",
        "name": "admin"
      },
      "target": {
        "guid": "2e3151ba-9a63-4345-9c5b-6d8c238f4e55",
        "type": "app",
        "name": "my-app"
      },
      "data": {
        "request": {
          "recursive": true
        }
      },
      "space": {
        "guid": "cb97dd25-d4f7-4185-9e6f-ad6e585c207c"
      },
      "organization": {
        "guid": "d9be96f5-ea8f-4549-923f-bec882e32e3c"
      },
      "links": {
        "self": {
          "href": "https://api.example.org/v3/audit_events/a595fe2f-01ff-4965-a50c-290258ab8582"
        }
      }
    }
  ]
}`, server.URL())
		})

		Context("when the event exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/audit_events", "target_guids=some-target-guid&order_by=-created_at&per_page=1"),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),

					),
				)
			})

			It("returns the event guid of the most recent event", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(events).To(ConsistOf(
					Event{
						GUID: "some-event-guid",
						CreatedAt: "2016-06-08T16:41:23Z",
						Type: "audit.app.update",
						ActorName: "admin",
					},
				))
			})
		})

		Context("when the request fails", func() {
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
      "detail": "App not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/audit_events", "target_guids=some-target-guid&order_by=-created_at&per_page=1"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})

			It("returns CC warnings and error", func() {
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
							Detail: "App not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("warning"))
			})
		})
	})
})
