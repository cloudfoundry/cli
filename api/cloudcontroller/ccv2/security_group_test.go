package ccv2_test

import (
	"net/http"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Security Groups", func() {
	var client *Client

	BeforeEach(func() {
		client = NewTestClient()
	})

	Describe("AssociateSpaceWithRunningSecurityGroup", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns all warnings", func() {
				warnings, err := client.AssociateSpaceWithRunningSecurityGroup("security-group-guid", "space-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				warnings, err := client.AssociateSpaceWithRunningSecurityGroup("security-group-guid", "space-guid")

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("AssociateSpaceWithStagingSecurityGroup", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				response := `{}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/staging_spaces/space-guid"),
						RespondWith(http.StatusOK, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns all warnings", func() {
				warnings, err := client.AssociateSpaceWithStagingSecurityGroup("security-group-guid", "space-guid")

				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPut, "/v2/security_groups/security-group-guid/staging_spaces/space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				warnings, err := client.AssociateSpaceWithStagingSecurityGroup("security-group-guid", "space-guid")

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSecurityGroups", func() {
		Context("when no errors are encountered", func() {
			Context("when results are paginated", func() {
				BeforeEach(func() {
					response1 := `{
						"next_url": "/v2/security_groups?q=some-query:some-value&page=2",
						"resources": [
							{
								"metadata": {
									"guid": "security-group-guid-1",
									"url": "/v2/security_groups/security-group-guid-1"
								},
								"entity": {
									"name": "security-group-1",
									"rules": [
									],
									"running_default": false,
									"staging_default": true,
									"spaces_url": "/v2/security_groups/security-group-guid-1/spaces"
								}
							}
						]
					}`
					response2 := `{
						"next_url": null,
						"resources": [
							{
								"metadata": {
									"guid": "security-group-guid-2",
									"url": "/v2/security_groups/security-group-guid-2"
								},
								"entity": {
									"name": "security-group-2",
									"rules": [
									],
									"running_default": true,
									"staging_default": false,
									"spaces_url": "/v2/security_groups/security-group-guid-2/spaces"
								}
							}
						]
					}`
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups", "q=some-query:some-value"),
							RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"warning-1"}}),
						))
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/v2/security_groups", "q=some-query:some-value&page=2"),
							RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"warning-2"}}),
						))
				})

				It("returns paginated results and all warnings", func() {
					securityGroups, warnings, err := client.GetSecurityGroups(QQuery{
						Filter:   "some-query",
						Operator: EqualOperator,
						Values:   []string{"some-value"},
					})

					Expect(err).NotTo(HaveOccurred())
					Expect(securityGroups).To(Equal([]SecurityGroup{
						{
							GUID:           "security-group-guid-1",
							Name:           "security-group-1",
							Rules:          []SecurityGroupRule{},
							RunningDefault: false,
							StagingDefault: true,
						},
						{
							GUID:           "security-group-guid-2",
							Name:           "security-group-2",
							Rules:          []SecurityGroupRule{},
							RunningDefault: true,
							StagingDefault: false,
						},
					}))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				})
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/security_groups"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1, warning-2"}}),
					))
			})

			It("returns an error and all warnings", func() {
				_, warnings, err := client.GetSecurityGroups()

				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})
	})

	Describe("GetSpaceRunningSecurityGroupsBySpace", func() {
		Context("when the space exists", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/spaces/some-space-guid/security_groups?q=some-query:some-value&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "running-security-group-guid-1",
								"updated_at": null
							},
							"entity": {
								"name": "running-security-group-name-1",
								"rules": [
									{
										"protocol": "udp",
										"ports": "8080",
										"description": "description-1",
										"destination": "198.41.191.47/1"
									},
									{
										"protocol": "tcp",
										"ports": "80,443",
										"description": "description-2",
										"destination": "254.41.191.47-254.44.255.255"
									}
								]
							}
						},
						{
							"metadata": {
								"guid": "running-security-group-guid-2",
								"updated_at": null
							},
							"entity": {
								"name": "running-security-group-name-2",
								"rules": [
									{
										"protocol": "udp",
										"ports": "8080",
										"description": "description-3",
										"destination": "198.41.191.47/24"
									},
									{
										"protocol": "tcp",
										"ports": "80,443",
										"description": "description-4",
										"destination": "254.41.191.4-254.44.255.4"
									}
								]
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "running-security-group-guid-3",
								"updated_at": null
							},
							"entity": {
								"name": "running-security-group-name-3",
								"rules": [
									{
										"protocol": "udp",
										"ports": "32767",
										"description": "description-5",
										"destination": "127.0.0.1/32"
									},
									{
										"protocol": "tcp",
										"ports": "8008,4443",
										"description": "description-6",
										"destination": "254.41.191.0-254.44.255.1"
									}
								]
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/security_groups", "q=some-query:some-value"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/security_groups", "q=some-query:some-value&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the running security groups and all warnings", func() {
				securityGroups, warnings, err := client.GetSpaceRunningSecurityGroupsBySpace("some-space-guid", QQuery{
					Filter:   "some-query",
					Operator: EqualOperator,
					Values:   []string{"some-value"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
				Expect(securityGroups).To(ConsistOf(
					SecurityGroup{
						Name: "running-security-group-name-1",
						GUID: "running-security-group-guid-1",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "8080",
								Description: "description-1",
								Destination: "198.41.191.47/1",
							},
							{
								Protocol:    "tcp",
								Ports:       "80,443",
								Description: "description-2",
								Destination: "254.41.191.47-254.44.255.255",
							},
						},
					},
					SecurityGroup{
						Name: "running-security-group-name-2",
						GUID: "running-security-group-guid-2",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "8080",
								Description: "description-3",
								Destination: "198.41.191.47/24",
							},
							{
								Protocol:    "tcp",
								Ports:       "80,443",
								Description: "description-4",
								Destination: "254.41.191.4-254.44.255.4",
							},
						},
					},
					SecurityGroup{
						Name: "running-security-group-name-3",
						GUID: "running-security-group-guid-3",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "32767",
								Description: "description-5",
								Destination: "127.0.0.1/32",
							},
							{
								Protocol:    "tcp",
								Ports:       "8008,4443",
								Description: "description-6",
								Destination: "254.41.191.0-254.44.255.1",
							},
						},
					},
				))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
						"code": 40004,
						"description": "The space could not be found: some-space-guid",
						"error_code": "CF-SpaceNotFound"
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/security_groups"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				securityGroups, warnings, err := client.GetSpaceRunningSecurityGroupsBySpace("some-space-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The space could not be found: some-space-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(securityGroups).To(BeEmpty())
			})
		})
	})

	Describe("GetSpaceStagingSecurityGroupsBySpace", func() {
		Context("when the space exists", func() {
			BeforeEach(func() {
				response1 := `{
					"next_url": "/v2/spaces/some-space-guid/staging_security_groups?q=some-query:some-value&page=2",
					"resources": [
						{
							"metadata": {
								"guid": "staging-security-group-guid-1",
								"updated_at": null
							},
							"entity": {
								"name": "staging-security-group-name-1",
								"rules": [
									{
										"protocol": "udp",
										"ports": "8080",
										"description": "description-1",
										"destination": "198.41.191.47/1"
									},
									{
										"protocol": "tcp",
										"ports": "80,443",
										"description": "description-2",
										"destination": "254.41.191.47-254.44.255.255"
									}
								]
							}
						},
						{
							"metadata": {
								"guid": "staging-security-group-guid-2",
								"updated_at": null
							},
							"entity": {
								"name": "staging-security-group-name-2",
								"rules": [
									{
										"protocol": "udp",
										"ports": "8080",
										"description": "description-3",
										"destination": "198.41.191.47/24"
									},
									{
										"protocol": "tcp",
										"ports": "80,443",
										"description": "description-4",
										"destination": "254.41.191.4-254.44.255.4"
									}
								]
							}
						}
					]
				}`
				response2 := `{
					"next_url": null,
					"resources": [
						{
							"metadata": {
								"guid": "staging-security-group-guid-3",
								"updated_at": null
							},
							"entity": {
								"name": "staging-security-group-name-3",
								"rules": [
									{
										"protocol": "udp",
										"ports": "32767",
										"description": "description-5",
										"destination": "127.0.0.1/32"
									},
									{
										"protocol": "tcp",
										"ports": "8008,4443",
										"description": "description-6",
										"destination": "254.41.191.0-254.44.255.1"
									}
								]
							}
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/staging_security_groups", "q=some-query:some-value"),
						RespondWith(http.StatusOK, response1, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/staging_security_groups", "q=some-query:some-value&page=2"),
						RespondWith(http.StatusOK, response2, http.Header{"X-Cf-Warnings": {"this is another warning"}}),
					),
				)
			})

			It("returns the staging security groups and all warnings", func() {
				securityGroups, warnings, err := client.GetSpaceStagingSecurityGroupsBySpace("some-space-guid", QQuery{
					Filter:   "some-query",
					Operator: EqualOperator,
					Values:   []string{"some-value"},
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("this is a warning", "this is another warning"))
				Expect(securityGroups).To(ConsistOf(
					SecurityGroup{
						Name: "staging-security-group-name-1",
						GUID: "staging-security-group-guid-1",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "8080",
								Description: "description-1",
								Destination: "198.41.191.47/1",
							},
							{
								Protocol:    "tcp",
								Ports:       "80,443",
								Description: "description-2",
								Destination: "254.41.191.47-254.44.255.255",
							},
						},
					},
					SecurityGroup{
						Name: "staging-security-group-name-2",
						GUID: "staging-security-group-guid-2",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "8080",
								Description: "description-3",
								Destination: "198.41.191.47/24",
							},
							{
								Protocol:    "tcp",
								Ports:       "80,443",
								Description: "description-4",
								Destination: "254.41.191.4-254.44.255.4",
							},
						},
					},
					SecurityGroup{
						Name: "staging-security-group-name-3",
						GUID: "staging-security-group-guid-3",
						Rules: []SecurityGroupRule{
							{
								Protocol:    "udp",
								Ports:       "32767",
								Description: "description-5",
								Destination: "127.0.0.1/32",
							},
							{
								Protocol:    "tcp",
								Ports:       "8008,4443",
								Description: "description-6",
								Destination: "254.41.191.0-254.44.255.1",
							},
						},
					},
				))
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				response := `{
						"code": 40004,
						"description": "The space could not be found: some-space-guid",
						"error_code": "CF-SpaceNotFound"
					}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v2/spaces/some-space-guid/staging_security_groups"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"this is a warning"}}),
					),
				)
			})

			It("returns the error and warnings", func() {
				securityGroups, warnings, err := client.GetSpaceStagingSecurityGroupsBySpace("some-space-guid")
				Expect(err).To(MatchError(ccerror.ResourceNotFoundError{
					Message: "The space could not be found: some-space-guid",
				}))
				Expect(warnings).To(ConsistOf(Warnings{"this is a warning"}))
				Expect(securityGroups).To(BeEmpty())
			})
		})
	})

	Describe("RemoveSpaceFromRunningSecurityGroup", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = client.RemoveSpaceFromRunningSecurityGroup("security-group-guid", "space-guid")
		})

		Context("when the client call is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusOK, nil, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1"}))
			})
		})

		Context("when the client call is unsuccessful", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/security_groups/security-group-guid/spaces/space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("RemoveSpaceFromStagingSecurityGroup", func() {
		var (
			warnings Warnings
			err      error
		)

		JustBeforeEach(func() {
			warnings, err = client.RemoveSpaceFromStagingSecurityGroup("security-group-guid", "space-guid")
		})

		Context("when the client call is successful", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/security_groups/security-group-guid/staging_spaces/space-guid"),
						RespondWith(http.StatusOK, nil, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns all warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf(Warnings{"warning-1"}))
			})
		})

		Context("when the client call is unsuccessful", func() {
			BeforeEach(func() {
				response := `{
  "code": 10001,
  "description": "Some Error",
  "error_code": "CF-SomeError"
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodDelete, "/v2/security_groups/security-group-guid/staging_spaces/space-guid"),
						RespondWith(http.StatusTeapot, response, http.Header{"X-Cf-Warnings": {"warning-1"}}),
					))
			})

			It("returns the error and all warnings", func() {
				Expect(err).To(MatchError(ccerror.V2UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V2ErrorResponse: ccerror.V2ErrorResponse{
						Code:        10001,
						Description: "Some Error",
						ErrorCode:   "CF-SomeError",
					},
				}))
				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})
})
