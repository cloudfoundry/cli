package pushaction_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func actualizedStreamsDrainedAndClosed(
	configStream <-chan PushPlan,
	eventStream <-chan Event,
	warningsStream <-chan Warnings,
	errorStream <-chan error,
) bool {
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
func getNextEvent(c <-chan PushPlan, e <-chan Event, w <-chan Warnings) func() Event {
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

var _ = Describe("Actualize", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeV3Actor     *pushactionfakes.FakeV3Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor

		state           PushPlan
		fakeProgressBar *pushactionfakes.FakeProgressBar

		stateStream    <-chan PushPlan
		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeV3Actor = new(pushactionfakes.FakeV3Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		fakeSharedActor.ReadArchiveReturns(new(pushactionfakes.FakeReadCloser), 0, nil)
		actor = NewActor(fakeV2Actor, fakeV3Actor, fakeSharedActor)

		fakeProgressBar = new(pushactionfakes.FakeProgressBar)
		state = PushPlan{
			Application: v3action.Application{
				Name: "some-app",
			},
			SpaceGUID: "some-space-guid",
		}
	})

	AfterEach(func() {
		Eventually(actualizedStreamsDrainedAndClosed(stateStream, eventStream, warningsStream, errorStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		stateStream, eventStream, warningsStream, errorStream = actor.Actualize(state, fakeProgressBar)
	})

	Describe("application creation", func() {
		When("the application exists", func() {
			BeforeEach(func() {
				state.Application.GUID = "some-app-guid"
			})

			It("returns a skipped app creation event", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(SkippingApplicationCreation))

				Eventually(stateStream).Should(Receive(MatchFields(IgnoreExtras,
					Fields{
						"Application": Equal(v3action.Application{
							Name: "some-app",
							GUID: "some-app-guid",
						}),
					})))

				Consistently(fakeV3Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
			})
		})

		When("the application does not exist", func() {
			When("the creation is successful", func() {
				var expectedApp v3action.Application

				BeforeEach(func() {
					expectedApp = v3action.Application{
						GUID: "some-app-guid",
						Name: "some-app",
					}

					fakeV3Actor.CreateApplicationInSpaceReturns(expectedApp, v3action.Warnings{"some-app-warnings"}, nil)
				})

				It("returns an app created event, warnings, and updated state", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("some-app-warnings")))
					Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatedApplication))
					Eventually(stateStream).Should(Receive(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(expectedApp),
						})))
				})

				It("creates the application", func() {
					Eventually(fakeV3Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
					passedApp, passedSpaceGUID := fakeV3Actor.CreateApplicationInSpaceArgsForCall(0)
					Expect(passedApp).To(Equal(state.Application))
					Expect(passedSpaceGUID).To(Equal(state.SpaceGUID))
				})
			})

			When("the creation errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("SPICY!!")

					fakeV3Actor.CreateApplicationInSpaceReturns(v3action.Application{}, v3action.Warnings{"some-app-warnings"}, expectedErr)
				})

				It("returns warnings and error", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("some-app-warnings")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				})
			})
		})
	})

	Describe("package upload", func() {
		When("app bits are provided", func() {
			BeforeEach(func() {
				state = PushPlan{
					Application: v3action.Application{
						Name: "some-app",
						GUID: "some-app-guid",
					},
					BitsPath: "/some-bits-path",
					AllResources: []sharedaction.Resource{
						{Filename: "some-filename", Size: 6},
					},
					MatchedResources: []sharedaction.Resource{
						{Filename: "some-matched-filename", Size: 6},
					},
				}
			})

			It("creates the archive", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatingArchive))

				Eventually(fakeSharedActor.ZipDirectoryResourcesCallCount).Should(Equal(1))
				bitsPath, resources := fakeSharedActor.ZipDirectoryResourcesArgsForCall(0)
				Expect(bitsPath).To(Equal("/some-bits-path"))
				Expect(resources).To(ConsistOf(sharedaction.Resource{
					Filename: "some-filename",
					Size:     6,
				}))
			})

			When("the archive creation is successful", func() {
				BeforeEach(func() {
					fakeSharedActor.ZipDirectoryResourcesReturns("/some/archive/path", nil)
				})

				It("creates the package", func() {
					Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatingPackage))

					Eventually(fakeV3Actor.CreateBitsPackageByApplicationCallCount).Should(Equal(1))
					Expect(fakeV3Actor.CreateBitsPackageByApplicationArgsForCall(0)).To(Equal("some-app-guid"))
				})

				When("the package creation is successful", func() {
					BeforeEach(func() {
						fakeV3Actor.CreateBitsPackageByApplicationReturns(v3action.Package{GUID: "some-guid"}, v3action.Warnings{"some-create-package-warning"}, nil)
					})

					It("reads the archive", func() {
						Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(ReadingArchive))
						Eventually(fakeSharedActor.ReadArchiveCallCount).Should(Equal(1))
						Expect(fakeSharedActor.ReadArchiveArgsForCall(0)).To(Equal("/some/archive/path"))
					})

					When("reading the archive is successful", func() {
						BeforeEach(func() {
							fakeReadCloser := new(pushactionfakes.FakeReadCloser)
							fakeSharedActor.ReadArchiveReturns(fakeReadCloser, 6, nil)
						})

						It("uploads the bits package", func() {
							Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
							Eventually(fakeV3Actor.UploadBitsPackageCallCount).Should(Equal(1))
							pkg, resource, _, size := fakeV3Actor.UploadBitsPackageArgsForCall(0)

							Expect(pkg).To(Equal(v3action.Package{GUID: "some-guid"}))
							Expect(resource).To(ConsistOf(sharedaction.Resource{
								Filename: "some-matched-filename",
								Size:     6,
							}))
							Expect(size).To(BeNumerically("==", 6))
						})

						When("the upload is successful", func() {
							BeforeEach(func() {
								fakeV3Actor.UploadBitsPackageReturns(v3action.Package{GUID: "some-guid"}, v3action.Warnings{"some-upload-package-warning"}, nil)
							})

							It("returns an upload complete event and warnings", func() {
								Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
								Eventually(warningsStream).Should(Receive(ConsistOf("some-upload-package-warning")))
								Eventually(eventStream).Should(Receive(Equal(UploadWithArchiveComplete)))
							})

							When("the upload errors", func() {
								When("the upload error is a retryable error", func() {
									var someErr error

									BeforeEach(func() {
										someErr = errors.New("I AM A BANANA")
										fakeV3Actor.UploadBitsPackageReturns(v3action.Package{}, v3action.Warnings{"upload-warnings-1", "upload-warnings-2"}, ccerror.PipeSeekError{Err: someErr})
									})

									It("should send a RetryUpload event and retry uploading", func() {
										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
										Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(RetryUpload))

										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
										Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(RetryUpload))

										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
										Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(RetryUpload))

										Consistently(getNextEvent(stateStream, eventStream, warningsStream)).ShouldNot(EqualEither(RetryUpload, UploadWithArchiveComplete, Complete))
										Eventually(fakeV3Actor.UploadBitsPackageCallCount).Should(Equal(3))
										Expect(errorStream).To(Receive(MatchError(actionerror.UploadFailedError{Err: someErr})))
									})

								})

								When("the upload error is not a retryable error", func() {
									BeforeEach(func() {
										fakeV3Actor.UploadBitsPackageReturns(v3action.Package{}, v3action.Warnings{"upload-warnings-1", "upload-warnings-2"}, errors.New("dios mio"))
									})

									It("sends warnings and errors, then stops", func() {
										Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadingApplicationWithArchive))
										Eventually(warningsStream).Should(Receive(ConsistOf("upload-warnings-1", "upload-warnings-2")))
										Consistently(getNextEvent(stateStream, eventStream, warningsStream)).ShouldNot(EqualEither(RetryUpload, UploadWithArchiveComplete, Complete))
										Eventually(errorStream).Should(Receive(MatchError("dios mio")))
									})
								})
							})
						})

						When("reading the archive fails", func() {
							BeforeEach(func() {
								fakeSharedActor.ReadArchiveReturns(nil, 0, errors.New("the bits!"))
							})

							It("returns an error", func() {
								Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(ReadingArchive))
								Eventually(errorStream).Should(Receive(MatchError("the bits!")))
							})
						})
					})

					When("the package creation errors", func() {
						BeforeEach(func() {
							fakeV3Actor.CreateBitsPackageByApplicationReturns(v3action.Package{}, v3action.Warnings{"package-creation-warning"}, errors.New("the bits!"))
						})

						It("it returns errors and warnings", func() {
							Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatingPackage))

							Eventually(warningsStream).Should(Receive(ConsistOf("package-creation-warning")))
							Eventually(errorStream).Should(Receive(MatchError("the bits!")))
						})
					})
				})

				When("the archive creation errors", func() {
					BeforeEach(func() {
						fakeSharedActor.ZipDirectoryResourcesReturns("", errors.New("oh no"))
					})

					It("returns an error and exits", func() {
						Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatingArchive))

						Eventually(errorStream).Should(Receive(MatchError("oh no")))
					})
				})
			})
		})
	})

	Describe("polling package", func() {
		When("the the polling is succesful", func() {
			BeforeEach(func() {
				fakeV3Actor.PollPackageReturns(v3action.Package{}, v3action.Warnings{"some-poll-package-warning"}, nil)
			})

			It("returns warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadWithArchiveComplete))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-poll-package-warning")))
			})

		})

		When("the the polling returns an error", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV3Actor.PollPackageReturns(v3action.Package{}, v3action.Warnings{"some-poll-package-warning"}, someErr)
			})

			It("returns errors and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(UploadWithArchiveComplete))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-poll-package-warning")))
				Eventually(errorStream).Should(Receive(MatchError(someErr)))
			})
		})
	})

	Describe("staging package", func() {
		BeforeEach(func() {
			fakeV3Actor.PollPackageReturns(v3action.Package{GUID: "some-pkg-guid"}, nil, nil)
		})

		It("stages the application using the package guid", func() {
			Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(StartingStaging))
			Eventually(fakeV3Actor.StageApplicationPackageCallCount).Should(Equal(1))
			Expect(fakeV3Actor.StageApplicationPackageArgsForCall(0)).To(Equal("some-pkg-guid"))
		})

		When("staging is successful", func() {
			BeforeEach(func() {
				fakeV3Actor.StageApplicationPackageReturns(v3action.Build{GUID: "some-build-guid"}, v3action.Warnings{"some-staging-warning"}, nil)
			})

			It("returns a polling build event and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(StartingStaging))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-staging-warning")))
				Eventually(eventStream).Should(Receive(Equal(PollingBuild)))
			})
		})

		When("staging errors", func() {
			BeforeEach(func() {
				fakeV3Actor.StageApplicationPackageReturns(v3action.Build{}, v3action.Warnings{"some-staging-warning"}, errors.New("ahhh, i failed"))
			})

			It("returns errors and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(StartingStaging))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-staging-warning")))
				Eventually(errorStream).Should(Receive(MatchError("ahhh, i failed")))
			})
		})
	})

	Describe("polling build", func() {
		When("the the polling is succesful", func() {
			BeforeEach(func() {
				fakeV3Actor.PollBuildReturns(v3action.Droplet{}, v3action.Warnings{"some-poll-build-warning"}, nil)
			})

			It("returns a staging complete event and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(PollingBuild))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-poll-build-warning")))
				Eventually(eventStream).Should(Receive(Equal(StagingComplete)))
			})
		})

		When("the the polling returns an error", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV3Actor.PollBuildReturns(v3action.Droplet{}, v3action.Warnings{"some-poll-build-warning"}, someErr)
			})

			It("returns errors and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(PollingBuild))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-poll-build-warning")))
				Eventually(errorStream).Should(Receive(MatchError(someErr)))
			})
		})
	})

	Describe("setting droplet", func() {
		When("setting the droplet is successful", func() {
			BeforeEach(func() {
				fakeV3Actor.SetApplicationDropletReturns(v3action.Warnings{"some-set-droplet-warning"}, nil)
			})

			It("returns a SetDropletComplete event and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(SettingDroplet))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-set-droplet-warning")))
				Eventually(eventStream).Should(Receive(Equal(SetDropletComplete)))
			})
		})

		When("setting the droplet errors", func() {
			BeforeEach(func() {
				fakeV3Actor.SetApplicationDropletReturns(v3action.Warnings{"some-set-droplet-warning"}, errors.New("the climate is arid"))
			})

			It("returns an error and warnings", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(SettingDroplet))
				Eventually(warningsStream).Should(Receive(ConsistOf("some-set-droplet-warning")))
				Eventually(errorStream).Should(Receive(MatchError("the climate is arid")))
			})
		})
	})

	When("all operations are finished", func() {
		It("returns a complete event", func() {
			Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(Complete))
		})
	})
})
