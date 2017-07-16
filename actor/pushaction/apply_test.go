package pushaction_test

import (
	"errors"
	"io/ioutil"

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
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor

		config          ApplicationConfig
		fakeProgressBar *pushactionfakes.FakeProgressBar

		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error
		configStream   <-chan ApplicationConfig
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)

		config = ApplicationConfig{
			DesiredApplication: v2action.Application{
				Name:      "some-app-name",
				SpaceGUID: "some-space-guid",
			},
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
				Eventually(eventStream).Should(Receive(Equal(ConfiguringRoutes)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warnings-1", "create-route-warnings-2")))
				Eventually(eventStream).Should(Receive(Equal(CreatedRoutes)))
			})

			Context("when binding the routes is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warnings-1", "bind-route-warnings-2"}, nil)
				})

				JustBeforeEach(func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warnings-1", "bind-route-warnings-2")))
					Eventually(eventStream).Should(Receive(Equal(BoundRoutes)))
				})

				Context("when resource matching happens", func() {
					BeforeEach(func() {
						fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"resource-warnings-1", "resource-warnings-2"}, nil)
					})

					JustBeforeEach(func() {
						Eventually(eventStream).Should(Receive(Equal(ResourceMatching)))
						Eventually(warningsStream).Should(Receive(ConsistOf("resource-warnings-1", "resource-warnings-2")))
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
							fakeV2Actor.ZipDirectoryResourcesReturns(archivePath, nil)
						})

						JustBeforeEach(func() {
							Eventually(eventStream).Should(Receive(Equal(CreatingArchive)))
						})

						Context("when the upload is successful", func() {
							BeforeEach(func() {
								fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, nil)
							})

							JustBeforeEach(func() {
								Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
								Eventually(eventStream).Should(Receive(Equal(UploadComplete)))
								Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
							})

							It("sends the updated config and a complete event", func() {
								Eventually(configStream).Should(Receive(Equal(ApplicationConfig{
									CurrentApplication: createdApp,
									DesiredApplication: createdApp,
									CurrentRoutes:      createdRoutes,
									DesiredRoutes:      createdRoutes,
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
									Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
									Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(1))
									Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
									Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

									Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
									Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(2))
									Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
									Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

									Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
									Eventually(fakeProgressBar.NewProgressBarWrapperCallCount).Should(Equal(3))
									Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
									Eventually(eventStream).Should(Receive(Equal(RetryUpload)))

									Eventually(errorStream).Should(Receive(Equal(UploadFailedError{})))
								})
							})

							Context("with a generic error", func() {
								var expectedErr error

								BeforeEach(func() {
									expectedErr = errors.New("dios mio")
									fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warnings-1", "upload-warnings-2"}, expectedErr)
								})

								It("sends warnings and errors, then stops", func() {
									Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
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
							fakeV2Actor.ZipDirectoryResourcesReturns("", expectedErr)
						})

						It("sends warnings and errors, then stops", func() {
							Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
							Consistently(eventStream).ShouldNot(Receive())
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

						Expect(fakeV2Actor.ZipDirectoryResourcesCallCount()).To(Equal(0))
					})
				})
			})

			Context("when there are no routes to bind", func() {
				BeforeEach(func() {
					config.CurrentRoutes = createdRoutes
				})

				It("should not send the RouteCreated event", func() {
					Eventually(warningsStream).Should(Receive())
					Consistently(eventStream).ShouldNot(Receive(Equal(CreatedRoutes)))
				})
			})

			Context("when binding the routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("dios mio")
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warnings-1", "bind-route-warnings-2"}, expectedErr)
				})

				It("sends warnings and errors, then stops", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warnings-1", "bind-route-warnings-2")))
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
				Eventually(eventStream).Should(Receive(Equal(ConfiguringRoutes)))
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
				Eventually(eventStream).Should(Receive(Equal(ConfiguringRoutes)))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warnings-1", "create-route-warnings-2")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				Consistently(eventStream).ShouldNot(Receive())
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
