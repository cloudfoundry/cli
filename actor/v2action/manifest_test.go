package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Manifest Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateApplicationManifestByNameAndSpace", func() {
		var (
			manifestApp manifest.Application
			warnings    Warnings
			createErr   error
		)

		JustBeforeEach(func() {
			manifestApp, warnings, createErr = actor.CreateApplicationManifestByNameAndSpace("some-app", "some-space-guid")
		})

		When("getting the application summary errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, ccv2.Warnings{"some-app-warning"}, errors.New("some-app-error"))
			})

			It("returns the error and all warnings", func() {
				Expect(createErr).To(MatchError("some-app-error"))
				Expect(warnings).To(ConsistOf("some-app-warning"))
			})
		})

		When("getting the application summary succeeds", func() {
			var app ccv2.Application

			Describe("buildpacks", func() {
				When("buildpack is not set", func() {
					It("does not populate buildpacks field", func() {
						Expect(manifestApp.Buildpacks).To(BeNil())
					})
				})

				When("buildpack is set", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID: "some-app-guid",
							Name: "some-app",
							Buildpack: types.FilteredString{
								IsSet: true,
								Value: "some-buildpack",
							},
						}

						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})

					It("populates buildpacks field", func() {
						Expect(manifestApp.Buildpacks).To(ConsistOf("some-buildpack"))
					})
				})

				When("buildpack and detected buildpack are set", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID: "some-app-guid",
							Name: "some-app",
							Buildpack: types.FilteredString{
								IsSet: true,
								Value: "some-buildpack",
							},
							DetectedBuildpack: types.FilteredString{
								IsSet: true,
								Value: "some-detected-buildpack",
							},
						}

						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})

					It("populates buildpacks field with the buildpack", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp.Buildpacks).To(ConsistOf("some-buildpack"))
					})
				})
			})

			Describe("docker images", func() {
				When("docker image and username are provided", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID:        "some-app-guid",
							Name:        "some-app",
							DockerImage: "some-docker-image",
							DockerCredentials: ccv2.DockerCredentials{
								Username: "some-docker-username",
								Password: "some-docker-password", // CC currently always returns an empty string
							},
						}

						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})

					It("populates docker image and username", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
							"DockerImage":    Equal("some-docker-image"),
							"DockerUsername": Equal("some-docker-username"),
						}))
					})
				})

				When("docker image and username are not provided", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID: "some-app-guid",
							Name: "some-app",
						}

						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})
					It("does not include it in manifest", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
							"DockerImage":    BeEmpty(),
							"DockerUsername": BeEmpty(),
						}))
					})
				})
			})

			Describe("health check", func() {
				When("the health check type is http", func() {
					When("the health check endpoint path is '/'", func() {
						BeforeEach(func() {
							app = ccv2.Application{
								GUID:                    "some-app-guid",
								Name:                    "some-app",
								HealthCheckType:         constant.ApplicationHealthCheckHTTP,
								HealthCheckHTTPEndpoint: "/",
							}
							fakeCloudControllerClient.GetApplicationsReturns(
								[]ccv2.Application{app},
								ccv2.Warnings{"some-app-warning"},
								nil)
						})

						It("does not include health check endpoint in manifest", func() {
							Expect(createErr).NotTo(HaveOccurred())
							Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
								"HealthCheckType":         Equal("http"),
								"HealthCheckHTTPEndpoint": BeEmpty(),
							}))
						})
					})

					When("the health check endpoint path is not the default", func() {
						BeforeEach(func() {
							app = ccv2.Application{
								GUID:                    "some-app-guid",
								Name:                    "some-app",
								HealthCheckType:         constant.ApplicationHealthCheckHTTP,
								HealthCheckHTTPEndpoint: "/whatever",
							}
							fakeCloudControllerClient.GetApplicationsReturns(
								[]ccv2.Application{app},
								ccv2.Warnings{"some-app-warning"},
								nil)
						})

						It("populates the health check endpoint in manifest", func() {
							Expect(createErr).NotTo(HaveOccurred())
							Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
								"HealthCheckType":         Equal("http"),
								"HealthCheckHTTPEndpoint": Equal("/whatever"),
							}))
						})
					})
				})

				When("the health check type is process", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID:                    "some-app-guid",
							Name:                    "some-app",
							HealthCheckType:         constant.ApplicationHealthCheckProcess,
							HealthCheckHTTPEndpoint: "/",
						}
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})

					It("only populates health check type", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
							"HealthCheckType":         Equal("process"),
							"HealthCheckHTTPEndpoint": BeEmpty(),
						}))
					})
				})

				When("the health check type is port", func() {
					BeforeEach(func() {
						app = ccv2.Application{
							GUID:                    "some-app-guid",
							Name:                    "some-app",
							HealthCheckType:         constant.ApplicationHealthCheckPort,
							HealthCheckHTTPEndpoint: "/",
						}
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv2.Application{app},
							ccv2.Warnings{"some-app-warning"},
							nil)
					})

					It("does not include health check type and endpoint", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
							"HealthCheckType":         BeEmpty(),
							"HealthCheckHTTPEndpoint": BeEmpty(),
						}))
					})
				})
			})

			Describe("routes", func() {
				BeforeEach(func() {
					app = ccv2.Application{
						GUID: "some-app-guid",
						Name: "some-app",
					}

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{app},
						ccv2.Warnings{"some-app-warning"},
						nil)
				})

				When("routes are set", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationRoutesReturns(
							[]ccv2.Route{
								{
									GUID:       "some-route-1-guid",
									Host:       "host-1",
									DomainGUID: "some-domain-guid",
								},
								{
									GUID:       "some-route-2-guid",
									Host:       "host-2",
									DomainGUID: "some-domain-guid",
								},
							},
							ccv2.Warnings{"some-routes-warning"},
							nil)

						fakeCloudControllerClient.GetSharedDomainReturns(
							ccv2.Domain{GUID: "some-domain-guid", Name: "some-domain"},
							ccv2.Warnings{"some-domain-warning"},
							nil)
					})
					It("populates the routes", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp.Routes).To(ConsistOf("host-1.some-domain", "host-2.some-domain"))
					})
				})

				When("there are no routes", func() {
					It("returns the app with no-route set to true", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
							"Routes":  BeEmpty(),
							"NoRoute": Equal(true),
						}))
					})
				})
			})

			Describe("services", func() {
				BeforeEach(func() {
					app = ccv2.Application{
						GUID: "some-app-guid",
						Name: "some-app",
					}

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{app},
						ccv2.Warnings{"some-app-warning"},
						nil)
				})

				When("getting services fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceBindingsReturns(
							[]ccv2.ServiceBinding{},
							ccv2.Warnings{"some-service-warning"},
							errors.New("some-service-error"),
						)
					})

					It("returns the error and all warnings", func() {
						Expect(createErr).To(MatchError("some-service-error"))
						Expect(warnings).To(ConsistOf("some-app-warning", "some-service-warning"))
					})
				})

				When("getting services succeeds", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetServiceBindingsReturns(
							[]ccv2.ServiceBinding{
								{ServiceInstanceGUID: "service-1-guid"},
								{ServiceInstanceGUID: "service-2-guid"},
							},
							ccv2.Warnings{"some-service-warning"},
							nil,
						)
						fakeCloudControllerClient.GetServiceInstanceStub = func(serviceInstanceGUID string) (ccv2.ServiceInstance, ccv2.Warnings, error) {
							switch serviceInstanceGUID {
							case "service-1-guid":
								return ccv2.ServiceInstance{Name: "service-1"}, ccv2.Warnings{"some-service-1-warning"}, nil
							case "service-2-guid":
								return ccv2.ServiceInstance{Name: "service-2"}, ccv2.Warnings{"some-service-2-warning"}, nil
							default:
								panic("unknown service instance")
							}
						}
					})

					It("creates the corresponding manifest application", func() {
						Expect(createErr).NotTo(HaveOccurred())
						Expect(manifestApp.Services).To(ConsistOf("service-1", "service-2"))
					})
				})
			})

			Describe("everything else", func() {
				BeforeEach(func() {
					app = ccv2.Application{
						GUID:      "some-app-guid",
						Name:      "some-app",
						DiskQuota: types.NullByteSizeInMb{IsSet: true, Value: 1024},
						Command: types.FilteredString{
							IsSet: true,
							Value: "some-command",
						},
						DetectedStartCommand: types.FilteredString{
							IsSet: true,
							Value: "some-detected-command",
						},
						EnvironmentVariables: map[string]string{
							"env_1": "foo",
							"env_2": "182837403930483038",
							"env_3": "true",
							"env_4": "1.00001",
						},
						HealthCheckTimeout: 120,
						Instances: types.NullInt{
							Value: 10,
							IsSet: true,
						},
						Memory:    types.NullByteSizeInMb{IsSet: true, Value: 200},
						StackGUID: "some-stack-guid",
					}

					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv2.Application{app},
						ccv2.Warnings{"some-app-warning"},
						nil)

					fakeCloudControllerClient.GetStackReturns(
						ccv2.Stack{Name: "some-stack"},
						ccv2.Warnings{"some-stack-warning"},
						nil)
				})

				It("creates the corresponding manifest application", func() {
					Expect(warnings).To(ConsistOf("some-app-warning", "some-stack-warning"))
					Expect(manifestApp).To(MatchFields(IgnoreExtras, Fields{
						"Name":      Equal("some-app"),
						"DiskQuota": Equal(types.NullByteSizeInMb{IsSet: true, Value: 1024}),
						"Command": Equal(types.FilteredString{
							IsSet: true,
							Value: "some-command",
						}),
						"EnvironmentVariables": Equal(map[string]string{
							"env_1": "foo",
							"env_2": "182837403930483038",
							"env_3": "true",
							"env_4": "1.00001",
						}),
						"HealthCheckTimeout": BeEquivalentTo(120),
						"Instances": Equal(types.NullInt{
							Value: 10,
							IsSet: true,
						}),
						"Memory":    Equal(types.NullByteSizeInMb{IsSet: true, Value: 200}),
						"StackName": Equal("some-stack"),
					}))
				})
			})
		})
	})
})
