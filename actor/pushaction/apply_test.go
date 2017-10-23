package pushaction_test

import (
	"errors"
	"io/ioutil"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			Path:          "some-path",
		}
		fakeProgressBar = new(pushactionfakes.FakeProgressBar)
	})

	JustBeforeEach(func() {
		configStream, eventStream, warningsStream, errorStream = actor.Apply(config, fakeProgressBar)
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

								Context("when the upload errors", func() {
									Context("with a retryable error", func() {
										BeforeEach(func() {
											fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, ccerror.PipeSeekError{})
										})

										It("retries the download up to three times", func() {
											Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
											Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(1))
											Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
											Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

											Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
											Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(2))
											Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
											Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

											Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
											Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(3))
											Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
											Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

											Eventually(errorStream).Should(Receive(Equal(actionerror.UploadFailedError{})))
										})
									})

									Context("with a generic error", func() {
										var expectedErr error

										BeforeEach(func() {
											expectedErr = errors.New("dios mio")
											fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, expectedErr)
										})

										It("sends warnings and errors, then stops", func() {
											Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
											Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
											Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
											Consistently(eventStream).ShouldNot(Receive())
										})
									})
								})
							})

							Context("when the archive creation errors", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("dios mio")
									fakeSharedActor.ZipDirectoryResourcesReturns("", expectedErr)
								})

								It("sends warnings and errors, then stops", func() {
									Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
									Consistently(eventStream).ShouldNot(Receive())
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

							Context("when the upload errors", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("dios mio")
									fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, expectedErr)
								})

								It("sends warnings and errors, then stops", func() {
									Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
									Consistently(eventStream).ShouldNot(Receive())
								})
							})
						})
					})

					Context("when a docker image is provided", func() {
						BeforeEach(func() {
							config.DesiredApplication.DockerImage = "some-docker-image-path"
						})

						It("skips achiving and uploading", func() {
							Eventually(configStream).Should(Receive())
							Eventually(eventStream).Should(Receive(Equal(Complete)))

							Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(0))
						})
					})
				})

				Context("when there are no services to bind", func() {
					BeforeEach(func() {
						services := map[string]v2action.ServiceInstance{
							"service_1": {Name: "service_1", GUID: "service_guid"},
						}
						config.CurrentServices = services
						config.DesiredServices = services
					})

					It("should not send the BoundServices event", func() {
						Eventually(eventStream).ShouldNot(Receive(Equal(ConfiguringServices)))
						Consistently(eventStream).ShouldNot(Receive(Equal(BoundServices)))
					})
				})

				Context("when mapping routes fails", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("dios mio")
						fakeV2Actor.BindServiceByApplicationAndServiceInstanceReturns(v2action.Warnings{"bind-service-warnings-1", "bind-service-warnings-2"}, expectedErr)
					})

					It("sends warnings and errors, then stops", func() {
						Eventually(eventStream).Should(Receive(Equal(ConfiguringServices)))
						Eventually(warningsStream).Should(Receive(ConsistOf("bind-service-warnings-1", "bind-service-warnings-2")))
						Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
						Consistently(eventStream).ShouldNot(Receive())
					})
				})
			})

			Context("when there are no routes to map", func() {
				BeforeEach(func() {
					config.CurrentRoutes = createdRoutes
				})

				It("should not send the RouteCreated event", func() {
					Eventually(warningsStream).Should(Receive())
					Consistently(eventStream).ShouldNot(Receive(Equal(CreatedRoutes)))
				})
			})

			Context("when mapping the routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					fakeV2Actor.MapRouteToApplicationReturns(v2action.Warnings{"map-route-warnings-1", "map-route-warnings-2"}, expectedErr)
				})

				It("sends warnings and errors, then stops", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("map-route-warnings-1", "map-route-warnings-2")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
					Consistently(eventStream).ShouldNot(Receive())
				})
			})
		})

		Context("when there are no routes to create", func() {
			BeforeEach(func() {
				config.DesiredRoutes[0].GUID = "some-route-guid"
			})

			It("should not send the RouteCreated event", func() {
				Eventually(eventStream).Should(Receive(Equal(CreatingAndMappingRoutes)))
				Eventually(warningsStream).Should(Receive())
				Consistently(eventStream).ShouldNot(Receive())
			})
		})

		Context("when the route creation errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("dios mio")
				fakeV2Actor.CreateRouteReturns(v2action.Route{}, v2action.Warnings{"create-route-warnings-1", "create-route-warnings-2"}, expectedErr)
			})

			It("sends warnings and errors, then stops", func() {
				Eventually(eventStream).Should(Receive(Equal(CreatingAndMappingRoutes)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warnings-1", "create-route-warnings-2")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				Consistently(eventStream).ShouldNot(Receive())
			})
		})

		Context("when routes are to be removed", func() {
			BeforeEach(func() {
				config.NoRoute = true
			})

			Context("when there are routes", func() {
				BeforeEach(func() {
					config.CurrentRoutes = []v2action.Route{{GUID: "some-route-guid-1"}, {GUID: "some-route-guid-2"}}
				})

				JustBeforeEach(func() {
					Eventually(eventStream).Should(Receive(Equal(UnmappingRoutes)))
				})

				Context("when the unmap is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.UnmapRouteFromApplicationReturnsOnCall(0, v2action.Warnings{"unmapping-route-warnings-1"}, nil)
						fakeV2Actor.UnmapRouteFromApplicationReturnsOnCall(1, v2action.Warnings{"unmapping-route-warnings-2"}, nil)
					})

					It("unmaps the routes and returns all warnings", func() {
						Eventually(warningsStream).Should(Receive(ConsistOf("unmapping-route-warnings-1", "unmapping-route-warnings-2")))
						Expect(streamsDrainedAndClosed(configStream, eventStream, warningsStream, errorStream)).To(BeTrue())

						Expect(fakeV2Actor.UnmapRouteFromApplicationCallCount()).To(Equal(2))
					})
				})

				Context("when unmapping routes fails", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("dios mio")
						fakeV2Actor.UnmapRouteFromApplicationReturns(v2action.Warnings{"unmapping-route-warnings-1", "unmapping-route-warnings-2"}, expectedErr)
					})

					It("sends warnings and errors, then stops", func() {
						Eventually(warningsStream).Should(Receive(ConsistOf("unmapping-route-warnings-1", "unmapping-route-warnings-2")))
						Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
						Consistently(eventStream).ShouldNot(Receive())
					})
				})
			})

			Context("when there are no routes", func() {
				BeforeEach(func() {
					config.CurrentRoutes = nil
				})

				It("does not send an UnmappingRoutes event", func() {
					Consistently(eventStream).ShouldNot(Receive(Equal(UnmappingRoutes)))
					Expect(streamsDrainedAndClosed(configStream, eventStream, warningsStream, errorStream)).To(BeTrue())

					Expect(fakeV2Actor.UnmapRouteFromApplicationCallCount()).To(Equal(0))
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
})
