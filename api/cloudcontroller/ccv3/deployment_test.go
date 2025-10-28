package ccv3_test

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v8/resources"
    "code.cloudfoundry.org/cli/v8/types"

	. "code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("Deployment", func() {
	var client *Client

	BeforeEach(func() {
		client, _ = NewTestClient()
	})

	Describe("GetDeployments", func() {
		var (
			deployments []resources.Deployment
			warnings    Warnings
			executeErr  error
		)

		JustBeforeEach(func() {
			deployments, warnings, executeErr = client.GetDeployments(Query{Key: AppGUIDFilter, Values: []string{"some-app-guid"}}, Query{Key: OrderBy, Values: []string{"-created_at"}}, Query{Key: PerPage, Values: []string{"1"}})
		})

		var response string
		var response2 string
		BeforeEach(func() {
			response = fmt.Sprintf(`{
	"pagination": {
		"next": {
			"href": "%s/v3/deployments?app_guids=some-app-guid&order_by=-created_at&page=2&per_page=1"
		}
	},
	"resources": [
		{
      		"guid": "newest-deployment-guid",
      		"created_at": "2018-05-25T22:42:10Z"
   	}
	]
}`, server.URL())
			response2 = `{
  	"pagination": {
		"next": null
	},
	"resources": [
		{
      		"guid": "oldest-deployment-guid",
      		"created_at": "2018-04-25T22:42:10Z"
   	}
	]
}`
		})

		Context("when the deployment exists", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments", "app_guids=some-app-guid&order_by=-created_at&per_page=1"),
						RespondWith(http.StatusAccepted, response),
					),
				)

				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments", "app_guids=some-app-guid&order_by=-created_at&page=2&per_page=1"),
						RespondWith(http.StatusAccepted, response2, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)

			})

			It("returns the deployment guid of the most recent deployment", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(deployments).To(ConsistOf(
					resources.Deployment{GUID: "newest-deployment-guid", CreatedAt: "2018-05-25T22:42:10Z"},
					resources.Deployment{GUID: "oldest-deployment-guid", CreatedAt: "2018-04-25T22:42:10Z"},
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
						VerifyRequest(http.MethodGet, "/v3/deployments", "app_guids=some-app-guid&order_by=-created_at&per_page=1"),
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

	Describe("CancelDeployment", func() {
		var (
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = client.CancelDeployment("some-deployment-guid")
		})

		Context("when the deployment exists", func() {
			Context("when cancelling the deployment succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments/some-deployment-guid/actions/cancel"),
							RespondWith(http.StatusAccepted, "", http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("cancels the deployment with no errors and returns all warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})
			})
		})
	})

	Describe("Create a deployment with app and droplet guids", func() {
		var (
			deploymentGUID string
			warnings       Warnings
			executeErr     error
			dropletGUID    string
		)

		JustBeforeEach(func() {
			var dep resources.Deployment
			dep.Strategy = constant.DeploymentStrategyRolling
			dep.DropletGUID = dropletGUID
			dep.Relationships = resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"}}
			deploymentGUID, warnings, executeErr = client.CreateApplicationDeployment(dep)
		})

		Context("when the application exists", func() {
			var response string
			BeforeEach(func() {
				dropletGUID = "some-droplet-guid"
				response = `{
  "guid": "some-deployment-guid",
  "created_at": "2018-04-25T22:42:10Z",
	"strategy": "rolling",
  "relationships": {
    "app": {
      "data": {
        "guid": "some-app-guid"
      }
    }
  }
}`
			})

			Context("when creating the deployment succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments"),
							VerifyJSON(`{"droplet":{ "guid":"some-droplet-guid" }, "strategy": "rolling", "relationships":{"app":{"data":{"guid":"some-app-guid"}}}}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates the deployment with no errors and returns all warnings", func() {
					Expect(deploymentGUID).To(Equal("some-deployment-guid"))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})

			})

			Context("when no droplet guid is provided", func() {
				BeforeEach(func() {
					dropletGUID = ""
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments"),
							VerifyJSON(`{"strategy":"rolling", "relationships":{"app":{"data":{"guid":"some-app-guid"}}}}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("omits the droplet object in the JSON", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})
			})
		})
	})

	Describe("Create a deployment with app and revision guids", func() {
		var (
			deploymentGUID string
			warnings       Warnings
			executeErr     error
			revisionGUID   string
		)

		JustBeforeEach(func() {
			var dep resources.Deployment
			dep.Strategy = constant.DeploymentStrategyCanary
			dep.RevisionGUID = revisionGUID
			dep.Relationships = resources.Relationships{constant.RelationshipTypeApplication: resources.Relationship{GUID: "some-app-guid"}}
			dep.Options.CanaryDeploymentOptions = &resources.CanaryDeploymentOptions{Steps: []resources.CanaryStep{{InstanceWeight: 1}, {InstanceWeight: 2}}}
			dep.Options.Instances = types.NullInt{IsSet: true, Value: 2}
			dep.Options.MemoryInMB = types.NullUint64{IsSet: true, Value: 1024}
			dep.Options.DiskInMB = types.NullUint64{IsSet: true, Value: 2048}
			dep.Options.LogRateLimitInBPS = types.NullInt{IsSet: true, Value: 10}
			deploymentGUID, warnings, executeErr = client.CreateApplicationDeployment(dep)
		})

		Context("when the application exists", func() {
			var response string
			BeforeEach(func() {
				revisionGUID = "some-revision-guid"
				response = `{
  "guid": "some-deployment-guid",
  "created_at": "2018-04-25T22:42:10Z",
	"strategy": "canary",
  "relationships": {
    "app": {
      "data": {
        "guid": "some-app-guid"
      }
    }
  }
}`
			})

			Context("when creating the deployment succeeds", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodPost, "/v3/deployments"),
							VerifyJSON(`{"revision":{ "guid":"some-revision-guid" }, "strategy": "canary", "relationships":{"app":{"data":{"guid":"some-app-guid"}}},"options":{"canary": {"steps": [{"instance_weight": 1}, {"instance_weight": 2}]},"web_instances": 2,"memory_in_mb": 1024,"disk_in_mb": 2048,"log_rate_limit_in_bytes_per_second": 10}}`),
							RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
						),
					)
				})

				It("creates the deployment with no errors and returns all warnings", func() {
					Expect(deploymentGUID).To(Equal("some-deployment-guid"))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning"))
				})
			})
		})

		When("the cloud controller version does not support canary deployment", func() {
			BeforeEach(func() {
				revisionGUID = "some-revision-guid"
				response := `{
  "errors": [
    {
      "code": 10008,
      "detail": "Strategy 'canary' is not a supported deployment strategy",
      "title": "CF-UnprocessableEntity"
    }
  ]
}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodPost, "/v3/deployments"),
						VerifyJSON(`{"revision":{ "guid":"some-revision-guid" }, "strategy": "canary","options":{"canary": {"steps": [{"instance_weight": 1}, {"instance_weight": 2}]},"web_instances": 2,"memory_in_mb": 1024,"disk_in_mb": 2048,"log_rate_limit_in_bytes_per_second": 10}, "relationships":{"app":{"data":{"guid":"some-app-guid"}}}}`),
						RespondWith(http.StatusTeapot, response, http.Header{}),
					),
				)
			})

			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
				Expect(executeErr).To(MatchError(ccerror.V3UnexpectedResponseError{
					ResponseCode: http.StatusTeapot,
					V3ErrorResponse: ccerror.V3ErrorResponse{
						Errors: []ccerror.V3Error{{
							Code:   10008,
							Detail: "Strategy 'canary' is not a supported deployment strategy",
							Title:  "CF-UnprocessableEntity",
						}}},
				},
				))
			})
		})
	})

	Describe("GetDeployment", func() {
		var response string
		Context("When the deployments exists", func() {
			BeforeEach(func() {
				response = `{
				    "guid": "some-deployment-guid",
					"state": "DEPLOYED",
					"strategy": "canary",
					"status": {
						"value": "FINALIZED",
						"reason": "SUPERSEDED",
						"canary": {
							"steps": {
								"current": 4,
								"total": 5
							}
   						}
					},
					"droplet": {
 					  "guid": "some-droplet-guid"
					},
 					"previous_droplet": {
 					  "guid": "some-other-droplet-guid"
 					},
 					"created_at": "some-time",
 					"updated_at": "some-later-time",
 					"relationships": {
 					  "app": {
 					    "data": {
 					      "guid": "some-app-guid"
 					    }
 					  }
 					}
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments/some-deployment-guid"),
						RespondWith(http.StatusAccepted, response, http.Header{"X-Cf-Warnings": {"warning"}}),
					),
				)
			})
			It("Successfully returns a deployment object", func() {
				deployment, warnings, err := client.GetDeployment("some-deployment-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning"))
				Expect(deployment).To(Not(BeNil()))
				Expect(deployment.GUID).To(Equal("some-deployment-guid"))
				Expect(deployment.State).To(Equal(constant.DeploymentDeployed))
				Expect(deployment.StatusValue).To(Equal(constant.DeploymentStatusValueFinalized))
				Expect(deployment.StatusReason).To(Equal(constant.DeploymentStatusReasonSuperseded))
				Expect(deployment.Strategy).To(Equal(constant.DeploymentStrategyCanary))
				Expect(deployment.CanaryStatus.Steps.CurrentStep).To(Equal(4))
				Expect(deployment.CanaryStatus.Steps.TotalSteps).To(Equal(5))
			})
		})

		Context("when the deployment doesn't exist", func() {
			BeforeEach(func() {
				response := `{
					"errors": [
						{
							"code": 10010,
							"detail": "Deployment not found",
							"title": "CF-ResourceNotFound"
						}
					]
				}`
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/v3/deployments/not-a-deployment"),
						RespondWith(http.StatusNotFound, response, http.Header{"X-Cf-Warnings": {"warning-deployment"}}),
					),
				)
			})

			It("returns the error", func() {
				_, warnings, err := client.GetDeployment("not-a-deployment")
				Expect(err).To(MatchError(ccerror.DeploymentNotFoundError{}))
				Expect(warnings).To(ConsistOf("warning-deployment"))
			})
		})
	})
})
