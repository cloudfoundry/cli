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

var _ = Describe("Role", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("CreateRole", func() {
		var (
			roleType  constant.RoleType
			userGUID  string
			userName  string
			origin    string
			orgGUID   string
			spaceGUID string

			createdRole Role
			warnings    Warnings
			executeErr  error
		)

		BeforeEach(func() {
			userGUID = ""
			userName = ""
			origin = "uaa"
			orgGUID = ""
			spaceGUID = ""
		})

		JustBeforeEach(func() {
			createdRole, warnings, executeErr = client.CreateRole(Role{
				Type:      roleType,
				UserGUID:  userGUID,
				UserName:  userName,
				Origin:    origin,
				OrgGUID:   orgGUID,
				SpaceGUID: spaceGUID,
			})
		})

		Describe("create org role by username/origin", func() {
			BeforeEach(func() {
				roleType = constant.OrgAuditorRole
				userName = "user-name"
				origin = "uaa"
				orgGUID = "org-guid"
			})

			When("the request succeeds", func() {
				When("no additional flags", func() {
					BeforeEach(func() {
						response := `{
							"guid": "some-role-guid",
							"type": "organization_auditor",
							"relationships": {
								"organization": {
									"data": { "guid": "org-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						expectedBody := `{
							"type": "organization_auditor",
							"relationships": {
								"organization": {
									"data": { "guid": "org-guid" }
								},
								"user": {
									"data": { "name": "user-name", "origin": "uaa" }
								}
							}
						}`

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPost, "/v3/roles"),
								VerifyJSON(expectedBody),
								RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
							),
						)
					})

					It("returns the given route and all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1"))

						Expect(createdRole).To(Equal(Role{
							GUID:     "some-role-guid",
							Type:     constant.OrgAuditorRole,
							UserGUID: "user-guid",
							OrgGUID:  "org-guid",
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

		Describe("create space role by username/origin", func() {
			BeforeEach(func() {
				roleType = constant.SpaceAuditorRole
				userName = "user-name"
				spaceGUID = "space-guid"
			})

			When("the request succeeds", func() {
				When("no additional flags", func() {
					BeforeEach(func() {
						response := `{
							"guid": "some-role-guid",
							"type": "space_auditor",
							"relationships": {
								"space": {
									"data": { "guid": "space-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						expectedBody := `{
							"type": "space_auditor",
							"relationships": {
								"space": {
									"data": { "guid": "space-guid" }
								},
								"user": {
									"data": { "name": "user-name", "origin": "uaa" }
								}
							}
						}`

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPost, "/v3/roles"),
								VerifyJSON(expectedBody),
								RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
							),
						)
					})

					It("returns the given route and all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1"))

						Expect(createdRole).To(Equal(Role{
							GUID:      "some-role-guid",
							Type:      constant.SpaceAuditorRole,
							UserGUID:  "user-guid",
							SpaceGUID: "space-guid",
						}))
					})
				})
			})
		})

		Describe("create org role by guid", func() {
			BeforeEach(func() {
				roleType = constant.OrgAuditorRole
				userGUID = "user-guid"
				orgGUID = "org-guid"
			})

			When("the request succeeds", func() {
				When("no additional flags", func() {
					BeforeEach(func() {
						response := `{
							"guid": "some-role-guid",
							"type": "organization_auditor",
							"relationships": {
								"organization": {
									"data": { "guid": "org-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						expectedBody := `{
							"type": "organization_auditor",
							"relationships": {
								"organization": {
									"data": { "guid": "org-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPost, "/v3/roles"),
								VerifyJSON(expectedBody),
								RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
							),
						)
					})

					It("returns the given route and all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1"))

						Expect(createdRole).To(Equal(Role{
							GUID:     "some-role-guid",
							Type:     constant.OrgAuditorRole,
							UserGUID: "user-guid",
							OrgGUID:  "org-guid",
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

		Describe("create space role by guid", func() {
			BeforeEach(func() {
				roleType = constant.SpaceAuditorRole
				userGUID = "user-guid"
				spaceGUID = "space-guid"
			})

			When("the request succeeds", func() {
				When("no additional flags", func() {
					BeforeEach(func() {
						response := `{
							"guid": "some-role-guid",
							"type": "space_auditor",
							"relationships": {
								"space": {
									"data": { "guid": "space-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						expectedBody := `{
							"type": "space_auditor",
							"relationships": {
								"space": {
									"data": { "guid": "space-guid" }
								},
								"user": {
									"data": { "guid": "user-guid" }
								}
							}
						}`

						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodPost, "/v3/roles"),
								VerifyJSON(expectedBody),
								RespondWith(http.StatusCreated, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
							),
						)
					})

					It("returns the given route and all warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1"))

						Expect(createdRole).To(Equal(Role{
							GUID:      "some-role-guid",
							Type:      constant.SpaceAuditorRole,
							UserGUID:  "user-guid",
							SpaceGUID: "space-guid",
						}))
					})
				})
			})
		})
	})
})
