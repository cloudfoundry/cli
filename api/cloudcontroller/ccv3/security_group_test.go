package ccv3_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/resources"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Role", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateSecurityGroup()", func() {
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
			})
		})

		When("the request succeeds", func() {
			When("no additional flags", func() {
				BeforeEach(func() {
					response := `{
						"guid": "some-group-guid",
						"name": "some-group-name"
					}`

					expectedBody := `{
						"name": "some-group-name"
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
					Expect(warnings).To(ConsistOf("warning-1"))

					Expect(createdSecurityGroup).To(Equal(resources.SecurityGroup{
						GUID: "some-group-guid",
						Name: securityGroupName,
					}))
				})
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
						VerifyRequest(http.MethodPost, "/v3/roles"),
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
