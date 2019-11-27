package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = FDescribe("Role", func() {
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
				Username:  userName,
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
									"data": { "username": "user-name", "origin": "uaa" }
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

					It("returns the given role and all warnings", func() {
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
									"data": { "username": "user-name", "origin": "uaa" }
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
				"detail": "Something was not processable",
				"title": "CF-UnprocessableEntity"
			},
			{
				"code": 10010,
				"detail": "Something was not found",
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
								Detail: "Something was not processable",
								Title:  "CF-UnprocessableEntity",
							},
							{
								Code:   10010,
								Detail: "Something was not found",
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

	Describe("GetRoles", func() {
		var (
			roles      []Role
			includes   IncludedResources
			warnings   Warnings
			executeErr error
			query      []Query
		)

		BeforeEach(func() {
			query = []Query{
				{
					Key:    OrganizationGUIDFilter,
					Values: []string{"some-org-name"},
				},
				{
					Key:    Include,
					Values: []string{"users"},
				},
			}
		})
		JustBeforeEach(func() {
			roles, includes, warnings, executeErr = client.GetRoles(query...)
		})

		Describe("listing roles", func() {
			When("the request succeeds", func() {
				BeforeEach(func() {
					response1 := fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/roles?organization_guids=some-org-name&page=2&per_page=1&include=users"
		}
	},
  "resources": [
    {
      "guid": "role-guid-1",
      "type": "organization_user"
    }
  ]
}`, server.URL())
					response2 := `{
							"pagination": {
								"next": null
							},
						 "resources": [
						   {
						     "guid": "role-guid-2",
						     "type": "organization_manager"
						   }
						 ]
						}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&include=users"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&page=2&per_page=1&include=users"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(roles).To(Equal([]Role{{
						GUID: "role-guid-1",
						Type: constant.OrgUserRole,
					}, {
						GUID: "role-guid-2",
						Type: constant.OrgManagerRole,
					}}))
				})
			})

			When("the request uses the `include` query key", func() {
				BeforeEach(func() {
					response1 := fmt.Sprintf(`{
						"pagination": {
							"next": {
								"href": "%s/v3/roles?organization_guids=some-org-name&page=2&per_page=1&include=users"
							}
						},
						"resources": [
							{
							  "guid": "role-guid-1",
							  "type": "organization_user",
							  "relationships": {
								"user": {
								  "data": {"guid": "user-guid-1"}
								}
							  }
							}
						],
						"included": {
							"users": [
								{
									"guid": "user-guid-1",
									"username": "user-name-1",
									"origin": "uaa"
							  	}
							]
						}
}`, server.URL())
					response2 := `{
							"pagination": {
								"next": null
							},
						 "resources": [
						   {
						     "guid": "role-guid-2",
						     "type": "organization_manager",
							  "relationships": {
								"user": {
								  "data": {"guid": "user-guid-2"}
								}
							  }
						   }
						 ],
						"included": {
							"users": [
							  {
								"guid": "user-guid-2",
								"username": "user-name-2",
								"origin": "uaa"
							  }
							]
						  }
						}`

					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&include=users"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						),
					)
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles", "organization_guids=some-org-name&page=2&per_page=1&include=users"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						),
					)
				})

				It("returns the given route and all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))

					Expect(roles).To(Equal([]Role{{
						GUID:     "role-guid-1",
						Type:     constant.OrgUserRole,
						UserGUID: "user-guid-1",
					}, {
						GUID:     "role-guid-2",
						Type:     constant.OrgManagerRole,
						UserGUID: "user-guid-2",
					}}))

					Expect(includes).To(Equal(IncludedResources{
						Users: []User{
							{GUID: "user-guid-1", Username: "user-name-1", Origin: "uaa"},
							{GUID: "user-guid-2", Username: "user-name-2", Origin: "uaa"},
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
      "detail": "Org not found",
      "title": "CF-ResourceNotFound"
    }
  ]
}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v3/roles"),
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
								Detail: "Org not found",
								Title:  "CF-ResourceNotFound",
							},
						},
					}))
					Expect(warnings).To(ConsistOf("this is a warning"))
				})
			})
		})
	})

	Describe("DeleteRoles", func() {
		var (
			roleGUID string
			jobURL 	JobURL
			jobURLString	 string
			warnings	Warnings
			executeErr 	error
		)

		BeforeEach(func() {
			roleGUID = "role-guid"
		})

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = client.DeleteRole(roleGUID)
		})

		When("role exists", func() {
			roleGUID = "role-guid"
			jobURLString = "https://api.test.com/v3/jobs/job-guid"

			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v3/roles/role-guid"),
						RespondWith(http.StatusAccepted, nil, http.Header{
							"X-Cf-Warnings": {"this is a warning"},
							"Location":      {jobURLString},
						}),
					),
				)
			})

			It("returns all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(jobURL).To(Equal(JobURL(jobURLString)))
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
										  "code": 10010,
										  "detail": "Isolation segment not found",
										  "title": "CF-ResourceNotFound"
										}
									  ]
									}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodDelete, "/v3/roles/role-guid"),
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
