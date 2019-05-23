package v7pushaction_test

import (
	"errors"
	"strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateDropletForApplication", func() {
	var (
		actor           *Actor
		fakeV7Actor     *v7pushactionfakes.FakeV7Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor

		returnedPushPlan PushPlan
		paramPlan        PushPlan
		fakeProgressBar  *v7pushactionfakes.FakeProgressBar

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, fakeSharedActor = getTestPushActor()

		fakeProgressBar = new(v7pushactionfakes.FakeProgressBar)

		fakeSharedActor.ReadArchiveReturns(new(v7pushactionfakes.FakeReadCloser), 0, nil)

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			returnedPushPlan, warnings, executeErr = actor.CreateDropletForApplication(paramPlan, eventStream, fakeProgressBar)
		})
	})

	When("the plan has a droplet path specified", func() {
		BeforeEach(func() {
			paramPlan.DropletPath = "path/to/some-droplet.tgz"
		})

		When("creating the droplet for app fails", func() {
			var createError = errors.New("create droplet failed")

			BeforeEach(func() {
				fakeV7Actor.CreateApplicationDropletReturns(
					v7action.Droplet{},
					v7action.Warnings{"create-droplet-warning"},
					createError,
				)
			})

			It("only records an event for creating droplet", func() {
				Expect(events).To(Equal([]Event{
					CreatingDroplet,
				}))
			})

			It("returns the unmodified plan, warnings, and error", func() {
				Expect(returnedPushPlan).To(Equal(paramPlan))
				Expect(warnings).To(ConsistOf("create-droplet-warning"))
				Expect(executeErr).To(Equal(createError))
			})
		})

		When("reading droplet file fails", func() {
			var readError = errors.New("reading file failed")

			BeforeEach(func() {
				fakeV7Actor.CreateApplicationDropletReturns(
					v7action.Droplet{},
					v7action.Warnings{"create-droplet-warning"},
					nil,
				)

				fakeSharedActor.ReadArchiveReturns(
					nil,
					0,
					readError,
				)
			})

			It("records events for creating droplet, reading archive", func() {
				Expect(events).To(Equal([]Event{
					CreatingDroplet,
					ReadingArchive,
				}))
			})

			It("returns the unmodified plan, warnings, and error", func() {
				Expect(returnedPushPlan).To(Equal(paramPlan))
				Expect(warnings).To(ConsistOf("create-droplet-warning"))
				Expect(executeErr).To(Equal(readError))
			})
		})

		When("uploading droplet fails", func() {
			var createdDroplet = v7action.Droplet{GUID: "created-droplet-guid"}
			var progressReader = strings.NewReader("123456")
			var uploadError = errors.New("uploading droplet failed")

			BeforeEach(func() {
				fakeV7Actor.CreateApplicationDropletReturns(
					createdDroplet,
					v7action.Warnings{"create-droplet-warning"},
					nil,
				)

				fakeSharedActor.ReadArchiveReturns(
					new(v7pushactionfakes.FakeReadCloser),
					int64(128),
					nil,
				)

				fakeProgressBar.NewProgressBarWrapperReturns(progressReader)

				fakeV7Actor.UploadDropletReturns(
					v7action.Warnings{"upload-droplet-warning"},
					uploadError,
				)
			})

			It("tries to upload the droplet with given path", func() {
				Expect(fakeV7Actor.UploadDropletCallCount()).To(Equal(1))
				givenDropletGUID, givenDropletPath, givenProgressReader, givenSize := fakeV7Actor.UploadDropletArgsForCall(0)
				Expect(givenDropletGUID).To(Equal(createdDroplet.GUID))
				Expect(givenDropletPath).To(Equal(paramPlan.DropletPath))
				Expect(givenProgressReader).To(Equal(progressReader))
				Expect(givenSize).To(Equal(int64(128)))
			})

			It("records events for creating droplet, reading archive, uploading droplet", func() {
				Expect(events).To(Equal([]Event{
					CreatingDroplet,
					ReadingArchive,
					UploadingDroplet,
				}))
			})

			It("returns the unmodified plan, warnings, and error", func() {
				Expect(returnedPushPlan).To(Equal(paramPlan))
				Expect(warnings).To(ConsistOf("create-droplet-warning", "upload-droplet-warning"))
				Expect(executeErr).To(Equal(uploadError))
			})
		})

		When("a retryable failure occurs", func() {
			var createdDroplet = v7action.Droplet{GUID: "created-droplet-guid"}
			var progressReader = strings.NewReader("123456")
			var retryableError = ccerror.PipeSeekError{
				Err: errors.New("network error"),
			}

			BeforeEach(func() {
				fakeV7Actor.CreateApplicationDropletReturns(
					createdDroplet,
					v7action.Warnings{"create-droplet-warning"},
					nil,
				)

				fakeSharedActor.ReadArchiveReturns(
					new(v7pushactionfakes.FakeReadCloser),
					int64(128),
					nil,
				)

				fakeProgressBar.NewProgressBarWrapperReturns(progressReader)
			})

			When("all upload attempts fail", func() {
				BeforeEach(func() {
					fakeV7Actor.UploadDropletReturns(
						v7action.Warnings{"upload-droplet-warning"},
						retryableError,
					)
				})

				It("records events for creating droplet, reading archive, uploading droplet, retrying", func() {
					Expect(events).To(Equal([]Event{
						CreatingDroplet,
						ReadingArchive,
						UploadingDroplet,
						RetryUpload,
						ReadingArchive,
						UploadingDroplet,
						RetryUpload,
						ReadingArchive,
						UploadingDroplet,
						RetryUpload,
					}))
				})

				It("returns the unmodified plan, all warnings, and wrapped error", func() {
					Expect(returnedPushPlan).To(Equal(paramPlan))
					Expect(warnings).To(ConsistOf(
						"create-droplet-warning",
						"upload-droplet-warning",
						"upload-droplet-warning",
						"upload-droplet-warning",
					))
					Expect(executeErr).To(Equal(actionerror.UploadFailedError{
						Err: retryableError.Err,
					}))
				})
			})

			When("the upload eventually succeeds", func() {
				BeforeEach(func() {
					fakeV7Actor.UploadDropletReturnsOnCall(0,
						v7action.Warnings{"upload-droplet-warning"},
						retryableError,
					)

					fakeV7Actor.UploadDropletReturnsOnCall(1,
						v7action.Warnings{"upload-droplet-warning"},
						retryableError,
					)

					fakeV7Actor.UploadDropletReturnsOnCall(2,
						v7action.Warnings{"upload-droplet-warning"},
						nil,
					)
				})

				It("records events for creating droplet, reading archive, uploading droplet, retrying, completing", func() {
					Expect(events).To(Equal([]Event{
						CreatingDroplet,
						ReadingArchive,
						UploadingDroplet,
						RetryUpload,
						ReadingArchive,
						UploadingDroplet,
						RetryUpload,
						ReadingArchive,
						UploadingDroplet,
						UploadDropletComplete,
					}))
				})

				It("returns the modified plan, all warnings, and no error", func() {
					Expect(returnedPushPlan.DropletGUID).To(Equal(createdDroplet.GUID))
					Expect(warnings).To(ConsistOf(
						"create-droplet-warning",
						"upload-droplet-warning",
						"upload-droplet-warning",
						"upload-droplet-warning",
					))
					Expect(executeErr).NotTo(HaveOccurred())
				})
			})
		})

		When("upload completes successfully", func() {
			var createdDroplet = v7action.Droplet{GUID: "created-droplet-guid"}
			var progressReader = strings.NewReader("123456")

			BeforeEach(func() {
				fakeV7Actor.CreateApplicationDropletReturns(
					createdDroplet,
					v7action.Warnings{"create-droplet-warning"},
					nil,
				)

				fakeSharedActor.ReadArchiveReturns(
					new(v7pushactionfakes.FakeReadCloser),
					int64(128),
					nil,
				)

				fakeProgressBar.NewProgressBarWrapperReturns(progressReader)

				fakeV7Actor.UploadDropletReturns(
					v7action.Warnings{"upload-droplet-warning"},
					nil,
				)
			})

			It("records events for creating droplet, reading archive, uploading droplet, retrying, completing", func() {
				Expect(events).To(Equal([]Event{
					CreatingDroplet,
					ReadingArchive,
					UploadingDroplet,
					UploadDropletComplete,
				}))
			})

			It("returns the modified plan, all warnings, and no error", func() {
				Expect(returnedPushPlan.DropletGUID).To(Equal(createdDroplet.GUID))
				Expect(warnings).To(ConsistOf(
					"create-droplet-warning",
					"upload-droplet-warning",
				))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})
	})
})
