package pushaction_test

import (
	"errors"
	"io/ioutil"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

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
			createdApp = config.DesiredApplication
			createdApp.GUID = "some-app-guid"

			fakeV2Actor.CreateApplicationReturns(createdApp, v2action.Warnings{"create-application-warnings-1", "create-application-warnings-2"}, nil)
		})

		JustBeforeEach(func() {
			Eventually(warningsStream).Should(Receive(ConsistOf("create-application-warnings-1", "create-application-warnings-2")))
			Eventually(eventStream).Should(Receive(Equal(ApplicationCreated)))
		})

		Context("when the route creation is successful", func() {
			var createdRoutes []v2action.Route

			BeforeEach(func() {
				createdRoutes = []v2action.Route{{Host: "banana", GUID: "some-route-guid"}}
				fakeV2Actor.CreateRouteReturns(createdRoutes[0], v2action.Warnings{"create-route-warnings-1", "create-route-warnings-2"}, nil)
			})

			JustBeforeEach(func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-route-warnings-1", "create-route-warnings-2")))
				Eventually(eventStream).Should(Receive(Equal(RouteCreated)))
			})

			Context("when binding the routes is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warnings-1", "bind-route-warnings-2"}, nil)
				})

				JustBeforeEach(func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("bind-route-warnings-1", "bind-route-warnings-2")))
					Eventually(eventStream).Should(Receive(Equal(RouteBound)))
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
						fakeV2Actor.ZipResourcesReturns(archivePath, nil)
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
						})
					})

					Context("when the upload errors", func() {
						Context("with a retryable error", func() {
							It("retries the download", func() {
								Skip("until error handling in api has been completed")
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
						fakeV2Actor.ZipResourcesReturns("", expectedErr)
					})

					It("sends warnings and errors, then stops", func() {
						Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
						Consistently(eventStream).ShouldNot(Receive())
					})
				})
			})

			Context("when there are no routes to bind", func() {
				BeforeEach(func() {
					config.CurrentRoutes = createdRoutes
				})

				It("should not send the RouteCreated event", func() {
					Eventually(warningsStream).Should(Receive())
					Consistently(eventStream).ShouldNot(Receive(Equal(RouteCreated)))
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
			Eventually(warningsStream).Should(Receive(ConsistOf("create-application-warnings-1", "create-application-warnings-2")))
			Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			Consistently(eventStream).ShouldNot(Receive())
		})
	})
})
