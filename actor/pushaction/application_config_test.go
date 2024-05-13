package pushaction_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/manifest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Application Config", func() {
	var (
		actor                   *Actor
		fakeV2Actor             *pushactionfakes.FakeV2Actor
		fakeV3Actor             *pushactionfakes.FakeV3Actor
		fakeSharedActor         *pushactionfakes.FakeSharedActor
		fakeRandomWordGenerator *pushactionfakes.FakeRandomWordGenerator
	)

	BeforeEach(func() {
		actor, fakeV2Actor, fakeV3Actor, fakeSharedActor = getTestPushActor()

		fakeRandomWordGenerator = new(pushactionfakes.FakeRandomWordGenerator)
		actor.WordGenerator = fakeRandomWordGenerator
	})

	Describe("ApplicationConfig", func() {
		Describe("CreatingApplication", func() {
			When("the app did not exist", func() {
				It("returns true", func() {
					config := ApplicationConfig{}
					Expect(config.CreatingApplication()).To(BeTrue())
				})
			})

			When("the app exists", func() {
				It("returns false", func() {
					config := ApplicationConfig{CurrentApplication: Application{Application: v2action.Application{GUID: "some-app-guid"}}}
					Expect(config.CreatingApplication()).To(BeFalse())
				})
			})
		})

		Describe("UpdatedApplication", func() {
			When("the app did not exist", func() {
				It("returns false", func() {
					config := ApplicationConfig{}
					Expect(config.UpdatingApplication()).To(BeFalse())
				})
			})

			When("the app exists", func() {
				It("returns true", func() {
					config := ApplicationConfig{CurrentApplication: Application{Application: v2action.Application{GUID: "some-app-guid"}}}
					Expect(config.UpdatingApplication()).To(BeTrue())
				})
			})
		})
	})

	Describe("ConvertToApplicationConfigs", func() {
		var (
			appName   string
			domain    v2action.Domain
			filesPath string

			orgGUID      string
			spaceGUID    string
			noStart      bool
			manifestApps []manifest.Application

			configs    []ApplicationConfig
			warnings   Warnings
			executeErr error

			firstConfig ApplicationConfig
		)

		BeforeEach(func() {
			appName = "some-app"
			orgGUID = "some-org-guid"
			spaceGUID = "some-space-guid"
			noStart = false

			var err error
			filesPath, err = ioutil.TempDir("", "convert-to-application-configs")
			Expect(err).ToNot(HaveOccurred())

			// The temp directory created on OSX contains a symlink and needs to be evaluated.
			filesPath, err = filepath.EvalSymlinks(filesPath)
			Expect(err).ToNot(HaveOccurred())

			manifestApps = []manifest.Application{{
				Name: appName,
				Path: filesPath,
			}}

			domain = v2action.Domain{
				Name: "private-domain.com",
				GUID: "some-private-domain-guid",
			}

			// Prevents NoDomainsFoundError
			fakeV2Actor.GetOrganizationDomainsReturns(
				[]v2action.Domain{domain},
				v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"},
				nil,
			)
		})

		JustBeforeEach(func() {
			configs, warnings, executeErr = actor.ConvertToApplicationConfigs(orgGUID, spaceGUID, noStart, manifestApps)
			if len(configs) > 0 {
				firstConfig = configs[0]
			}
		})

		AfterEach(func() {
			Expect(os.RemoveAll(filesPath)).ToNot(HaveOccurred())
		})

		When("the path is a symlink", func() {
			var target string

			BeforeEach(func() {
				parentDir := filepath.Dir(filesPath)
				target = filepath.Join(parentDir, fmt.Sprintf("i-r-symlink%d", GinkgoParallelProcess()))
				Expect(os.Symlink(filesPath, target)).To(Succeed())
				manifestApps[0].Path = target
			})

			AfterEach(func() {
				Expect(os.RemoveAll(target)).ToNot(HaveOccurred())
			})

			It("evaluates the symlink into an absolute path", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(firstConfig.Path).To(Equal(filesPath))
			})

			Context("given a path that does not exist", func() {
				BeforeEach(func() {
					manifestApps[0].Path = "/i/will/fight/you/if/this/exists"
				})

				It("returns errors and warnings", func() {
					Expect(os.IsNotExist(executeErr)).To(BeTrue())

					Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
					Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(0))
				})
			})
		})

		When("the application exists", func() {
			var app Application
			var route v2action.Route

			BeforeEach(func() {
				app = Application{
					Application: v2action.Application{
						Name:      appName,
						GUID:      "some-app-guid",
						SpaceGUID: spaceGUID,
					}}

				route = v2action.Route{
					Domain: v2action.Domain{
						Name: "some-domain.com",
						GUID: "some-domain-guid",
					},
					Host:      app.Name,
					GUID:      "route-guid",
					SpaceGUID: spaceGUID,
				}

				fakeV2Actor.GetApplicationByNameAndSpaceReturns(app.Application, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, nil)
			})

			When("there is an existing route and retrieving the route(s) is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{route}, v2action.Warnings{"app-route-warnings"}, nil)
				})

				When("retrieving the application's services is successful", func() {
					var services []v2action.ServiceInstance

					BeforeEach(func() {
						services = []v2action.ServiceInstance{
							{Name: "service-1", GUID: "service-instance-1"},
							{Name: "service-2", GUID: "service-instance-2"},
						}

						fakeV2Actor.GetServiceInstancesByApplicationReturns(services, v2action.Warnings{"service-instance-warning-1", "service-instance-warning-2"}, nil)
					})

					It("return warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings", "service-instance-warning-1", "service-instance-warning-2"))
					})

					It("sets the current application to the existing application", func() {
						Expect(firstConfig.CurrentApplication).To(Equal(app))

						Expect(fakeV2Actor.GetApplicationByNameAndSpaceCallCount()).To(Equal(1))
						passedName, passedSpaceGUID := fakeV2Actor.GetApplicationByNameAndSpaceArgsForCall(0)
						Expect(passedName).To(Equal(app.Name))
						Expect(passedSpaceGUID).To(Equal(spaceGUID))
					})

					It("sets the current routes", func() {
						Expect(firstConfig.CurrentRoutes).To(ConsistOf(route))

						Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(1))
						Expect(fakeV2Actor.GetApplicationRoutesArgsForCall(0)).To(Equal(app.GUID))
					})

					It("sets the bound services", func() {
						Expect(firstConfig.CurrentServices).To(Equal(map[string]v2action.ServiceInstance{
							"service-1": v2action.ServiceInstance{Name: "service-1", GUID: "service-instance-1"},
							"service-2": v2action.ServiceInstance{Name: "service-2", GUID: "service-instance-2"},
						}))

						Expect(fakeV2Actor.GetServiceInstancesByApplicationCallCount()).To(Equal(1))
						Expect(fakeV2Actor.GetServiceInstancesByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
					})
				})

				When("retrieving the application's services errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("dios mio")
						fakeV2Actor.GetServiceInstancesByApplicationReturns(nil, v2action.Warnings{"service-instance-warning-1", "service-instance-warning-2"}, expectedErr)
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings", "service-instance-warning-1", "service-instance-warning-2"))
					})
				})

				When("the --random-route flag is provided", func() {
					BeforeEach(func() {
						manifestApps[0].RandomRoute = true
					})

					It("does not generate a random route", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf(
							"some-app-warning-1",
							"some-app-warning-2",
							"app-route-warnings",
						))
						Expect(firstConfig.DesiredRoutes).To(ConsistOf(route))

						Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(0))
					})
				})
			})

			When("there is not an existing route and no errors are encountered", func() {
				BeforeEach(func() {
					fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"app-route-warnings"}, nil)
				})

				When("the --random-route flag is provided", func() {
					BeforeEach(func() {
						manifestApps[0].RandomRoute = true
						fakeRandomWordGenerator.RandomAdjectiveReturns("striped")
						fakeRandomWordGenerator.RandomNounReturns("apple")
					})

					It("appends a random route to the current route for desired routes", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings", "private-domain-warnings", "shared-domain-warnings"))
						Expect(firstConfig.DesiredRoutes).To(ConsistOf(
							v2action.Route{
								Domain:    domain,
								SpaceGUID: spaceGUID,
								Host:      "some-app-striped-apple",
							},
						))
					})
				})
			})

			When("retrieving the application's routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"app-route-warnings"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "app-route-warnings"))
				})
			})
		})

		When("the application does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, actionerror.ApplicationNotFoundError{})
			})

			It("creates a new application and sets it to the desired application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "private-domain-warnings", "shared-domain-warnings"))
				Expect(firstConfig.CurrentApplication).To(Equal(Application{Application: v2action.Application{}}))
				Expect(firstConfig.DesiredApplication).To(Equal(Application{
					Application: v2action.Application{
						Name:      appName,
						SpaceGUID: spaceGUID,
					}}))
			})

			When("the --random-route flag is provided", func() {
				BeforeEach(func() {
					manifestApps[0].RandomRoute = true
					fakeRandomWordGenerator.RandomAdjectiveReturns("striped")
					fakeRandomWordGenerator.RandomNounReturns("apple")
				})

				It("appends a random route to the current route for desired routes", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "private-domain-warnings", "shared-domain-warnings"))
					Expect(firstConfig.DesiredRoutes).To(ConsistOf(
						v2action.Route{
							Domain:    domain,
							SpaceGUID: spaceGUID,
							Host:      "some-app-striped-apple",
						},
					))
				})

				When("the -d flag is provided", func() {
					When("the domain is an http domain", func() {
						var httpDomain v2action.Domain

						BeforeEach(func() {
							httpDomain = v2action.Domain{
								Name: "some-http-domain.com",
							}

							manifestApps[0].Domain = "some-http-domain.com"
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns([]v2action.Domain{httpDomain}, v2action.Warnings{"domain-warnings-1", "domain-warnings-2"}, nil)
							fakeRandomWordGenerator.RandomAdjectiveReturns("striped")
							fakeRandomWordGenerator.RandomNounReturns("apple")
						})

						It("appends a random HTTP route to the desired routes", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "domain-warnings-1", "domain-warnings-2"))
							Expect(firstConfig.DesiredRoutes).To(ConsistOf(
								v2action.Route{
									Host:      "some-app-striped-apple",
									Domain:    httpDomain,
									SpaceGUID: spaceGUID,
								},
							))
						})
					})

					When("the domain is a tcp domain", func() {
						var tcpDomain v2action.Domain
						BeforeEach(func() {
							tcpDomain = v2action.Domain{
								Name:            "some-tcp-domain.com",
								RouterGroupType: constant.TCPRouterGroup,
							}

							manifestApps[0].Domain = "some-tcp-domain.com"
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns([]v2action.Domain{tcpDomain}, v2action.Warnings{"domain-warnings-1", "domain-warnings-2"}, nil)
						})

						It("appends a random TCP route to the desired routes", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2", "domain-warnings-1", "domain-warnings-2"))
							Expect(firstConfig.DesiredRoutes).To(ConsistOf(
								v2action.Route{
									Domain:    tcpDomain,
									SpaceGUID: spaceGUID,
									Port:      types.NullInt{IsSet: false},
								},
							))
						})
					})
				})
			})
		})

		When("retrieving the application errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("dios mio")
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
			})
		})

		When("overriding application properties", func() {
			var stack v2action.Stack

			When("the manifest contains all the properties", func() {
				BeforeEach(func() {
					manifestApps[0].Buildpack = types.FilteredString{IsSet: true, Value: "some-buildpack"}
					manifestApps[0].Command = types.FilteredString{IsSet: true, Value: "some-command"}
					manifestApps[0].DockerImage = "some-docker-image"
					manifestApps[0].DockerUsername = "some-docker-username"
					manifestApps[0].DockerPassword = "some-docker-password"
					manifestApps[0].HealthCheckTimeout = 5
					manifestApps[0].Instances = types.NullInt{Value: 1, IsSet: true}
					manifestApps[0].DiskQuota = types.NullByteSizeInMb{Value: 2, IsSet: true}
					manifestApps[0].Memory = types.NullByteSizeInMb{Value: 3, IsSet: true}
					manifestApps[0].StackName = "some-stack"
					manifestApps[0].EnvironmentVariables = map[string]string{
						"env1": "1",
						"env3": "3",
					}

					stack = v2action.Stack{
						Name: "some-stack",
						GUID: "some-stack-guid",
					}

					fakeV2Actor.GetStackByNameReturns(stack, v2action.Warnings{"some-stack-warning"}, nil)
				})

				It("overrides the current application properties", func() {
					Expect(warnings).To(ConsistOf("some-stack-warning", "private-domain-warnings", "shared-domain-warnings"))

					Expect(firstConfig.DesiredApplication).To(MatchFields(IgnoreExtras, Fields{
						"Application": MatchFields(IgnoreExtras, Fields{
							"Command":     Equal(types.FilteredString{IsSet: true, Value: "some-command"}),
							"DockerImage": Equal("some-docker-image"),
							"DockerCredentials": MatchFields(IgnoreExtras, Fields{
								"Username": Equal("some-docker-username"),
								"Password": Equal("some-docker-password"),
							}),
							"EnvironmentVariables": Equal(map[string]string{
								"env1": "1",
								"env3": "3",
							}),
							"HealthCheckTimeout": BeEquivalentTo(5),
							"Instances":          Equal(types.NullInt{Value: 1, IsSet: true}),
							"DiskQuota":          Equal(types.NullByteSizeInMb{IsSet: true, Value: 2}),
							"Memory":             Equal(types.NullByteSizeInMb{IsSet: true, Value: 3}),
							"StackGUID":          Equal("some-stack-guid"),
						}),
						"Buildpacks": ConsistOf("some-buildpack"),
						"Stack":      Equal(stack),
					}))

					Expect(fakeV2Actor.GetStackByNameCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GetStackByNameArgsForCall(0)).To(Equal("some-stack"))
				})
			})

			When("the manifest does not contain any properties", func() {
				BeforeEach(func() {
					stack = v2action.Stack{
						Name: "some-stack",
						GUID: "some-stack-guid",
					}
					fakeV2Actor.GetStackReturns(stack, nil, nil)

					app := v2action.Application{
						Buildpack: types.FilteredString{IsSet: true, Value: "some-buildpack"},
						Command:   types.FilteredString{IsSet: true, Value: "some-command"},
						DockerCredentials: ccv2.DockerCredentials{
							Username: "some-docker-username",
							Password: "some-docker-password",
						},
						DockerImage: "some-docker-image",
						DiskQuota:   types.NullByteSizeInMb{IsSet: true, Value: 2},
						EnvironmentVariables: map[string]string{
							"env2": "2",
							"env3": "9",
						},
						GUID:                    "some-app-guid",
						HealthCheckHTTPEndpoint: "/some-endpoint",
						HealthCheckTimeout:      5,
						HealthCheckType:         constant.ApplicationHealthCheckPort,
						Instances:               types.NullInt{Value: 3, IsSet: true},
						Memory:                  types.NullByteSizeInMb{IsSet: true, Value: 3},
						Name:                    appName,
						StackGUID:               stack.GUID,
					}
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(app, nil, nil)
				})

				It("keeps the original app properties", func() {
					Expect(firstConfig.DesiredApplication.Buildpack).To(Equal(types.FilteredString{IsSet: true, Value: "some-buildpack"}))
					Expect(firstConfig.DesiredApplication.Command).To(Equal(types.FilteredString{IsSet: true, Value: "some-command"}))
					Expect(firstConfig.DesiredApplication.DockerImage).To(Equal("some-docker-image"))
					Expect(firstConfig.DesiredApplication.DockerCredentials.Username).To(Equal("some-docker-username"))
					Expect(firstConfig.DesiredApplication.DockerCredentials.Password).To(Equal("some-docker-password"))
					Expect(firstConfig.DesiredApplication.EnvironmentVariables).To(Equal(map[string]string{
						"env2": "2",
						"env3": "9",
					}))
					Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(Equal("/some-endpoint"))
					Expect(firstConfig.DesiredApplication.HealthCheckTimeout).To(BeEquivalentTo(5))
					Expect(firstConfig.DesiredApplication.HealthCheckType).To(Equal(constant.ApplicationHealthCheckPort))
					Expect(firstConfig.DesiredApplication.Instances).To(Equal(types.NullInt{Value: 3, IsSet: true}))
					Expect(firstConfig.DesiredApplication.DiskQuota).To(Equal(types.NullByteSizeInMb{IsSet: true, Value: 2}))
					Expect(firstConfig.DesiredApplication.Memory).To(Equal(types.NullByteSizeInMb{IsSet: true, Value: 3}))
					Expect(firstConfig.DesiredApplication.StackGUID).To(Equal("some-stack-guid"))
					Expect(firstConfig.DesiredApplication.Stack).To(Equal(stack))
				})
			})

			When("setting health check variables", func() {
				When("setting the type to 'http'", func() {
					BeforeEach(func() {
						manifestApps[0].HealthCheckType = "http"
					})

					When("the http health check endpoint is set", func() {
						BeforeEach(func() {
							manifestApps[0].HealthCheckHTTPEndpoint = "/some/endpoint"
						})

						It("should overried the health check type and the endpoint should be set", func() {
							Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(Equal("/some/endpoint"))
							Expect(firstConfig.DesiredApplication.HealthCheckType).To(Equal(constant.ApplicationHealthCheckHTTP))
						})
					})

					When("the http health check endpoint is not set", func() {
						It("should override the health check type and the endpoint should be defaulted to \"/\"", func() {
							Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(Equal("/"))
							Expect(firstConfig.DesiredApplication.HealthCheckType).To(Equal(constant.ApplicationHealthCheckHTTP))
						})
					})
				})

				When("setting type to 'port'", func() {
					BeforeEach(func() {
						manifestApps[0].HealthCheckType = "port"
					})

					It("should override the health check type and the endpoint should not be set", func() {
						Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(BeEmpty())
						Expect(firstConfig.DesiredApplication.HealthCheckType).To(Equal(constant.ApplicationHealthCheckPort))
					})
				})

				When("setting type to 'process'", func() {
					BeforeEach(func() {
						manifestApps[0].HealthCheckType = "process"
					})

					It("should override the health check type and the endpoint should not be set", func() {
						Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(BeEmpty())
						Expect(firstConfig.DesiredApplication.HealthCheckType).To(Equal(constant.ApplicationHealthCheckProcess))
					})
				})

				When("type is unset", func() {
					It("leaves the previously set values", func() {
						Expect(firstConfig.DesiredApplication.HealthCheckHTTPEndpoint).To(BeEmpty())
						Expect(firstConfig.DesiredApplication.HealthCheckType).To(BeEmpty())
					})
				})
			})

			When("retrieving the stack errors", func() {
				var expectedErr error

				BeforeEach(func() {
					manifestApps[0].StackName = "some-stack"

					expectedErr = errors.New("potattototototototootot")
					fakeV2Actor.GetStackByNameReturns(v2action.Stack{}, v2action.Warnings{"some-stack-warning"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-stack-warning"))
				})
			})

			When("both the manifest and application contains environment variables", func() {
				BeforeEach(func() {
					manifestApps[0].EnvironmentVariables = map[string]string{
						"env1": "1",
						"env3": "3",
					}

					app := v2action.Application{
						EnvironmentVariables: map[string]string{
							"env2": "2",
							"env3": "9",
						},
					}
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(app, nil, nil)
				})

				It("adds/overrides the existing environment variables", func() {
					Expect(firstConfig.DesiredApplication.EnvironmentVariables).To(Equal(map[string]string{
						"env1": "1",
						"env2": "2",
						"env3": "3",
					}))

					// Does not modify original set of env vars
					Expect(firstConfig.CurrentApplication.EnvironmentVariables).To(Equal(map[string]string{
						"env2": "2",
						"env3": "9",
					}))
				})
			})

			When("neither the manifest or the application contains environment variables", func() {
				It("leaves the EnvironmentVariables as nil", func() {
					Expect(firstConfig.DesiredApplication.EnvironmentVariables).To(BeNil())
				})
			})

			When("no-start is set to true", func() {
				BeforeEach(func() {
					noStart = true
				})

				It("sets the desired app state to stopped", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(firstConfig.DesiredApplication.Stopped()).To(BeTrue())
				})
			})

			Describe("Buildpacks", func() {
				When("the application is new", func() {
					When("the 'buildpack' field is set", func() {
						When("a buildpack name is provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpack = types.FilteredString{
									IsSet: true,
									Value: "banana",
								}
							})

							It("sets buildpacks to the provided buildpack name", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana"))
							})
						})

						When("buildpack auto detection is set", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpack = types.FilteredString{
									IsSet: true,
								}
							})

							It("sets buildpacks to the provided buildpack name", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(Equal([]string{}))
							})
						})
					})

					When("the 'buildpacks' field is set", func() {
						When("multiple buildpacks are provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{"banana", "strawberry"}
							})

							It("sets buildpacks to the provided buildpack names", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana", "strawberry"))
							})
						})

						When("a single buildpack is provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{"banana"}
							})

							It("sets buildpacks to the provided buildpack names", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana"))
							})
						})

						When("buildpack auto detection is set", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{}
							})

							It("sets buildpacks to an empty slice", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(Equal([]string{}))
							})
						})
					})

					When("nothing is set", func() {
						It("leaves buildpack and buildpacks unset", func() {
							Expect(firstConfig.DesiredApplication.Buildpack).To(Equal(types.FilteredString{}))
							Expect(firstConfig.DesiredApplication.Buildpacks).To(BeNil())
						})
					})
				})

				When("the application exists", func() {
					BeforeEach(func() {
						fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{
							Name:      appName,
							GUID:      "some-app-guid",
							SpaceGUID: spaceGUID,
							Buildpack: types.FilteredString{IsSet: true, Value: "something-I-don't-care"},
						},
							nil, nil)

						fakeV3Actor.GetApplicationByNameAndSpaceReturns(
							v3action.Application{LifecycleBuildpacks: []string{"something-I-don't-care"}},
							nil, nil)
					})

					When("the 'buildpack' field is set", func() {
						When("a buildpack name is provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpack = types.FilteredString{
									IsSet: true,
									Value: "banana",
								}
							})

							It("sets buildpacks to the provided buildpack name", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana"))
							})
						})

						When("buildpack auto detection is set", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpack = types.FilteredString{
									IsSet: true,
								}
							})

							It("sets buildpacks to the provided buildpack name", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(Equal([]string{}))
							})
						})
					})

					When("the 'buildpacks' field is set", func() {
						When("multiple buildpacks are provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{"banana", "strawberry"}
							})

							It("sets buildpacks to the provided buildpack names", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana", "strawberry"))
							})
						})

						When("a single buildpack is provided", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{"banana"}
							})

							It("sets buildpacks to the provided buildpack names", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("banana"))
							})
						})

						When("buildpack auto detection is set", func() {
							BeforeEach(func() {
								manifestApps[0].Buildpacks = []string{}
							})

							It("sets buildpacks to an empty slice", func() {
								Expect(firstConfig.DesiredApplication.Buildpacks).To(Equal([]string{}))
							})
						})
					})

					When("nothing is set", func() {
						It("use the original values", func() {
							Expect(firstConfig.DesiredApplication.Buildpack).To(Equal(types.FilteredString{IsSet: true, Value: "something-I-don't-care"}))
							Expect(firstConfig.DesiredApplication.Buildpacks).To(ConsistOf("something-I-don't-care"))
						})
					})
				})
			})
		})

		When("the manifest contains services", func() {
			BeforeEach(func() {
				manifestApps[0].Services = []string{"service_1", "service_2"}
				fakeV2Actor.GetServiceInstancesByApplicationReturns([]v2action.ServiceInstance{
					{Name: "service_1", SpaceGUID: spaceGUID},
					{Name: "service_3", SpaceGUID: spaceGUID},
				}, v2action.Warnings{"some-service-warning-1"}, nil)
			})

			When("retrieving services is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(v2action.ServiceInstance{Name: "service_2", SpaceGUID: spaceGUID}, v2action.Warnings{"some-service-warning-2"}, nil)
				})

				It("sets DesiredServices", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "some-service-warning-1", "some-service-warning-2"))
					Expect(firstConfig.DesiredServices).To(Equal(map[string]v2action.ServiceInstance{
						"service_1": v2action.ServiceInstance{Name: "service_1", SpaceGUID: spaceGUID},
						"service_2": v2action.ServiceInstance{Name: "service_2", SpaceGUID: spaceGUID},
						"service_3": v2action.ServiceInstance{Name: "service_3", SpaceGUID: spaceGUID},
					}))

					Expect(fakeV2Actor.GetServiceInstanceByNameAndSpaceCallCount()).To(Equal(1))

					inputServiceName, inputSpaceGUID := fakeV2Actor.GetServiceInstanceByNameAndSpaceArgsForCall(0)
					Expect(inputServiceName).To(Equal("service_2"))
					Expect(inputSpaceGUID).To(Equal(spaceGUID))
				})
			})

			When("retrieving services fails", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("po-tat-toe")
					fakeV2Actor.GetServiceInstanceByNameAndSpaceReturns(v2action.ServiceInstance{}, v2action.Warnings{"some-service-warning-2"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("some-service-warning-1", "some-service-warning-2"))
				})
			})
		})

		When("no-route is set", func() {
			BeforeEach(func() {
				manifestApps[0].NoRoute = true

				fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, nil, actionerror.ApplicationNotFoundError{})
			})

			It("should set NoRoute to true", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())
				Expect(firstConfig.NoRoute).To(BeTrue())
				Expect(firstConfig.DesiredRoutes).To(BeEmpty())
			})

			It("should skip route generation", func() {
				Expect(fakeV2Actor.GetDomainsByNameAndOrganizationCallCount()).To(Equal(0))
				Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(0))
			})
		})

		When("routes are defined", func() {
			BeforeEach(func() {
				manifestApps[0].Routes = []string{"route-1.private-domain.com", "route-2.private-domain.com"}
			})

			When("retrieving the routes are successful", func() {
				BeforeEach(func() {
					// Assumes new routes
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, nil, actionerror.ApplicationNotFoundError{})
					fakeV2Actor.GetDomainsByNameAndOrganizationReturns([]v2action.Domain{domain}, v2action.Warnings{"domain-warnings-1", "domains-warnings-2"}, nil)
					fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, actionerror.RouteNotFoundError{})
				})

				When("the --random-route flag is provided", func() {
					BeforeEach(func() {
						manifestApps[0].RandomRoute = true
					})

					It("adds the new routes to the desired routes, and does not generate a random route", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("domain-warnings-1", "domains-warnings-2", "get-route-warnings", "get-route-warnings"))
						Expect(firstConfig.DesiredRoutes).To(ConsistOf(v2action.Route{
							Domain:    domain,
							Host:      "route-1",
							SpaceGUID: spaceGUID,
						}, v2action.Route{
							Domain:    domain,
							Host:      "route-2",
							SpaceGUID: spaceGUID,
						}))

						Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(0))
					})
				})

				It("adds the new routes to the desired routes", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("domain-warnings-1", "domains-warnings-2", "get-route-warnings", "get-route-warnings"))
					Expect(firstConfig.DesiredRoutes).To(ConsistOf(v2action.Route{
						Domain:    domain,
						Host:      "route-1",
						SpaceGUID: spaceGUID,
					}, v2action.Route{
						Domain:    domain,
						Host:      "route-2",
						SpaceGUID: spaceGUID,
					}))
				})
			})

			When("retrieving the routes fails", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					// Assumes new routes
					fakeV2Actor.GetApplicationByNameAndSpaceReturns(v2action.Application{}, nil, actionerror.ApplicationNotFoundError{})
					fakeV2Actor.GetDomainsByNameAndOrganizationReturns([]v2action.Domain{domain}, v2action.Warnings{"domain-warnings-1", "domains-warnings-2"}, expectedErr)
				})

				It("returns errors and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("domain-warnings-1", "domains-warnings-2"))
				})
			})
		})

		When("routes are not defined", func() {
			var app v2action.Application

			BeforeEach(func() {
				app = v2action.Application{
					GUID: "some-app-guid",
					Name: appName,
				}
				fakeV2Actor.GetApplicationByNameAndSpaceReturns(app, nil, nil)
			})

			When("the default route is mapped", func() {
				var existingRoute v2action.Route

				BeforeEach(func() {
					existingRoute = v2action.Route{
						Domain: v2action.Domain{
							Name: "some-domain.com",
							GUID: "some-domain-guid",
						},
						Host:      app.Name,
						GUID:      "route-guid",
						SpaceGUID: spaceGUID,
					}
					fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{existingRoute}, v2action.Warnings{"app-route-warnings"}, nil)
				})

				When("only the -d flag is provided", func() {
					BeforeEach(func() {
						manifestApps[0].Domain = "some-private-domain"
					})

					When("the provided domain exists", func() {
						BeforeEach(func() {
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns(
								[]v2action.Domain{domain},
								v2action.Warnings{"some-organization-domain-warning"},
								nil,
							)
							fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, actionerror.RouteNotFoundError{})
						})

						It("it uses the provided domain instead of the first shared domain", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-organization-domain-warning", "app-route-warnings", "get-route-warnings"))

							Expect(firstConfig.DesiredRoutes).To(ConsistOf(
								existingRoute,
								v2action.Route{
									Domain:    domain,
									Host:      appName,
									SpaceGUID: spaceGUID,
								}),
							)
							Expect(fakeV2Actor.GetDomainsByNameAndOrganizationCallCount()).To(Equal(1))
							domainNamesArg, orgGUIDArg := fakeV2Actor.GetDomainsByNameAndOrganizationArgsForCall(0)
							Expect(domainNamesArg).To(Equal([]string{"some-private-domain"}))
							Expect(orgGUIDArg).To(Equal(orgGUID))
						})
					})

					When("the provided domain does not exist", func() {
						BeforeEach(func() {
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns(
								[]v2action.Domain{},
								v2action.Warnings{"some-organization-domain-warning"},
								nil,
							)
						})

						It("returns an DomainNotFoundError", func() {
							Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "some-private-domain"}))
							Expect(warnings).To(ConsistOf("some-organization-domain-warning", "app-route-warnings"))

							Expect(fakeV2Actor.GetDomainsByNameAndOrganizationCallCount()).To(Equal(1))
							domainNamesArg, orgGUIDArg := fakeV2Actor.GetDomainsByNameAndOrganizationArgsForCall(0)
							Expect(domainNamesArg).To(Equal([]string{"some-private-domain"}))
							Expect(orgGUIDArg).To(Equal(orgGUID))
						})
					})
				})
			})

			When("the default route is not mapped to the app", func() {
				When("only the -d flag is provided", func() {
					BeforeEach(func() {
						manifestApps[0].Domain = "some-private-domain"
						fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"app-route-warnings"}, nil)
					})

					When("the provided domain exists", func() {
						BeforeEach(func() {
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns(
								[]v2action.Domain{domain},
								v2action.Warnings{"some-organization-domain-warning"},
								nil,
							)
							fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, actionerror.RouteNotFoundError{})
						})

						It("it uses the provided domain instead of the first shared domain", func() {
							Expect(executeErr).ToNot(HaveOccurred())
							Expect(warnings).To(ConsistOf("some-organization-domain-warning", "app-route-warnings", "get-route-warnings"))

							Expect(firstConfig.DesiredRoutes).To(ConsistOf(
								v2action.Route{
									Domain:    domain,
									Host:      appName,
									SpaceGUID: spaceGUID,
								}),
							)
							Expect(fakeV2Actor.GetDomainsByNameAndOrganizationCallCount()).To(Equal(1))
							domainNamesArg, orgGUIDArg := fakeV2Actor.GetDomainsByNameAndOrganizationArgsForCall(0)
							Expect(domainNamesArg).To(Equal([]string{"some-private-domain"}))
							Expect(orgGUIDArg).To(Equal(orgGUID))
						})
					})

					When("the provided domain does not exist", func() {
						BeforeEach(func() {
							fakeV2Actor.GetDomainsByNameAndOrganizationReturns(
								[]v2action.Domain{},
								v2action.Warnings{"some-organization-domain-warning"},
								nil,
							)
						})

						It("returns an DomainNotFoundError", func() {
							Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "some-private-domain"}))
							Expect(warnings).To(ConsistOf("some-organization-domain-warning", "app-route-warnings"))

							Expect(fakeV2Actor.GetDomainsByNameAndOrganizationCallCount()).To(Equal(1))
							domainNamesArg, orgGUIDArg := fakeV2Actor.GetDomainsByNameAndOrganizationArgsForCall(0)
							Expect(domainNamesArg).To(Equal([]string{"some-private-domain"}))
							Expect(orgGUIDArg).To(Equal(orgGUID))
						})
					})
				})

				When("there are routes bound to the application", func() {
					var existingRoute v2action.Route

					BeforeEach(func() {
						existingRoute = v2action.Route{
							Domain: v2action.Domain{
								Name: "some-domain.com",
								GUID: "some-domain-guid",
							},
							Host:      app.Name,
							GUID:      "route-guid",
							SpaceGUID: spaceGUID,
						}
						fakeV2Actor.GetApplicationRoutesReturns([]v2action.Route{existingRoute}, v2action.Warnings{"app-route-warnings"}, nil)
					})

					It("does not bind any addition routes", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("app-route-warnings"))
						Expect(firstConfig.DesiredRoutes).To(ConsistOf(existingRoute))
						Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(0))
					})
				})

				When("there are no routes bound to the application", func() {
					BeforeEach(func() {
						fakeV2Actor.GetApplicationRoutesReturns(nil, v2action.Warnings{"app-route-warnings"}, nil)

						// Assumes new route
						fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, actionerror.RouteNotFoundError{})
					})

					It("adds the default route to desired routes", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "app-route-warnings", "get-route-warnings"))
						Expect(firstConfig.DesiredRoutes).To(ConsistOf(
							v2action.Route{
								Domain:    domain,
								Host:      appName,
								SpaceGUID: spaceGUID,
							}))
					})
				})
			})
		})

		When("scanning for files", func() {
			Context("given a directory", func() {
				When("scanning is successful", func() {
					var resources []sharedaction.Resource

					BeforeEach(func() {
						resources = []sharedaction.Resource{
							{Filename: "I am a file!"},
							{Filename: "I am not a file"},
						}
						fakeSharedActor.GatherDirectoryResourcesReturns(resources, nil)
					})

					It("sets the full resource list on the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
						Expect(firstConfig.AllResources).To(Equal(actor.ConvertSharedResourcesToV2Resources(resources)))
						Expect(firstConfig.Path).To(Equal(filesPath))
						Expect(firstConfig.Archive).To(BeFalse())

						Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(1))
						Expect(fakeSharedActor.GatherDirectoryResourcesArgsForCall(0)).To(Equal(filesPath))
					})
				})

				When("scanning errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("dios mio")
						fakeSharedActor.GatherDirectoryResourcesReturns(nil, expectedErr)
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
					})
				})
			})

			Context("given an archive", func() {
				var archive string

				BeforeEach(func() {
					f, err := ioutil.TempFile("", "convert-to-application-configs-archive")
					Expect(err).ToNot(HaveOccurred())
					archive = f.Name()
					Expect(f.Close()).ToNot(HaveOccurred())

					// The temp file created on OSX contains a symlink and needs to be evaluated.
					archive, err = filepath.EvalSymlinks(archive)
					Expect(err).ToNot(HaveOccurred())

					manifestApps[0].Path = archive
				})

				AfterEach(func() {
					Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
				})

				When("scanning is successful", func() {
					var resources []sharedaction.Resource

					BeforeEach(func() {
						resources = []sharedaction.Resource{
							{Filename: "I am a file!"},
							{Filename: "I am not a file"},
						}
						fakeSharedActor.GatherArchiveResourcesReturns(resources, nil)
					})

					It("sets the full resource list on the config", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
						Expect(firstConfig.AllResources).To(Equal(actor.ConvertSharedResourcesToV2Resources(resources)))
						Expect(firstConfig.Path).To(Equal(archive))
						Expect(firstConfig.Archive).To(BeTrue())

						Expect(fakeSharedActor.GatherArchiveResourcesCallCount()).To(Equal(1))
						Expect(fakeSharedActor.GatherArchiveResourcesArgsForCall(0)).To(Equal(archive))
					})
				})

				When("scanning errors", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("dios mio")
						fakeSharedActor.GatherArchiveResourcesReturns(nil, expectedErr)
					})

					It("returns the error and warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
					})
				})
			})
		})

		When("a docker image is configured", func() {
			BeforeEach(func() {
				manifestApps[0].DockerImage = "some-docker-image-path"
			})

			It("sets the docker image on DesiredApplication and does not gather resources", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(firstConfig.DesiredApplication.DockerImage).To(Equal("some-docker-image-path"))

				Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
			})
		})

		When("a droplet is provided", func() {
			BeforeEach(func() {
				manifestApps[0].DropletPath = filesPath
			})

			It("sets the droplet on DesiredApplication and does not gather resources", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(firstConfig.DropletPath).To(Equal(filesPath))

				Expect(fakeSharedActor.GatherDirectoryResourcesCallCount()).To(Equal(0))
			})
		})

		When("buildpacks (plural) are provided", func() {
			BeforeEach(func() {
				manifestApps[0].Buildpacks = []string{
					"some-buildpack-1",
					"some-buildpack-2",
				}
			})

			It("sets the buildpacks on DesiredApplication", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(len(firstConfig.DesiredApplication.Buildpacks)).To(Equal(2))
				Expect(firstConfig.DesiredApplication.Buildpacks[0]).To(Equal("some-buildpack-1"))
				Expect(firstConfig.DesiredApplication.Buildpacks[1]).To(Equal("some-buildpack-2"))
			})

			When("the buildpacks are an empty array", func() {
				BeforeEach(func() {
					manifestApps[0].Buildpacks = []string{}
				})

				It("set the buildpacks on DesiredApplication to empty array", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(firstConfig.DesiredApplication.Buildpacks).To(Equal([]string{}))
				})
			})
		})
	})
})
