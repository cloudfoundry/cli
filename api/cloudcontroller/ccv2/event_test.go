package ccv2_test

import (
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Event", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("GetEvents", func() {
		var (
			events     []Event
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			events, warnings, executeErr = client.GetEvents(
				Filter{
					Type:     constant.TimestampFilter,
					Operator: constant.GreaterThanOperator,
					Values:   []string{"2015-03-10T23:11:54Z"},
				},
				Filter{
					Type:     constant.TypeFilter,
					Operator: constant.InOperator,
					Values:   []string{"audit.app.create", "audit.app.delete-request"},
				},
			)
		})
		Context("when getting events errors", func() {
			BeforeEach(func() {
				response := `{
					"code": 1,
					"description": "some error description",
					"error_code": "CF-SomeError"
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/events", "q=timestamp>2015-03-10T23:11:54Z&q=type+IN+audit.app.create,audit.app.delete-request"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError(ccerror.V2UnexpectedResponseError{
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        1,
						Description: "some error description",
						ErrorCode:   "CF-SomeError",
					},
					ResponseCode: http.StatusTeapot,
				}))

				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		Context("when getting events succeeds", func() {
			BeforeEach(func() {
				response1 := `{
				"next_url": "/v2/events?q=timestamp>2015-03-10T23:11:54Z&q=type+IN+audit.app.create,audit.app.delete-request&page=2",
				"resources": [
					{
						"metadata": {
							"guid": "some-event-guid-1",
							"updated_at": null
						},
						"entity": {
							"type": "audit.app.create",
							"actor": "some-actor-guid",
							"actor_type": "some-actor-type",
							"actor_name": "some-actor-name",
							"actee": "some-actee-guid",
							"actee_type": "some-actee-type",
							"actee_name": "some-actee-name",
							"timestamp": "2015-03-10T23:11:54Z",
							"metadata": {
								"route_mapping_guid": "some-route-mapping-guid"
							}
						}
					},
					{
						"metadata": {
							"guid": "some-event-guid-2",
							"updated_at": null
						},
						"entity": {
							"type": "audit.app.delete-request",
							"actor": "some-actor-guid",
							"actor_type": "some-actor-type",
							"actor_name": "some-actor-name",
							"actee": "some-actee-guid",
							"actee_type": "some-actee-type",
							"actee_name": "some-actee-name",
							"timestamp": "2015-03-10T23:11:54Z"
						}
					}
				]
			}`
				response2 := `{
				"next_url": null,
				"resources": [
					{
						"metadata": {
							"guid": "some-event-guid-3",
							"updated_at": null
						},
						"entity": {
							"type": "audit.app.create",
							"actor": "some-actor-guid",
							"actor_type": "some-actor-type",
							"actor_name": "some-actor-name",
							"actee": "some-actee-guid",
							"actee_type": "some-actee-type",
							"actee_name": "some-actee-name",
							"timestamp": "2015-03-10T23:11:54Z"
						}
					},
					{
						"metadata": {
							"guid": "some-event-guid-4",
							"updated_at": null
						},
						"entity": {
							"type": "audit.app.delete-request",
							"actor": "some-actor-guid",
							"actor_type": "some-actor-type",
							"actor_name": "some-actor-name",
							"actee": "some-actee-guid",
							"actee_type": "some-actee-type",
							"actee_name": "some-actee-name",
							"timestamp": "2015-03-10T23:11:54Z"
						}
					}
				]
			}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/events", "q=timestamp>2015-03-10T23:11:54Z&q=type+IN+audit.app.create,audit.app.delete-request"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/events", "q=timestamp>2015-03-10T23:11:54Z&q=type+IN+audit.app.create,audit.app.delete-request&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-3, warning-4"}}),
					),
				)
			})

			It("returns all the events", func() {
				expectedTimestamp, err := time.Parse(time.RFC3339, "2015-03-10T23:11:54Z")
				Expect(err).ToNot(HaveOccurred())

				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				Expect(events).To(Equal([]Event{
					{
						GUID:      "some-event-guid-1",
						Type:      constant.EventTypeAuditApplicationCreate,
						ActorGUID: "some-actor-guid",
						ActorType: "some-actor-type",
						ActorName: "some-actor-name",
						ActeeGUID: "some-actee-guid",
						ActeeType: "some-actee-type",
						ActeeName: "some-actee-name",
						Timestamp: expectedTimestamp,
						Metadata: map[string]interface{}{
							"route_mapping_guid": "some-route-mapping-guid",
						},
					},
					{
						GUID:      "some-event-guid-2",
						Type:      constant.EventTypeAuditApplicationDelete,
						ActorGUID: "some-actor-guid",
						ActorType: "some-actor-type",
						ActorName: "some-actor-name",
						ActeeGUID: "some-actee-guid",
						ActeeType: "some-actee-type",
						ActeeName: "some-actee-name",
						Timestamp: expectedTimestamp,
						Metadata:  nil,
					},
					{
						GUID:      "some-event-guid-3",
						Type:      constant.EventTypeAuditApplicationCreate,
						ActorGUID: "some-actor-guid",
						ActorType: "some-actor-type",
						ActorName: "some-actor-name",
						ActeeGUID: "some-actee-guid",
						ActeeType: "some-actee-type",
						ActeeName: "some-actee-name",
						Timestamp: expectedTimestamp,
						Metadata:  nil,
					},
					{
						GUID:      "some-event-guid-4",
						Type:      constant.EventTypeAuditApplicationDelete,
						ActorGUID: "some-actor-guid",
						ActorType: "some-actor-type",
						ActorName: "some-actor-name",
						ActeeGUID: "some-actee-guid",
						ActeeType: "some-actee-type",
						ActeeName: "some-actee-name",
						Timestamp: expectedTimestamp,
						Metadata:  nil,
					},
				}))
			})
		})
	})
})
