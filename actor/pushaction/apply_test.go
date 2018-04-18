package pushaction_test

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func streamsDrainedAndClosed(configStream <-chan ApplicationConfig, eventStream <-chan Event, warningsStream <-chan Warnings, errorStream <-chan error) bool {
	var configStreamClosed, eventStreamClosed, warningsStreamClosed, errorStreamClosed bool
	for {
		select {
		case _, ok := <-configStream:
			if !ok {
				configStreamClosed = true
			}
		case _, ok := <-eventStream:
			if !ok {
				eventStreamClosed = true
			}
		case _, ok := <-warningsStream:
			if !ok {
				warningsStreamClosed = true
			}
		case _, ok := <-errorStream:
			if !ok {
				errorStreamClosed = true
			}
		}
		if configStreamClosed && eventStreamClosed && warningsStreamClosed && errorStreamClosed {
			break
		}
	}
	return true
}

// TODO: for refactor: We can use the following style of code to validate that
// each event is received in a specific order

// Expect(nextEvent()).Should(Equal(SettingUpApplication))
// Expect(nextEvent()).Should(Equal(CreatingApplication))
// Expect(nextEvent()).Should(Equal(...))
// Expect(nextEvent()).Should(Equal(...))
// Expect(nextEvent()).Should(Equal(...))
func setUpNextEvent(c <-chan ApplicationConfig, e <-chan Event, w <-chan Warnings) func() Event {
	timeOut := time.Tick(500 * time.Millisecond)

	return func() Event {
		for {
			select {
			case <-c:
			case event, ok := <-e:
				if ok {
					return event
				}
				return ""
			case <-w:
			case <-timeOut:
				return ""
			}
		}
	}
}

func EqualEither(events ...Event) GomegaMatcher {
	var equals []GomegaMatcher
	for _, event := range events {
		equals = append(equals, Equal(event))
	}

	return Or(equals...)
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

		nextEvent func() Event
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		actor = NewActor(fakeV2Actor, fakeSharedActor)
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

		nextEvent = setUpNextEvent(configStream, eventStream, warningsStream)
	})

	AfterEach(func() {
		Eventually(streamsDrainedAndClosed(configStream, eventStream, warningsStream, errorStream)).Should(BeTrue())
	})

	Context("when creating/updating the application is successful", func() {
		var createdApp v2action.Application

		BeforeEach(func() {
			fakeV2Actor.CreateApplicationStub = func(application v2action.Application) (v2action.Application, v2action.Warnings, error) {
				createdApp = application
				createdApp.GUID = "some-app-guid"

				return createdApp, v2action.Warnings{"create-application-warnings-1", "create-application-warnings-2"}, nil
			}
		})

		JustBeforeEach(func() {
			Eventually(eventStream).Should(Receive(Equal(SettingUpApplication)))
			Eventually(warningsStream).Should(Receive(ConsistOf("create-application-warnings-1", "create-application-warnings-2")))
			Eventually(eventStream).Should(Receive(Equal(CreatedApplication)))
		})

		Context("when the route creation is successful", func() {
			var createdRoutes []v2action.Route

			BeforeEach(func() {
				createdRoutes = []v2action.Route{{Host: "banana", GUID: "some-route-guid"}}
				fakeV2Actor.CreateRouteReturns(createdRoutes[0], v2action.Warnings{"create-route-warnings-1", "create-route-warnings-2"}, nil)
			})

			JustBeforeEach(func() {
				Eventually(eventStream).Should(Receive(Equal(CreatingAndMappingRoutes)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warnings-1", "create-route-warnings-2")))
				Eventually(eventStream).Should(Receive(Equal(CreatedRoutes)))
			})

			Context("when mapping the routes is successful", func() {
				var desiredServices map[string]v2action.ServiceInstance

				BeforeEach(func() {
					desiredServices = map[string]v2action.ServiceInstance{
						"service_1": {Name: "service_1", GUID: "service_guid"},
					}
					config.DesiredServices = desiredServices
					fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"map-route-warnings-1", "map-route-warnings-2"}, nil)
				})

				JustBeforeEach(func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("map-route-warnings-1", "map-route-warnings-2")))
					Eventually(eventStream).Should(Receive(Equal(BoundRoutes)))
				})

				Context("when service binding is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warnings-1", "bind-service-warnings-2"}, nil)
					})

					JustBeforeEach(func() {
						Eventually(eventStream).Should(Receive(Equal(ConfiguringServices)))
						Eventually(warningsStream).Should(Receive(ConsistOf("bind-service-warnings-1", "bind-service-warnings-2")))
						Eventually(eventStream).Should(Receive(Equal(BoundServices)))
					})

					Context("when resource matching happens", func() {
						BeforeEach(func() {
							config.Path = "some-path"
						})

						JustBeforeEach(func() {
							Eventually(eventStream).Should(Receive(Equal(ResourceMatching)))
							Eventually(warningsStream).Should(Receive(ConsistOf("resource-warnings-1", "resource-warnings-2")))
						})

						Context("when there is at least one resource that has not been matched", func() {
							BeforeEach(func() {
								fakeV2Actor.ResourceMatchReturns(nil, []v2action.Resource{{}}, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
							})

							Context("when the archive creation is successful", func() {
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
									Eventually(eventStream).Should(Receive(Equal(CreatingArchive)))
								})

								Context("when the upload is successful", func() {
									BeforeEach(func() {
										fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
									})

									JustBeforeEach(func() {
										Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
										Eventually(eventStream).Should(Receive(Equal(UploadWithArchiveComplete)))
										Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
									})

									It("sends the updated config and a complete event", func() {
										Eventually(configStream).Should(Receive(Equal(ApplicationConfig{
											CurrentApplication: Application{Application: createdApp},
											CurrentRoutes:      createdRoutes,
											CurrentServices:    desiredServices,
											DesiredApplication: Application{Application: createdApp},
											DesiredRoutes:      createdRoutes,
											DesiredServices:    desiredServices,
											UnmatchedResources: []v2action.Resource{{}},
											Path:               "some-path",
										})))
										Eventually(eventStream).Should(Receive(Equal(Complete)))

										Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
									})
								})
							})
						})

						Context("when all the resources have been matched", func() {
							BeforeEach(func() {
								fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
							})

							JustBeforeEach(func() {
								Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
							})

							Context("when the upload is successful", func() {
								BeforeEach(func() {
									fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
								})

								It("sends the updated config and a complete event", func() {
									Eventually(configStream).Should(Receive(Equal(ApplicationConfig{
										CurrentApplication: Application{Application: createdApp},
										CurrentRoutes:      createdRoutes,
										CurrentServices:    desiredServices,
										DesiredApplication: Application{Application: createdApp},
										DesiredRoutes:      createdRoutes,
										DesiredServices:    desiredServices,
										Path:               "some-path",
									})))
									Eventually(eventStream).Should(Receive(Equal(Complete)))

									Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
									_, _, reader, readerLength := fakeV2Actor.UploadApplicationPackageArgsForCall(0)
									Expect(reader).To(BeNil())
									Expect(readerLength).To(BeNumerically("==", 0))
								})
							})
						})
					})

					Context("when a droplet is provided", func() {
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

						Context("when the upload is successful", func() {
							BeforeEach(func() {
								fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
							})

							It("sends an updated config and a complete event", func() {
								Eventually(eventStream).Should(Receive(Equal(UploadDropletComplete)))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
								Eventually(configStream).Should(Receive(Equal(ApplicationConfig{
									CurrentApplication: Application{Application: createdApp},
									CurrentRoutes:      createdRoutes,
									CurrentServices:    desiredServices,
									DesiredApplication: Application{Application: createdApp},
									DesiredRoutes:      createdRoutes,
									DesiredServices:    desiredServices,
									DropletPath:        dropletPath,
								})))

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

	Context("when creating/updating errors", func() {
		var expectedErr error

		BeforeEach(func() {
			expectedErr = errors.New("dios mio")
			fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-application-warnings-1", "create-application-warnings-2"}, expectedErr)
		})

		It("sends warnings and errors, then stops", func() {
			Eventually(eventStream).Should(Receive(Equal(SettingUpApplication)))
			Eventually(warningsStream).Should(Receive(ConsistOf("create-application-warnings-1", "create-application-warnings-2")))
			Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			Consistently(eventStream).ShouldNot(Receive())
		})
	})

	Describe("Routes", func() {
		BeforeEach(func() {
			config.DesiredRoutes = v2action.Routes{{GUID: "some-route-guid"}}
		})

		Context("when NoRoutes is set", func() {
			BeforeEach(func() {
				config.NoRoute = true
			})

			Context("when config has at least one route", func() {
				BeforeEach(func() {
					config.CurrentRoutes = []v2action.Route{{GUID: "some-route-guid-1"}}
				})

				Context("when unmapping routes succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.UnmapRouteFromApplicationReturns(v2action.Warnings{"unmapping-route-warnings"}, nil)
					})

					It("sends the UnmappingRoutes event and does not raise an error", func() {
						Eventually(nextEvent).Should(Equal(UnmappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("unmapping-route-warnings")))
						Eventually(nextEvent).Should(Equal(Complete))
					})
				})

				Context("when unmapping routes fails", func() {
					BeforeEach(func() {
						fakeV2Actor.UnmapRouteFromApplicationReturns(v2action.Warnings{"unmapping-route-warnings"}, errors.New("ohno"))
					})

					It("sends the UnmappingRoutes event and raise an error", func() {
						Eventually(nextEvent).Should(Equal(UnmappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("unmapping-route-warnings")))
						Eventually(errorStream).Should(Receive(MatchError("ohno")))
						Consistently(nextEvent).ShouldNot(Equal(Complete))
					})
				})
			})

			Context("when config has no routes", func() {
				BeforeEach(func() {
					config.CurrentRoutes = nil
				})

				It("should not send the UnmappingRoutes event", func() {
					Consistently(nextEvent).ShouldNot(Equal(UnmappingRoutes))
					Consistently(errorStream).ShouldNot(Receive())

					Expect(fakeV2Actor.UnmapRouteFromApplicationCallCount()).To(Equal(0))
				})
			})
		})

		Context("when NoRoutes is NOT set", func() {
			BeforeEach(func() {
				config.NoRoute = false
			})

			It("should send the CreatingAndMappingRoutes event", func() {
				Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
			})

			Context("when no new routes are provided", func() {
				BeforeEach(func() {
					config.DesiredRoutes = nil
				})

				It("should not send the CreatedRoutes event", func() {
					Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
					Eventually(warningsStream).Should(Receive(BeEmpty()))
					Consistently(nextEvent).ShouldNot(Equal(CreatedRoutes))
				})
			})

			Context("when new routes are provided", func() {
				BeforeEach(func() {
					config.DesiredRoutes = []v2action.Route{{}}
				})

				Context("when route creation fails", func() {
					BeforeEach(func() {
						fakeV2Actor.CreateRouteReturns(v2action.Route{}, v2action.Warnings{"create-route-warning"}, errors.New("ohno"))
					})

					It("raise an error", func() {
						Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warning")))
						Eventually(errorStream).Should(Receive(MatchError("ohno")))
						Consistently(nextEvent).ShouldNot(EqualEither(CreatedRoutes, Complete))
					})
				})

				Context("when route creation succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.CreateRouteReturns(v2action.Route{}, v2action.Warnings{"create-route-warning"}, nil)
					})

					It("should send the CreatedRoutes event", func() {
						Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warning")))
						Expect(nextEvent()).To(Equal(CreatedRoutes))
					})
				})
			})

			Context("when there are no routes to map", func() {
				BeforeEach(func() {
					config.CurrentRoutes = config.DesiredRoutes
				})

				It("should not send the BoundRoutes event", func() {
					Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))

					// First warning picks up CreatedRoute warnings, second one picks up
					// MapRoute warnings. No easy way to improve this today
					Eventually(warningsStream).Should(Receive())
					Eventually(warningsStream).Should(Receive())
					Consistently(nextEvent).ShouldNot(Equal(BoundRoutes))
				})
			})

			Context("when there are routes to map", func() {
				BeforeEach(func() {
					config.DesiredRoutes = []v2action.Route{{GUID: "new-guid"}}
				})

				Context("when binding the route fails", func() {
					BeforeEach(func() {
						fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, errors.New("ohno"))
					})

					It("raise an error", func() {
						Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))
						Eventually(errorStream).Should(Receive(MatchError("ohno")))
						Consistently(nextEvent).ShouldNot(EqualEither(BoundRoutes, Complete))
					})
				})

				Context("when binding the route succeeds", func() {
					BeforeEach(func() {
						fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, nil)
					})

					It("should send the BoundRoutes event", func() {
						Eventually(nextEvent).Should(Equal(CreatingAndMappingRoutes))
						Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warning")))
						Expect(nextEvent()).To(Equal(BoundRoutes))
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

		Context("when there are no new services", func() {
			BeforeEach(func() {
				config.CurrentServices = map[string]v2action.ServiceInstance{"service1": service1}
				config.DesiredServices = map[string]v2action.ServiceInstance{"service1": service1}
			})

			It("should not send the ConfiguringServices or BoundServices event", func() {
				Consistently(nextEvent).ShouldNot(EqualEither(ConfiguringServices, BoundServices))
			})
		})

		Context("when are new services", func() {
			BeforeEach(func() {
				config.CurrentServices = map[string]v2action.ServiceInstance{"service1": service1}
				config.DesiredServices = map[string]v2action.ServiceInstance{"service1": service1, "service2": service2}
			})

			Context("when binding services fails", func() {
				BeforeEach(func() {
					fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warning"}, errors.New("ohno"))
				})

				It("raises an error", func() {
					Eventually(nextEvent).Should(Equal(ConfiguringServices))
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-service-warning")))
					Eventually(errorStream).Should(Receive(MatchError("ohno")))
					Consistently(nextEvent).ShouldNot(EqualEither(BoundServices, Complete))
				})
			})

			Context("when binding services suceeds", func() {
				BeforeEach(func() {
					fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warning"}, nil)
				})

				It("sends the ConfiguringServices and BoundServices events", func() {
					Eventually(nextEvent).Should(Equal(ConfiguringServices))
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-service-warning")))
					Expect(nextEvent()).To(Equal(BoundServices))
				})
			})
		})
	})

	Describe("Upload", func() {
		Context("when a droplet is provided", func() {
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

			Context("when uploading the droplet fails", func() {
				Context("when the error is a retryable error", func() {
					var someErr error
					BeforeEach(func() {
						someErr = errors.New("I AM A BANANA")
						fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, ccerror.PipeSeekError{Err: someErr})
					})

					It("should send a RetryUpload event and retry uploading up to 3x", func() {
						Eventually(nextEvent).Should(Equal(UploadingDroplet))
						Eventually(warningsStream).Should(Receive(ConsistOf("droplet-upload-warning")))
						Expect(nextEvent()).To(Equal(RetryUpload))

						Expect(nextEvent()).To(Equal(UploadingDroplet))
						Eventually(warningsStream).Should(Receive(ConsistOf("droplet-upload-warning")))
						Expect(nextEvent()).To(Equal(RetryUpload))

						Expect(nextEvent()).To(Equal(UploadingDroplet))
						Eventually(warningsStream).Should(Receive(ConsistOf("droplet-upload-warning")))
						Expect(nextEvent()).To(Equal(RetryUpload))

						Consistently(nextEvent).ShouldNot(EqualEither(RetryUpload, UploadDropletComplete, Complete))
						Eventually(fakeV2Actor.UploadDropletCallCount()).Should(Equal(3))
						Expect(errorStream).To(Receive(MatchError(actionerror.UploadFailedError{Err: someErr})))
					})
				})

				Context("when the error is not a retryable error", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, errors.New("ohnos"))
					})

					It("raises an error", func() {
						Eventually(nextEvent).Should(Equal(UploadingDroplet))
						Eventually(warningsStream).Should(Receive(ConsistOf("droplet-upload-warning")))
						Eventually(errorStream).Should(Receive(MatchError("ohnos")))

						Consistently(nextEvent).ShouldNot(EqualEither(RetryUpload, UploadDropletComplete, Complete))
					})
				})
			})

			Context("when uploading the droplet is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.UploadDropletReturns(v2action.Job{}, v2action.Warnings{"droplet-upload-warning"}, nil)
				})

				It("sends the UploadingDroplet event", func() {
					Eventually(nextEvent).Should(Equal(UploadingDroplet))
					Expect(nextEvent()).To(Equal(UploadDropletComplete))
					Eventually(warningsStream).Should(Receive(ConsistOf("droplet-upload-warning")))
				})
			})
		})

		Context("when app bits are provided", func() {
			Context("when there is at least one unmatched resource", func() {
				BeforeEach(func() {
					fakeV2Actor.ResourceMatchReturns(nil, []v2action.Resource{{}}, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
				})

				It("returns resource match warnings", func() {
					Eventually(nextEvent).Should(Equal(ResourceMatching))
					Eventually(warningsStream).Should(Receive(ConsistOf("resource-warnings-1", "resource-warnings-2")))
				})

				Context("when creating the archive is successful", func() {
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
						Eventually(nextEvent).Should(Equal(CreatingArchive))
					})

					Context("when the upload is successful", func() {
						BeforeEach(func() {
							fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
						})

						It("sends a UploadingApplicationWithArchive event", func() {
							Eventually(nextEvent).Should(Equal(UploadingApplicationWithArchive))
							Expect(nextEvent()).To(Equal(UploadWithArchiveComplete))
							Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
						})
					})

					Context("when the upload fails", func() {
						Context("when the upload error is a retryable error", func() {
							var someErr error

							BeforeEach(func() {
								someErr = errors.New("I AM A BANANA")
								fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, ccerror.PipeSeekError{Err: someErr})
							})

							It("should send a RetryUpload event and retry uploading", func() {
								Eventually(nextEvent).Should(Equal(UploadingApplicationWithArchive))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
								Expect(nextEvent()).To(Equal(RetryUpload))

								Expect(nextEvent()).To(Equal(UploadingApplicationWithArchive))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
								Expect(nextEvent()).To(Equal(RetryUpload))

								Expect(nextEvent()).To(Equal(UploadingApplicationWithArchive))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
								Expect(nextEvent()).To(Equal(RetryUpload))

								Consistently(nextEvent).ShouldNot(EqualEither(RetryUpload, UploadWithArchiveComplete, Complete))
								Eventually(fakeV2Actor.UploadApplicationPackageCallCount()).Should(Equal(3))
								Expect(errorStream).To(Receive(MatchError(actionerror.UploadFailedError{Err: someErr})))
							})

						})

						Context("when the upload error is not a retryable error", func() {
							BeforeEach(func() {
								fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("dios mio"))
							})

							It("sends warnings and errors, then stops", func() {
								Eventually(nextEvent).Should(Equal(UploadingApplicationWithArchive))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
								Consistently(nextEvent).ShouldNot(EqualEither(RetryUpload, UploadWithArchiveComplete, Complete))
								Eventually(errorStream).Should(Receive(MatchError("dios mio")))
							})
						})
					})
				})

				Context("when creating the archive fails", func() {
					BeforeEach(func() {
						fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("some-error"))
					})

					It("raises an error", func() {
						Consistently(nextEvent).ShouldNot(Equal(Complete))
						Eventually(errorStream).Should(Receive(MatchError("some-error")))
					})
				})
			})

			Context("when all resources have been matched", func() {
				BeforeEach(func() {
					fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
				})

				It("sends the UploadingApplication event", func() {
					Eventually(nextEvent).Should(Equal(ResourceMatching))
					Eventually(warningsStream).Should(Receive(ConsistOf("resource-warnings-1", "resource-warnings-2")))
					Expect(nextEvent()).To(Equal(UploadingApplication))
				})

				Context("when the upload is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
					})

					It("uploads the application and completes", func() {
						Eventually(nextEvent).Should(Equal(UploadingApplication))
						Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
						Expect(nextEvent()).To(Equal(Complete))
					})
				})

				Context("when the upload fails", func() {
					BeforeEach(func() {
						fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("some-upload-error"))
					})

					It("returns an error", func() {
						Eventually(nextEvent).Should(Equal(UploadingApplication))
						Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
						Eventually(errorStream).Should(Receive(MatchError("some-upload-error")))
						Consistently(nextEvent).ShouldNot(Equal(Complete))
					})
				})
			})
		})

		Context("when a docker image is provided", func() {
			BeforeEach(func() {
				config.DesiredApplication.DockerImage = "hi-im-a-ge"

				fakeV2Actor.CreateApplicationReturns(config.DesiredApplication.Application, nil, nil)
			})

			It("should skip uploading anything", func() {
				Consistently(nextEvent).ShouldNot(EqualEither(UploadingDroplet, UploadingApplication))
			})
		})
	})
})
