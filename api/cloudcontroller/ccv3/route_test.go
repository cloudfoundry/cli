package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Route", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("Create route with shared domains", func() {
		var (
			route      Route
			warnings   Warnings
			executeErr error
			spaceGUID  string
			domainGUID string
		)

		JustBeforeEach(func() {
			spaceGUID = "space-guid"
			domainGUID = "domain-guid"
			route, warnings, executeErr = client.CreateRoute(Route{SpaceGUID: spaceGUID, DomainGUID: domainGUID})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := `{ 
  "guid": "some-route-guid",
  "relationships": {
    "space": {
	  "data": { "guid": "space-guid" }
    },
    "domain": {
	  "data": { "guid": "domain-guid" }
    }
  }
}`

				expectedBody := `{
  "relationships": {
  	"space": {
      "data": { "guid": "space-guid" }
    },
    "domain": {
	  "data": { "guid": "domain-guid" }
    }
  }
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/routes"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given route and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))

				Expect(route).To(Equal(Route{
					GUID:       "some-route-guid",
					SpaceGUID:  "space-guid",
					DomainGUID: "domain-guid",
				}))
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
      "detail": "Isolation segment not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/routes"),
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
							Detail: "Isolation segment not found",
							Title:  "CF-ResourceNotFound",
						},
					},
				}))
				Expect(warnings).To(ConsistOf("this is a warning"))
			})
		})
	})

})
