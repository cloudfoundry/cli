package pushaction_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resources", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		actor = NewActor(fakeV2Actor, fakeSharedActor)
	})

	Describe("CreateArchive", func() {
		var (
			config ApplicationConfig

			archivePath string
			executeErr  error

			resourcesToArchive []v2action.Resource
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				Path: "some-path",
				DesiredApplication: Application{
					Application: v2action.Application{
						GUID: "some-app-guid",
					}},
			}

			resourcesToArchive = []v2action.Resource{{Filename: "file1"}, {Filename: "file2"}}
			config.UnmatchedResources = resourcesToArchive
		})

		JustBeforeEach(func() {
			archivePath, executeErr = actor.CreateArchive(config)
		})

		Context("when the source is an archive", func() {
			BeforeEach(func() {
				config.Archive = true
			})

			Context("when the zipping is successful", func() {
				var fakeArchivePath string

				BeforeEach(func() {
					fakeArchivePath = "some-archive-path"
					fakeSharedActor.ZipArchiveResourcesReturns(fakeArchivePath, nil)
				})

				It("returns the path to the zip", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(archivePath).To(Equal(fakeArchivePath))

					Expect(fakeSharedActor.ZipArchiveResourcesCallCount()).To(Equal(1))
					sourceDir, passedResources := fakeSharedActor.ZipArchiveResourcesArgsForCall(0)
					Expect(sourceDir).To(Equal("some-path"))
					sharedResourcesToArchive := actor.ConvertV2ResourcesToSharedResources(resourcesToArchive)
					Expect(passedResources).To(Equal(sharedResourcesToArchive))

				})
			})

			Context("when creating the archive errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh no")
					fakeSharedActor.ZipArchiveResourcesReturns("", expectedErr)
				})

				It("sends errors and returns true", func() {
					Expect(executeErr).To(MatchError(expectedErr))
				})
			})
		})

		Context("when the source is a directory", func() {
			Context("when the zipping is successful", func() {
				var fakeArchivePath string
				BeforeEach(func() {
					fakeArchivePath = "some-archive-path"
					fakeSharedActor.ZipDirectoryResourcesReturns(fakeArchivePath, nil)
				})

				It("returns the path to the zip", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(archivePath).To(Equal(fakeArchivePath))

					Expect(fakeSharedActor.ZipDirectoryResourcesCallCount()).To(Equal(1))
					sourceDir, passedResources := fakeSharedActor.ZipDirectoryResourcesArgsForCall(0)
					Expect(sourceDir).To(Equal("some-path"))
					sharedResourcesToArchive := actor.ConvertV2ResourcesToSharedResources(resourcesToArchive)
					Expect(passedResources).To(Equal(sharedResourcesToArchive))
				})
			})

			Context("when creating the archive errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh no")
					fakeSharedActor.ZipDirectoryResourcesReturns("", expectedErr)
				})

				It("sends errors and returns true", func() {
					Expect(executeErr).To(MatchError(expectedErr))
				})
			})
		})
	})

	Describe("SetMatchedResources", func() {
		var (
			inputConfig  ApplicationConfig
			outputConfig ApplicationConfig
			warnings     Warnings
		)
		JustBeforeEach(func() {
			outputConfig, warnings = actor.SetMatchedResources(inputConfig)
		})

		BeforeEach(func() {
			inputConfig.AllResources = []v2action.Resource{
				{Filename: "file-1"},
				{Filename: "file-2"},
			}
		})

		Context("when the resource matching is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.ResourceMatchReturns(
					[]v2action.Resource{{Filename: "file-1"}},
					[]v2action.Resource{{Filename: "file-2"}},
					v2action.Warnings{"warning-1"},
					nil,
				)
			})

			It("sets the matched and unmatched resources", func() {
				Expect(outputConfig.MatchedResources).To(ConsistOf(v2action.Resource{Filename: "file-1"}))
				Expect(outputConfig.UnmatchedResources).To(ConsistOf(v2action.Resource{Filename: "file-2"}))

				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})

		Context("when resource matching returns an error", func() {
			BeforeEach(func() {
				fakeV2Actor.ResourceMatchReturns(nil, nil, v2action.Warnings{"warning-1"}, errors.New("some-error"))
			})

			It("sets the unmatched resources to AllResources", func() {
				Expect(outputConfig.UnmatchedResources).To(Equal(inputConfig.AllResources))

				Expect(warnings).To(ConsistOf("warning-1"))
			})
		})
	})

	Describe("UploadPackage", func() {
		var (
			config ApplicationConfig

			warnings   Warnings
			executeErr error

			resources []v2action.Resource
		)

		BeforeEach(func() {
			resources = []v2action.Resource{
				{Filename: "file-1"},
				{Filename: "file-2"},
			}

			config = ApplicationConfig{
				DesiredApplication: Application{
					Application: v2action.Application{
						GUID: "some-app-guid",
					}},
				MatchedResources: resources,
			}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UploadPackage(config)
		})

		Context("when the upload is successful", func() {
			var uploadJob v2action.Job

			BeforeEach(func() {
				uploadJob.GUID = "some-job-guid"
				fakeV2Actor.UploadApplicationPackageReturns(uploadJob, v2action.Warnings{"upload-warning-1", "upload-warning-2"}, nil)
			})

			Context("when polling is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.PollJobReturns(v2action.Warnings{"poll-warning-1", "poll-warning-2"}, nil)
				})

				It("uploads the existing resources", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "poll-warning-1", "poll-warning-2"))

					Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
					appGUID, existingResources, reader, newResourcesLength := fakeV2Actor.UploadApplicationPackageArgsForCall(0)
					Expect(appGUID).To(Equal("some-app-guid"))
					Expect(existingResources).To(Equal(resources))
					Expect(reader).To(BeNil())
					Expect(newResourcesLength).To(BeNumerically("==", 0))

					Expect(fakeV2Actor.PollJobCallCount()).To(Equal(1))
					Expect(fakeV2Actor.PollJobArgsForCall(0)).To(Equal(uploadJob))
				})
			})

			Context("when the polling fails", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I can't let you do that starfox")
					fakeV2Actor.PollJobReturns(v2action.Warnings{"poll-warning-1", "poll-warning-2"}, expectedErr)
				})

				It("returns the warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "poll-warning-1", "poll-warning-2"))
				})
			})
		})

		Context("when the upload errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I can't let you do that starfox")
				fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warning-1", "upload-warning-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
			})
		})
	})

	Describe("UploadPackageWithArchive", func() {
		var (
			config          ApplicationConfig
			archivePath     string
			fakeProgressBar *pushactionfakes.FakeProgressBar
			eventStream     chan Event

			warnings   Warnings
			executeErr error

			resources []v2action.Resource
		)

		BeforeEach(func() {
			resources = []v2action.Resource{
				{Filename: "file-1"},
				{Filename: "file-2"},
			}

			config = ApplicationConfig{
				DesiredApplication: Application{
					Application: v2action.Application{
						GUID: "some-app-guid",
					}},
				MatchedResources: resources,
			}
			fakeProgressBar = new(pushactionfakes.FakeProgressBar)
			eventStream = make(chan Event)
		})

		AfterEach(func() {
			close(eventStream)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UploadPackageWithArchive(config, archivePath, fakeProgressBar, eventStream)
		})

		Context("when the archive can be accessed properly", func() {
			BeforeEach(func() {
				tmpfile, err := ioutil.TempFile("", "fake-archive")
				Expect(err).ToNot(HaveOccurred())
				_, err = tmpfile.Write([]byte("123456"))
				Expect(err).ToNot(HaveOccurred())
				Expect(tmpfile.Close()).ToNot(HaveOccurred())

				archivePath = tmpfile.Name()
			})

			AfterEach(func() {
				Expect(os.Remove(archivePath)).ToNot(HaveOccurred())
			})

			Context("when the upload is successful", func() {
				var (
					progressBarReader io.Reader
					uploadJob         v2action.Job
				)

				BeforeEach(func() {
					uploadJob.GUID = "some-job-guid"
					fakeV2Actor.UploadApplicationPackageReturns(uploadJob, v2action.Warnings{"upload-warning-1", "upload-warning-2"}, nil)

					progressBarReader = strings.NewReader("123456")
					fakeProgressBar.NewProgressBarWrapperReturns(progressBarReader)

					go func() {
						defer GinkgoRecover()

						Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
						Eventually(eventStream).Should(Receive(Equal(UploadWithArchiveComplete)))
					}()
				})

				Context("when the polling is successful", func() {
					BeforeEach(func() {
						fakeV2Actor.PollJobReturns(v2action.Warnings{"poll-warning-1", "poll-warning-2"}, nil)
					})

					It("returns the warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "poll-warning-1", "poll-warning-2"))

						Expect(fakeV2Actor.UploadApplicationPackageCallCount()).To(Equal(1))
						appGUID, existingResources, _, newResourcesLength := fakeV2Actor.UploadApplicationPackageArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))
						Expect(existingResources).To(Equal(resources))
						Expect(newResourcesLength).To(BeNumerically("==", 6))

						Expect(fakeV2Actor.PollJobCallCount()).To(Equal(1))
						Expect(fakeV2Actor.PollJobArgsForCall(0)).To(Equal(uploadJob))
					})

					It("passes the file reader to the progress bar", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(fakeProgressBar.NewProgressBarWrapperCallCount()).To(Equal(1))
						_, size := fakeProgressBar.NewProgressBarWrapperArgsForCall(0)
						Expect(size).To(BeNumerically("==", 6))
					})
				})

				Context("when the polling fails", func() {
					var expectedErr error

					BeforeEach(func() {
						expectedErr = errors.New("I can't let you do that starfox")
						fakeV2Actor.PollJobReturns(v2action.Warnings{"poll-warning-1", "poll-warning-2"}, expectedErr)
					})

					It("returns the warnings", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2", "poll-warning-1", "poll-warning-2"))
					})
				})
			})

			Context("when the upload errors", func() {
				var (
					expectedErr error
					done        chan bool
				)

				BeforeEach(func() {
					expectedErr = errors.New("I can't let you do that starfox")
					fakeV2Actor.UploadApplicationPackageReturns(v2action.Job{}, v2action.Warnings{"upload-warning-1", "upload-warning-2"}, expectedErr)

					done = make(chan bool)

					go func() {
						defer GinkgoRecover()

						Eventually(eventStream).Should(Receive(Equal(UploadingApplicationWithArchive)))
						Consistently(eventStream).ShouldNot(Receive())
						done <- true
					}()
				})

				AfterEach(func() {
					close(done)
				})

				It("returns the error and warnings", func() {
					Eventually(done).Should(Receive())
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("upload-warning-1", "upload-warning-2"))
				})
			})
		})

		Context("when the archive returns any access errors", func() {
			It("returns the error", func() {
				_, ok := executeErr.(*os.PathError)
				Expect(ok).To(BeTrue())
			})
		})
	})
})
