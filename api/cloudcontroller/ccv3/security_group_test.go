package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("SecurityGroup", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateSecurityGroup", func() {
		var (
			securityGroupName string

			createdSecurityGroup resources.SecurityGroup
			warnings             Warnings
			executeErr           error
		)

		BeforeEach(func() {
			securityGroupName = "some-group-name"
		})

		JustBeforeEach(func() {
			createdSecurityGroup, warnings, executeErr = client.CreateSecurityGroup(resources.SecurityGroup{
				Name: securityGroupName,
				Rules: []resources.Rule{
					{
						Protocol:    "tcp",
						Destination: "10.0.10.0/24",
					},
				},
			})
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response := fmt.Sprintf(`{
  "guid": "some-group-guid",
  "created_at": "2016-06-08T16:41:39Z",
  "updated_at": "2016-06-08T16:41:39Z",
  "name": "%s",
  "globally_enabled": {
    "running": false,
    "staging": false
  },
  "rules": [
    {
      "protocol": "tcp",
      "destination": "10.10.10.0/24"
    }
  ],
  "relationships": {
    "staging_spaces": {
      "data": []
    },
    "running_spaces": {
      "data": []
    }
  },
  "links": {
    "self": { "href": "https://api.example.org/v3/security_groups/985c09c5-bf5a-44eb-a260-41c532dc0f1d" }
  }
}
`, securityGroupName)

				expectedBody := `{
						"name": "some-group-name",
						"rules": [
							{
								"protocol": "tcp",
       							"destination": "10.0.10.0/24"
							}
						]
					}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/security_groups"),
						VerifyJSON(expectedBody),
						RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					),
				)
			})

			It("returns the given role and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"warning-1"}))

				Expect(createdSecurityGroup).To(Equal(resources.SecurityGroup{
					GUID: "some-group-guid",
					Name: securityGroupName,
					Rules: []resources.Rule{
						{
							Protocol:    "tcp",
							Destination: "10.10.10.0/24",
						},
					},
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
						VerifyRequest(http.MethodPost, "/v3/security_groups"),
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

	Describe("GetSecurityGroups", func() {
		var (
			returnedSecurityGroup []resources.SecurityGroup
			query                 Query
			warnings              Warnings
			executeErr            error
		)

		JustBeforeEach(func() {
			returnedSecurityGroup, warnings, executeErr = client.GetSecurityGroups(query)
		})

		When("the request succeeds", func() {
			BeforeEach(func() {
				response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/security_groups?names=some-group-name&page=2&per_page=2"
		}
	},
  "resources": [
    {
      	"name": "security-group-name-1",
      	"guid": "security-group-guid-1"
    },
    {
      	"name": "security-group-name-2",
      	"guid": "security-group-guid-2"
    }
  ]
}`, server.URL())
				response2 := `{
	"pagination": {
		"next": null
	},
	"resources": [
	  {
		"name": "security-group-name-3",
		  "guid": "security-group-guid-3"
		}
	]
}`

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/security_groups", "names=some-group-name"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/security_groups", "names=some-group-name&page=2&per_page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)

				query = Query{
					Key:    NameFilter,
					Values: []string{"some-group-name"},
				}
			})

			It("returns the given role and all warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(Equal(Warnings{"this is a warning", "this is another warning"}))

				Expect(returnedSecurityGroup).To(Equal([]resources.SecurityGroup{{
					GUID: "security-group-guid-1",
					Name: "security-group-name-1",
				}, {
					GUID: "security-group-guid-2",
					Name: "security-group-name-2",
				}, {
					GUID: "security-group-guid-3",
					Name: "security-group-name-3",
				}}))
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
						VerifyRequest(http.MethodGet, "/v3/security_groups"),
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
