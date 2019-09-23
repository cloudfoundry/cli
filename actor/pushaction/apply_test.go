package pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

func collectAllEvents(configStream <-chan ApplicationConfig, eventStream <-chan Event, warningsStream <-chan Warnings, errorStream <-chan error) ([]ApplicationConfig, []Event, Warnings, []error) {
	var (
		configStreamClosed, eventStreamClosed, warningsStreamClosed, errorStreamClosed bool

		allConfigs  []ApplicationConfig
		allEvents   []Event
		allWarnings Warnings
		allErrors   []error
	)

	for {
		select {
		case config, ok := <-configStream:
			if !ok {
				configStreamClosed = true
			}

			allConfigs = append(allConfigs, config)
		case event, ok := <-eventStream:
			if !ok {
				eventStreamClosed = true
			}

			allEvents = append(allEvents, event)
		case warning, ok := <-warningsStream:
			if !ok {
				warningsStreamClosed = true
			}

			allWarnings = append(allWarnings, warning...)
		case err, ok := <-errorStream:
			if !ok {
				errorStreamClosed = true
			}

			if err != nil {
				allErrors = append(allErrors, err)
			}
		}

		if configStreamClosed && eventStreamClosed && warningsStreamClosed && errorStreamClosed {
			break
		}
	}

	return allConfigs, allEvents, allWarnings, allErrors
}

var _ = Describe("Apply", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor

		config          ApplicationConfig
		fakeProgressBar *pushactionfakes.FakeProgressBar

		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error
		configStream   <-chan ApplicationConfig

		allConfigs  []ApplicationConfig
		allEvents   []Event
		allWarnings Warnings
		allErrors   []error
	)

	BeforeEach(func() {
		actor, fakeV2Actor, _, fakeSharedActor = getTestPushActor()
		config = ApplicationConfig{
			DesiredApplication: Application{
				Application: v2action.Application{
					Name:      "some-app-name",
					SpaceGUID: "some-space-guid",
				}},
			DesiredRoutes: []v2action.Route{{Host: "banana"}},
		}
		fakeProgressBar = new(pushactionfakes.FakeProgressBar)
	})

	JustBeforeEach(func() {
		configStream, eventStream, warningsStream, errorStream = actor.Apply(config, fakeProgressBar)

		allConfigs, allEvents, allWarnings, allErrors = collectAllEvents(configStream, eventStream, warningsStream, errorStream)
	})

	When("creating/updating the application is successful", func() {
		var createdApp v2action.Application

		BeforeEach(func() {
			fakeV2Actor.CreateApplicationStub = func(application v2action.Application) (v2action.Application, v2action.Warnings, error) {
				createdApp = application
				createdApp.GUID = "some-app-guid"

				return createdApp, v2action.Warnings{"create-application-warnings-1", "create-application-warnings-2"}, nil
			}
		})

		JustBeforeEach(func() {
			Expect(allEvents).To(matchers.ContainElementsInOrder(SettingUpApplication, CreatedApplication))
			Expect(allWarnings).To(matchers.ContainElementsInOrder("create-application-warnings-1", "create-application-warnings-2"))
		})

		When("the route creation is successful", func() {
			var createdRoutes []v2action.Route

			BeforeEach(func() {
				createdRoutes = []v2action.Route{{Host: "banana", GUID: "some-route-guid"}}
				fakeV2Actor.CreateRouteReturns(createdRoutes[0], v2action.Warnings{"create-route-warnings-1", "create-route-warnings-2"}, nil)
			})

			JustBeforeEach(func() {
				Expect(allEvents).To(matchers.ContainElementsInOrder(CreatingAndMappingRoutes))
				Expect(allWarnings).To(matchers.ContainElementsInOrder("create-route-warnings-1", "create-route-warnings-2"))
				Expect(allEvents).To(matchers.ContainElementsInOrder(CreatedRoutes))
			})

			When("mapping the routes is successful", func() {
				var desiredServices map[string]v2action.ServiceInstance

				BeforeEach(func() {
					desiredServices = map[string]v2action.ServiceInstance{
						"service_1": {Name: "service_1", GUID: "service_guid"},
					}
					config.DesiredServices = desiredServices
					fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"map-route-warnings-1", "map-route-warnings-2"}, nil)
				})

				JustBeforeEach(func() {
					Expect(allWarnings).To(matchers.ContainElementsInOrder("map-route-warnings-1", "map-route-warnings-2"))
					Expect(allEvents).To(matchers.ContainElementsInOrder(BoundRoutes))
				})

				When("service binding is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warnings-1", "bind-service-warnings-2"}, nil)
					})

					JustBeforeEach(func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(ConfiguringServices))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("bind-service-warnings-1", "bind-service-warnings-2"))
						Expect(allEvents).To(matchers.ContainElementsInOrder(BoundServices))
					})

					When("resource matching happens", func() {
						BeforeEach(func() {
							config.Path = "some-path"
						})

						JustBeforeEach(func() {
							Expect(allEvents).To(matchers.ContainElementsInOrder(ResourceMatching))
							Expect(allWarnings).To(matchers.ContainElementsInOrder("resource-warnings-1", "resource-warnings-2"))
						})

						When("there is at least one resource that has not been matched", func() {
							BeforeEach(func() {
								fakeV2Actor.ResourceMatchReturns(nil, []v2action.Resource{{}}, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
							})

							When("the archive creation is successful", func() {
								var archivePath string

								BeforeEach(func() {
									tmpfile, err := ioutil.TempFile("", "fake-archive")
									Expect(err).ToNot(HaveOccurred())
									_, err = tmpfile.Write([]byte("123456"))
									Expect(err).ToNot(HaveOccurred())
									Expect(tmpfile.Close()).ToNot(HaveOccurred())

									archivePath = tmpfile.Name()
									fakeSharedActor.ZipDirectoryResourcesReturns(archivePath, nil)
								})

								JustBeforeEach(func() {
									Expect(allEvents).To(matchers.ContainElementsInOrder(CreatingArchive))
								})

								When("the upload is successful", func() {
									BeforeEach(func() {
										fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
									})

									JustBeforeEach(func() {
										Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplicationWithArchive, UploadWithArchiveComplete))
										Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
									})

									It("sends the updated config and a complete event", func() {
										Expect(allConfigs).To(matchers.ContainElementsInOrder(ApplicationConfig{
											CurrentApplication: Application{Application: createdApp},
											CurrentRoutes:      createdRoutes,
											CurrentServices:    desiredServices,
											DesiredApplication: Application{Application: createdApp},
											DesiredRoutes:      createdRoutes,
											DesiredServices:    desiredServices,
											UnmatchedResources: []v2action.Resource{{}},
											Path:               "some-path",
										}))
										Expect(allEvents).To(matchers.ContainElementsInOrder(Complete))

										Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
									})
								})
							})
						})

						When("all the resources have been matched", func() {
							BeforeEach(func() {
								fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
							})

							JustBeforeEach(func() {
								Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplication))
								Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
							})

							When("the upload is successful", func() {
								BeforeEach(func() {
									fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
								})

								It("sends the updated config and a complete event", func() {
									Expect(allConfigs).To(matchers.ContainElementsInOrder(ApplicationConfig{
										CurrentApplication: Application{Application: createdApp},
										CurrentRoutes:      createdRoutes,
										CurrentServices:    desiredServices,
										DesiredApplication: Application{Application: createdApp},
										DesiredRoutes:      createdRoutes,
										DesiredServices:    desiredServices,
										Path:               "some-path",
									}))
									Expect(allEvents).To(matchers.ContainElementsInOrder(Complete))

									Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
									_, _, reader, readerLength := fakeV2Actor.UploadApplicationPackageArgsForCall(0)
									Expect(reader).To(BeNil())
									Expect(readerLength).To(BeNumerically("==", 0))
								})
							})
						})
					})

					When("a droplet is provided", func() {
						var dropletPath string

						BeforeEach(func() {
							tmpfile, err := ioutil.TempFile("", "fake-droplet")
							Expect(err).ToNot(HaveOccurred())
							_, err = tmpfile.Write([]byte("123456"))
							Expect(err).ToNot(HaveOccurred())
							Expect(tmpfile.Close()).ToNot(HaveOccurred())

							dropletPath = tmpfile.Name()
							config.DropletPath = dropletPath
						})

						AfterEach(func() {
							Expect(os.RemoveAll(dropletPath)).ToNot(HaveOccurred())
						})

						When("the upload is successful", func() {
							BeforeEach(func() {
								fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
							})

							It("sends an updated config and a complete event", func() {
								Expect(allEvents).To(matchers.ContainElementsInOrder(UploadDropletComplete))
								Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
								Expect(allConfigs).To(matchers.ContainElementsInOrder(ApplicationConfig{
									CurrentApplication: Application{Application: createdApp},
									CurrentRoutes:      createdRoutes,
									CurrentServices:    desiredServices,
									DesiredApplication: Application{Application: createdApp},
									DesiredRoutes:      createdRoutes,
									DesiredServices:    desiredServices,
									DropletPath:        dropletPath,
								}))

								Expect(fakeV2Actor.UploadDropletCallCount()).To(Equal(1))
								_, droplet, dropletLength := fakeV2Actor.UploadDropletArgsForCall(0)
								Expect(droplet).To(BeNil())
								Expect(dropletLength).To(BeNumerically("==", 6))
								Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(0))
							})
						})
					})
				})
			})
		})
	})

	When("creating/updating errors", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("dios mio")
			fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-application-warnings-1", "create-application-warnings-2"}, expectedErr)
		})

		It("sends warnings and errors, then stops", func() {
			Expect(allEvents).To(matchers.ContainElementsInOrder(SettingUpApplication))
			Expect(allWarnings).To(matchers.ContainElementsInOrder("create-application-warnings-1", "create-application-warnings-2"))
			Expect(allErrors).To(ContainElement(MatchError(expectedErr)))
		})
	})

	Describe("Routes", func() {
		BeforeEach(func() {
			config.DesiredRoutes = v2action.Routes{{GUID: "some-route-guid"}}
		})

		When("NoRoutes is set", func() {
			BeforeEach(func() {
				config.NoRoute = true
			})

			When("config has at least one route", func() {
				BeforeEach(func() {
					config.CurrentRoutes = []v2action.Route{{GUID: "some-route-guid-1"}}
				})

				When("unmapping routes succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.UnmapRouteFromApplicationReturns(v2action.Warnings{"unmapping-route-warnings"}, nil)
					})

					It("sends the UnmappingRoutes event and does not raise an error", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(UnmappingRoutes, Complete))
						Expect(allWarnings).To(ContainElement("unmapping-route-warnings"))
					})
				})

				When("unmapping routes fails", func() {
					BeforeEach(func() {
						fakeV2Actor.UnmapRouteFromApplicationReturns(v2action.Warnings{"unmapping-route-warnings"}, errors.New("ohno"))
					})

					It("sends the UnmappingRoutes event and raise an error", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(UnmappingRoutes))
						Expect(allEvents).NotTo(ContainElement(Complete))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("unmapping-route-warnings"))
						Expect(allErrors).To(ContainElement(MatchError("ohno")))
					})
				})
			})

			When("config has no routes", func() {
				BeforeEach(func() {
					config.CurrentRoutes = nil
				})

				It("should not send the UnmappingRoutes event", func() {
					Expect(allEvents).NotTo(ContainElement(UnmappingRoutes))
					Expect(allErrors).To(BeEmpty())

					Expect(fakeV2Actor.UnmapRouteFromApplicationCallCount()).To(Equal(0))
				})
			})
		})

		When("NoRoutes is NOT set", func() {
			BeforeEach(func() {
				config.NoRoute = false
			})

			It("should send the CreatingAndMappingRoutes event", func() {
				Expect(allEvents).To(ContainElement(CreatingAndMappingRoutes))
			})

			When("no new routes are provided", func() {
				BeforeEach(func() {
					config.DesiredRoutes = nil
				})

				It("should not send the CreatedRoutes event", func() {
					Expect(allEvents).To(ContainElement(CreatingAndMappingRoutes))
					Expect(allEvents).NotTo(ContainElement(CreatedRoutes))
					Expect(allWarnings).To(BeEmpty())
				})
			})

			When("new routes are provided", func() {
				BeforeEach(func() {
					config.DesiredRoutes = []v2action.Route{{}}
				})

				When("route creation fails", func() {
					BeforeEach(func() {
						fakeV2Actor.CreateRouteReturns(v2action.Route{}, v2action.Warnings{"create-route-warning"}, errors.New("ohno"))
					})

					It("raise an error", func() {
						Expect(allEvents).To(ContainElement(CreatingAndMappingRoutes))
						Expect(allEvents).NotTo(ContainElement(CreatedRoutes))
						Expect(allEvents).NotTo(ContainElement(Complete))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("create-route-warning"))
						Expect(allErrors).To(ContainElement(MatchError("ohno")))
					})
				})

				When("route creation succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.CreateRouteReturns(v2action.Route{}, v2action.Warnings{"create-route-warning"}, nil)
					})

					It("should send the CreatedRoutes event", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(CreatingAndMappingRoutes, CreatedRoutes))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("create-route-warning"))
					})
				})
			})

			When("there are no routes to map", func() {
				BeforeEach(func() {
					config.CurrentRoutes = config.DesiredRoutes
				})

				It("should not send the BoundRoutes event", func() {
					Expect(allEvents).To(ContainElement(CreatingAndMappingRoutes))

					Expect(allEvents).NotTo(ContainElement(BoundRoutes))
				})
			})

			When("there are routes to map", func() {
				BeforeEach(func() {
					config.DesiredRoutes = []v2action.Route{{GUID: "new-guid"}}
				})

				When("binding the route fails", func() {
					BeforeEach(func() {
						fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, errors.New("ohno"))
					})

					It("raise an error", func() {
						Expect(allEvents).To(ContainElement(CreatingAndMappingRoutes))
						Expect(allEvents).NotTo(ContainElement(BoundRoutes))
						Expect(allEvents).NotTo(ContainElement(Complete))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("bind-route-warning"))
						Expect(allErrors).To(ContainElement(MatchError("ohno")))
					})
				})

				When("binding the route succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, nil)
					})

					It("should send the BoundRoutes event", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(CreatingAndMappingRoutes, BoundRoutes))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("bind-route-warning"))
					})
				})
			})
		})
	})

	Describe("Services", func() {
		var (
			service1 v2action.ServiceInstance
			service2 v2action.ServiceInstance
		)

		BeforeEach(func() {
			service1 = v2action.ServiceInstance{Name: "service_1", GUID: "service_1_guid"}
			service2 = v2action.ServiceInstance{Name: "service_2", GUID: "service_2_guid"}
		})

		When("there are no new services", func() {
			BeforeEach(func() {
				config.CurrentServices = map[string]v2action.ServiceInstance{"service1": service1}
				config.DesiredServices = map[string]v2action.ServiceInstance{"service1": service1}
			})

			It("should not send the ConfiguringServices or BoundServices event", func() {
				Expect(allEvents).NotTo(ContainElement(ConfiguringServices))
				Expect(allEvents).NotTo(ContainElement(BoundServices))
			})
		})

		When("are new services", func() {
			BeforeEach(func() {
				config.CurrentServices = map[string]v2action.ServiceInstance{"service1": service1}
				config.DesiredServices = map[string]v2action.ServiceInstance{"service1": service1, "service2": service2}
			})

			When("binding services fails", func() {
				BeforeEach(func() {
					fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warning"}, errors.New("ohno"))
				})

				It("raises an error", func() {
					Expect(allEvents).To(matchers.ContainElementsInOrder(BoundRoutes, ConfiguringServices))
					Expect(allEvents).NotTo(ContainElement(BoundServices))
					Expect(allEvents).NotTo(ContainElement(Complete))
					Expect(allWarnings).To(matchers.ContainElementsInOrder("bind-service-warning"))
					Expect(allErrors).To(ContainElement(MatchError("ohno")))
				})
			})

			When("binding services suceeds", func() {
				BeforeEach(func() {
					fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warning"}, nil)
				})

				It("sends the ConfiguringServices and BoundServices events", func() {
					Expect(allEvents).To(matchers.ContainElementsInOrder(ConfiguringServices, BoundServices))
					Expect(allWarnings).To(matchers.ContainElementsInOrder("bind-service-warning"))
				})
			})
		})
	})

	Describe("Upload", func() {
		When("a droplet is provided", func() {
			var dropletPath string

			BeforeEach(func() {
				tmpfile, err := ioutil.TempFile("", "fake-droplet")
				Expect(err).ToNot(HaveOccurred())
				_, err = tmpfile.Write([]byte("123456"))
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpfile.Close()).ToNot(HaveOccurred())

				dropletPath = tmpfile.Name()
				config.DropletPath = dropletPath
			})

			AfterEach(func() {
				Expect(os.RemoveAll(dropletPath)).ToNot(HaveOccurred())
			})

			When("uploading the droplet fails", func() {
				When("the error is a retryable error", func() {
					var someErr error
					BeforeEach(func() {
						someErr = errors.New("I AM A BANANA")
						fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, ccerror.PipeSeekError{Err: someErr})
					})

					It("should send a RetryUpload event and retry uploading up to 3x", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(
							UploadingDroplet,
							RetryUpload,
							UploadingDroplet,
							RetryUpload,
							UploadingDroplet,
							RetryUpload,
						))

						Expect(allEvents).To(matchers.ContainElementTimes(RetryUpload, 3))
						Expect(allEvents).NotTo(ContainElement(UploadDropletComplete))
						Expect(allEvents).NotTo(ContainElement(Complete))

						Expect(allWarnings).To(matchers.ContainElementsInOrder(
							"droplet-upload-warning",
							"droplet-upload-warning",
							"droplet-upload-warning",
						))

						Expect(fakeV2Actor.UploadDropletCallCount()).To(Equal(3))
						Expect(allErrors).To(ContainElement(MatchError(actionerror.UploadFailedError{Err: someErr})))
					})
				})

				When("the error is not a retryable error", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, errors.New("ohno"))
					})

					It("raises an error", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingDroplet))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("droplet-upload-warning"))
						Expect(allErrors).To(ContainElement(MatchError("ohno")))

						Expect(allEvents).NotTo(ContainElement(EqualEither(RetryUpload, UploadDropletComplete, Complete)))
					})
				})
			})

			When("uploading the droplet is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, nil)
				})

				It("sends the UploadingDroplet event", func() {
					Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingDroplet, UploadDropletComplete))
					Expect(allWarnings).To(matchers.ContainElementsInOrder("droplet-upload-warning"))
				})
			})
		})

		When("app bits are provided", func() {
			When("there is at least one unmatched resource", func() {
				BeforeEach(func() {
					fakeV2Actor.ResourceMatchReturns(nil, []v2action.Resource{{}}, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
				})

				It("returns resource match warnings", func() {
					Expect(allEvents).To(matchers.ContainElementsInOrder(ResourceMatching))
					Expect(allWarnings).To(matchers.ContainElementsInOrder("resource-warnings-1", "resource-warnings-2"))
				})

				When("creating the archive is successful", func() {
					var archivePath string

					BeforeEach(func() {
						tmpfile, err := ioutil.TempFile("", "fake-archive")
						Expect(err).ToNot(HaveOccurred())
						_, err = tmpfile.Write([]byte("123456"))
						Expect(err).ToNot(HaveOccurred())
						Expect(tmpfile.Close()).ToNot(HaveOccurred())

						archivePath = tmpfile.Name()
						fakeSharedActor.ZipDirectoryResourcesReturns(archivePath, nil)
					})

					It("sends a CreatingArchive event", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(CreatingArchive))
					})

					When("the upload is successful", func() {
						BeforeEach(func() {
							fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
						})

						It("sends a UploadingApplicationWithArchive event", func() {
							Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplicationWithArchive))
							Expect(allEvents).To(matchers.ContainElementsInOrder(UploadWithArchiveComplete))
							Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
						})
					})

					When("the upload fails", func() {
						When("the upload error is a retryable error", func() {
							var someErr error

							BeforeEach(func() {
								someErr = errors.New("I AM A BANANA")
								fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, ccerror.PipeSeekError{Err: someErr})
							})

							It("should send a RetryUpload event and retry uploading", func() {
								Expect(allEvents).To(matchers.ContainElementsInOrder(
									UploadingApplicationWithArchive,
									RetryUpload,
									UploadingApplicationWithArchive,
									RetryUpload,
									UploadingApplicationWithArchive,
									RetryUpload,
								))

								Expect(allEvents).To(matchers.ContainElementTimes(RetryUpload, 3))
								Expect(allEvents).NotTo(ContainElement(EqualEither(UploadWithArchiveComplete, Complete)))

								Expect(allWarnings).To(matchers.ContainElementsInOrder(
									"upload-warnings-1", "upload-warnings-2",
									"upload-warnings-1", "upload-warnings-2",
									"upload-warnings-1", "upload-warnings-2",
								))

								Expect(fakeV2Actor.UploadApplicationPackageCallCount()).Should(Equal(3))
								Expect(allErrors).To(ContainElement(MatchError(actionerror.UploadFailedError{Err: someErr})))
							})

						})

						When("the upload error is not a retryable error", func() {
							BeforeEach(func() {
								fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("dios mio"))
							})

							It("sends warnings and errors, then stops", func() {

								Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplicationWithArchive))
								Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
								Expect(allEvents).NotTo(ContainElement(EqualEither(RetryUpload, UploadWithArchiveComplete, Complete)))
								Expect(allErrors).To(ContainElement(MatchError("dios mio")))
							})
						})
					})
				})

				When("creating the archive fails", func() {
					BeforeEach(func() {
						fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("some-error"))
					})

					It("raises an error", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(ResourceMatching))
						Expect(allEvents).NotTo(ContainElement(Complete))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("resource-warnings-1", "resource-warnings-2"))
						Expect(allErrors).To(ContainElement(MatchError("some-error")))

					})
				})
			})

			When("all resources have been matched", func() {
				BeforeEach(func() {
					fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
				})

				It("sends the UploadingApplication event", func() {
					Expect(allEvents).To(matchers.ContainElementsInOrder(ResourceMatching))
					Expect(allWarnings).To(matchers.ContainElementsInOrder("resource-warnings-1", "resource-warnings-2"))
					Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplication))
				})

				When("the upload is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
					})

					It("uploads the application and completes", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplication))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
						Expect(allEvents).To(matchers.ContainElementsInOrder(Complete))
					})
				})

				When("the upload fails", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("some-upload-error"))
					})

					It("returns an error", func() {
						Expect(allEvents).To(matchers.ContainElementsInOrder(UploadingApplication))
						Expect(allWarnings).To(matchers.ContainElementsInOrder("upload-warnings-1", "upload-warnings-2"))
						Expect(allErrors).To(ContainElement(MatchError("some-upload-error")))
						Expect(allEvents).NotTo(ContainElement(Complete))
					})
				})
			})
		})

		When("a docker image is provided", func() {
			BeforeEach(func() {
				config.DesiredApplication.DockerImage = "hi-im-a-ge"

				fakeV2Actor.CreateApplicationReturns(config.DesiredApplication.Application, nil, nil)
			})

			It("should skip uploading anything", func() {
				Expect(allEvents).NotTo(ContainElement(EqualEither(UploadingDroplet, UploadingApplication)))
			})
		})
	})
})
