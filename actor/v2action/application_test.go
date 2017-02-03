package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("Application", func() {
		var app Application
		BeforeEach(func() {
			app = Application{}
		})

		Describe("CalculatedBuildpack", func() {
			Context("when buildpack is set", func() {
				BeforeEach(func() {
					app.Buildpack = "foo"
					app.DetectedBuildpack = "bar"
				})

				It("returns back the buildpack", func() {
					Expect(app.CalculatedBuildpack()).To(Equal("foo"))
				})
			})

			Context("only detected buildpack is set", func() {
				BeforeEach(func() {
					app.DetectedBuildpack = "bar"
				})

				It("returns back the detected buildpack", func() {
					Expect(app.CalculatedBuildpack()).To(Equal("bar"))
				})
			})

			Context("neither buildpack or detected buildpack is set", func() {
				It("returns an empty string", func() {
					Expect(app.CalculatedBuildpack()).To(BeEmpty())
				})
			})
		})

		Describe("CalculatedHealthCheckEndpoint", func() {
			var application Application

			Context("when the health check type is http", func() {
				BeforeEach(func() {
					application = Application{
						HealthCheckType:         "http",
						HealthCheckHTTPEndpoint: "/some-endpoint",
					}
				})

				It("returns the endpoint field", func() {
					Expect(application.CalculatedHealthCheckEndpoint()).To(Equal(
						"/some-endpoint"))
				})
			})

			Context("when the health check type is not http", func() {
				BeforeEach(func() {
					application = Application{
						HealthCheckType:         "process",
						HealthCheckHTTPEndpoint: "/some-endpoint",
					}
				})

				It("returns the empty string", func() {
					Expect(application.CalculatedHealthCheckEndpoint()).To(Equal(""))
				})
			})
		})

		Describe("Started", func() {
			Context("when app is started", func() {
				It("returns true", func() {
					Expect(Application{State: ccv2.ApplicationStarted}.Started()).To(BeTrue())
				})
			})

			Context("when app is stopped", func() {
				It("returns false", func() {
					Expect(Application{State: ccv2.ApplicationStopped}.Started()).To(BeFalse())
				})
			})
		})
	})

	Describe("GetApplicationBySpace", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "some-app-guid",
							Name: "some-app",
						},
					},
					ccv2.Warnings{"foo"},
					nil,
				)
			})

			It("returns the application and warnings", func() {
				app, warnings, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(app).To(Equal(Application{
					GUID: "some-app-guid",
					Name: "some-app",
				}))
				Expect(warnings).To(Equal(Warnings{"foo"}))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf([]ccv2.Query{
					ccv2.Query{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-app",
					},
					ccv2.Query{
						Filter:   ccv2.SpaceGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "some-space-guid",
					},
				}))
			})
		})

		Context("when the application does not exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, nil)
			})

			It("returns an ApplicationNotFoundError", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(ApplicationNotFoundError{Name: "some-app"}))
			})
		})

		Context("when the cloud controller client returns an error", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("I am a CloudControllerClient Error")
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, nil, expectedError)
			})

			It("returns the error", func() {
				_, _, err := actor.GetApplicationByNameAndSpace("some-app", "some-space-guid")
				Expect(err).To(MatchError(expectedError))
			})
		})
	})

	Describe("GetRouteApplications", func() {
		Context("when the CC client returns no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{
						{
							GUID: "application-guid",
							Name: "application-name",
						},
					}, ccv2.Warnings{"route-applications-warning"}, nil)
			})
			It("returns the applications bound to the route and warnings", func() {
				applications, warnings, err := actor.GetRouteApplications("route-guid", nil)
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(applications).To(ConsistOf(
					Application{
						GUID: "application-guid",
						Name: "application-name",
					},
				))
			})
		})

		Context("when the CC client returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteApplicationsReturns(
					[]ccv2.Application{}, ccv2.Warnings{"route-applications-warning"}, errors.New("get-route-applications-error"))
			})

			It("returns the error and warnings", func() {
				apps, warnings, err := actor.GetRouteApplications("route-guid", nil)
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("route-guid"))

				Expect(err).To(MatchError("get-route-applications-error"))
				Expect(warnings).To(ConsistOf("route-applications-warning"))
				Expect(apps).To(BeNil())
			})
		})

		Context("when a query parameter exists", func() {
			It("passes the query to the client", func() {
				expectedQuery := []ccv2.Query{
					{
						Filter:   ccv2.RouteGUIDFilter,
						Operator: ccv2.EqualOperator,
						Value:    "route-guid",
					}}

				_, _, err := actor.GetRouteApplications("route-guid", expectedQuery)
				Expect(err).ToNot(HaveOccurred())
				_, query := fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
	})

	Describe("SetApplicationHealthCheckTypeByNameAndSpace", func() {
		Context("when the app exists", func() {
			Context("when the desired health check type is different", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{GUID: "some-app-guid"},
						},
						ccv2.Warnings{"get application warning"},
						nil,
					)
					fakeCloudControllerClient.UpdateApplicationReturns(
						ccv2.Application{
							GUID:            "some-app-guid",
							HealthCheckType: "some-health-check-type",
						},
						ccv2.Warnings{"update warnings"},
						nil,
					)
				})

				It("sets the desired health check type and returns the warnings", func() {
					returnedApp, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
						"some-app", "some-space-guid", "some-health-check-type", "/")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get application warning", "update warnings"))

					Expect(returnedApp).To(Equal(Application{
						GUID:            "some-app-guid",
						HealthCheckType: "some-health-check-type",
					}))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
					app := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
					Expect(app).To(Equal(ccv2.Application{
						GUID:            "some-app-guid",
						HealthCheckType: "some-health-check-type",
					}))
				})
			})

			Context("when the desired health check type is 'http'", func() {
				Context("when the desired http endpoint is already set", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{
								{GUID: "some-app-guid", HealthCheckType: "http", HealthCheckHTTPEndpoint: "/"},
							},
							ccv2.Warnings{"get application warning"},
							nil,
						)
					})

					It("does not send the update", func() {
						_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
							"some-app", "some-space-guid", "http", "/")
						Expect(err).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("get application warning"))

						Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(0))
					})
				})

				Context("when the desired http endpoint is not set", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{
								{GUID: "some-app-guid", HealthCheckType: "http", HealthCheckHTTPEndpoint: "/"},
							},
							ccv2.Warnings{"get application warning"},
							nil,
						)
						fakeCloudControllerClient.UpdateApplicationReturns(
							ccv2.Application{},
							ccv2.Warnings{"update warnings"},
							nil,
						)
					})

					It("sets the desired health check type and returns the warnings", func() {
						_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
							"some-app", "some-space-guid", "http", "/v2/anything")
						Expect(err).ToNot(HaveOccurred())

						Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
						app := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
						Expect(app).To(Equal(ccv2.Application{
							GUID:                    "some-app-guid",
							HealthCheckType:         "http",
							HealthCheckHTTPEndpoint: "/v2/anything",
						}))

						Expect(warnings).To(ConsistOf("get application warning", "update warnings"))
					})
				})
			})

			Context("when the application health check type is already set to the desired type", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{
							{
								GUID:            "some-app-guid",
								HealthCheckType: "some-health-check-type",
							},
						},
						ccv2.Warnings{"get application warning"},
						nil,
					)
				})

				It("does not update the health check type", func() {
					returnedApp, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
						"some-app", "some-space-guid", "some-health-check-type", "/")
					Expect(err).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get application warning"))
					Expect(returnedApp).To(Equal(Application{
						GUID:            "some-app-guid",
						HealthCheckType: "some-health-check-type",
					}))

					Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when getting the application returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{}, ccv2.Warnings{"get application warning"}, errors.New("get application error"))
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
					"some-app", "some-space-guid", "some-health-check-type", "/")

				Expect(warnings).To(ConsistOf("get application warning"))
				Expect(err).To(MatchError("get application error"))
			})
		})

		Context("when updating the application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("foo bar")
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv2.Application{
						{GUID: "some-app-guid"},
					},
					ccv2.Warnings{"get application warning"},
					nil,
				)
				fakeCloudControllerClient.UpdateApplicationReturns(
					ccv2.Application{},
					ccv2.Warnings{"update warnings"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				_, warnings, err := actor.SetApplicationHealthCheckTypeByNameAndSpace(
					"some-app", "some-space-guid", "some-health-check-type", "/")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get application warning", "update warnings"))
			})
		})
	})
})
