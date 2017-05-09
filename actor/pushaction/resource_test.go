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
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)
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
				DesiredApplication: v2action.Application{
					GUID: "some-app-guid",
				},
			}

			resourcesToArchive = []v2action.Resource{{Filename: "file1"}, {Filename: "file2"}}
			config.AllResources = resourcesToArchive
		})

		JustBeforeEach(func() {
			archivePath, executeErr = actor.CreateArchive(config)
		})

		Context("when the zipping is successful", func() {
			var fakeArchivePath string
			BeforeEach(func() {
				fakeArchivePath = "some-archive-path"
				fakeV2Actor.ZipResourcesReturns(fakeArchivePath, nil)
			})

			It("returns the path to the zip", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(archivePath).To(Equal(fakeArchivePath))

				Expect(fakeV2Actor.ZipResourcesCallCount()).To(Equal(1))
				sourceDir, passedResources := fakeV2Actor.ZipResourcesArgsForCall(0)
				Expect(sourceDir).To(Equal("some-path"))
				Expect(passedResources).To(Equal(resourcesToArchive))
			})
		})

		Context("when creating the archive errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("oh no")
				fakeV2Actor.ZipResourcesReturns("", expectedErr)
			})

			It("sends errors and returns true", func() {
				Expect(executeErr).To(MatchError(expectedErr))
			})
		})
	})

	Describe("UploadPackage", func() {
		var (
			config          ApplicationConfig
			archivePath     string
			fakeProgressBar *pushactionfakes.FakeProgressBar
			eventStream     chan Event

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				DesiredApplication: v2action.Application{
					GUID: "some-app-guid",
				},
			}
			fakeProgressBar = new(pushactionfakes.FakeProgressBar)
			eventStream = make(chan Event)
		})

		AfterEach(func() {
			close(eventStream)
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.UploadPackage(config, archivePath, fakeProgressBar, eventStream)
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

						Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
						Eventually(eventStream).Should(Receive(Equal(UploadComplete)))
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
						Expect(existingResources).To(BeEmpty())
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

						Eventually(eventStream).Should(Receive(Equal(UploadingApplication)))
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
