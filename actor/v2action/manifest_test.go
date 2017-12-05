package v2action_test

import (
	"errors"
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		manifestFilePath          string
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateApplicationManifestByNameAndSpace", func() {
		var (
			createWarnings Warnings
			createErr      error
		)

		JustBeforeEach(func() {
			createWarnings, createErr = actor.CreateApplicationManifestByNameAndSpace("some-app", "some-space-guid", manifestFilePath)
		})

		Context("when getting the application summary errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv2.Application{}, ccv2.Warnings{"some-app-warning"}, errors.New("some-app-error"))
			})

			It("returns the error and all warnings", func() {
				Expect(createErr).To(MatchError("some-app-error"))
				Expect(createWarnings).To(ConsistOf("some-app-warning"))
			})
		})

		Context("when getting the application summary succeeds", func() {
			var app ccv2.Application

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
					DiskQuota:   types.NullByteSizeInMb{IsSet: true, Value: 1024},
					DockerImage: "some-docker-image",
					DockerCredentials: ccv2.DockerCredentials{
						Username: "some-docker-username",
						Password: "some-docker-password", // CC currently always returns an empty string
					},
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
					HealthCheckTimeout:      120,
					HealthCheckHTTPEndpoint: "\\some-endpoint",
					HealthCheckType:         "http",
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

				fakeCloudControllerClient.GetStackReturns(
					ccv2.Stack{Name: "some-stack"},
					ccv2.Warnings{"some-stack-warning"},
					nil)
			})

			Context("when getting services fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetServiceBindingsReturns(
						[]ccv2.ServiceBinding{},
						ccv2.Warnings{"some-service-warning"},
						errors.New("some-service-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(createErr).To(MatchError("some-service-error"))
					Expect(createWarnings).To(ConsistOf("some-app-warning", "some-routes-warning", "some-domain-warning", "some-stack-warning", "some-service-warning"))
				})
			})

			Context("when getting services succeeds", func() {
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

				Context("when writing manifest succeeds", func() {
					BeforeEach(func() {
						manifestFile, err := ioutil.TempFile("", "manifest-test-")
						Expect(err).NotTo(HaveOccurred())
						Expect(manifestFile.Close()).To(Succeed())
						manifestFilePath = manifestFile.Name()
					})

					AfterEach(func() {
						Expect(os.Remove(manifestFilePath)).To(Succeed())
					})

					It("writes the manifest to the specified path", func() {
						manifestBytes, err := ioutil.ReadFile(manifestFilePath)
						Expect(err).NotTo(HaveOccurred())
						Expect(createWarnings).To(ConsistOf("some-app-warning", "some-routes-warning", "some-domain-warning", "some-stack-warning", "some-service-warning", "some-service-1-warning", "some-service-2-warning"))
						Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  routes:
  - route: host-1.some-domain
  - route: host-2.some-domain
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
					})

					Context("when there are no routes", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetApplicationRoutesReturns(nil, nil, nil)
						})

						It("writes the manifest with no-route set to true", func() {
							manifestBytes, err := ioutil.ReadFile(manifestFilePath)
							Expect(err).NotTo(HaveOccurred())
							Expect(createWarnings).To(ConsistOf("some-app-warning", "some-stack-warning", "some-service-warning", "some-service-1-warning", "some-service-2-warning"))
							Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  no-route: true
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
						})
					})

					Context("when docker image and username are not provided", func() {
						BeforeEach(func() {
							app.DockerImage = ""
							app.DockerCredentials = ccv2.DockerCredentials{}
							fakeCloudControllerClient.GetApplicationsReturns(
								[]ccv2.Application{app},
								ccv2.Warnings{"some-app-warning"},
								nil)
						})

						It("does not include it in manifest", func() {
							manifestBytes, err := ioutil.ReadFile(manifestFilePath)
							Expect(err).NotTo(HaveOccurred())
							Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-http-endpoint: \some-endpoint
  health-check-type: http
  instances: 10
  memory: 200M
  routes:
  - route: host-1.some-domain
  - route: host-2.some-domain
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
						})
					})

					Describe("default CC values", func() {
						// We ommitting default CC values from manifest
						// so that it won't get too big

						Context("when the health check type is port", func() {
							BeforeEach(func() {
								app.HealthCheckType = constant.ApplicationHealthCheckPort
								fakeCloudControllerClient.GetApplicationsReturns(
									[]ccv2.Application{app},
									ccv2.Warnings{"some-app-warning"},
									nil)
							})

							It("does not include health check type and endpoint", func() {
								manifestBytes, err := ioutil.ReadFile(manifestFilePath)
								Expect(err).NotTo(HaveOccurred())
								Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  instances: 10
  memory: 200M
  routes:
  - route: host-1.some-domain
  - route: host-2.some-domain
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
							})
						})

						Context("when the health check type is http", func() {
							Context("when the health check endpoint path is '/'", func() {
								BeforeEach(func() {
									app.HealthCheckType = constant.ApplicationHealthCheckHTTP
									app.HealthCheckHTTPEndpoint = "/"
									fakeCloudControllerClient.GetApplicationsReturns(
										[]ccv2.Application{app},
										ccv2.Warnings{"some-app-warning"},
										nil)
								})

								It("does not include health check endpoint in manifest", func() {
									manifestBytes, err := ioutil.ReadFile(manifestFilePath)
									Expect(err).NotTo(HaveOccurred())
									Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-type: http
  instances: 10
  memory: 200M
  routes:
  - route: host-1.some-domain
  - route: host-2.some-domain
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
								})
							})

							Context("when the health check type is process", func() {
								BeforeEach(func() {
									app.HealthCheckType = constant.ApplicationHealthCheckProcess
									fakeCloudControllerClient.GetApplicationsReturns(
										[]ccv2.Application{app},
										ccv2.Warnings{"some-app-warning"},
										nil)
								})

								It("does not include health check endpoint in manifest", func() {
									manifestBytes, err := ioutil.ReadFile(manifestFilePath)
									Expect(err).NotTo(HaveOccurred())
									Expect(string(manifestBytes)).To(Equal(`applications:
- name: some-app
  buildpack: some-buildpack
  command: some-command
  disk_quota: 1G
  docker:
    image: some-docker-image
    username: some-docker-username
  env:
    env_1: foo
    env_2: "182837403930483038"
    env_3: "true"
    env_4: "1.00001"
  health-check-type: process
  instances: 10
  memory: 200M
  routes:
  - route: host-1.some-domain
  - route: host-2.some-domain
  services:
  - service-1
  - service-2
  stack: some-stack
  timeout: 120
`))
								})
							})
						})
					})
				})

				Context("when writing the manifest fails", func() {
					BeforeEach(func() {
						var err error
						manifestFilePath, err = ioutil.TempDir("", "manifest-test-")
						Expect(err).NotTo(HaveOccurred())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(manifestFilePath)).To(Succeed())
					})

					It("returns an ManifestCreationError", func() {
						Expect(createErr).To(HaveOccurred())
						Expect(createErr.Error()).To(ContainSubstring("Error creating manifest file:"))
						Expect(createWarnings).To(ConsistOf("some-app-warning", "some-routes-warning", "some-domain-warning", "some-stack-warning", "some-service-warning", "some-service-1-warning", "some-service-2-warning"))
					})
				})
			})
		})
	})
})
