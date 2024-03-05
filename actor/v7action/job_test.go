package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("PollUploadBuildpackJob", func() {
		var (
			jobURL   ccv3.JobURL
			warnings Warnings
			err      error
		)

		BeforeEach(func() {
			jobURL = ccv3.JobURL("http://example.com/the-job-url")
		})

		JustBeforeEach(func() {
			warnings, err = actor.PollUploadBuildpackJob(jobURL)
		})

		When("the cc client returns success", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, nil)
			})

			It("returns success and warnings", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warnings"))
			})

			It("calls poll job", func() {
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				actualJobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(actualJobURL).To(Equal(jobURL))
			})
		})

		When("the cc client returns error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"some-warnings"}, ccerror.V3JobFailedError{Detail: "some-err"})
			})

			It("returns the errors and warnings", func() {
				warnings, err = actor.PollUploadBuildpackJob(jobURL)

				Expect(err).To(MatchError(ccerror.V3JobFailedError{Detail: "some-err"}))
				Expect(warnings).To(ConsistOf("some-warnings"))
			})
		})
	})

	Describe("PollJobToEventStream", func() {
		const fakeJobURL ccv3.JobURL = "fake-job-fakeJobURL"

		BeforeEach(func() {
			fakeCloudControllerClient.PollJobToEventStreamCalls(func(jobURL ccv3.JobURL) chan ccv3.PollJobEvent {
				Expect(jobURL).To(Equal(fakeJobURL))

				fakeStream := make(chan ccv3.PollJobEvent)
				go func() {
					fakeStream <- ccv3.PollJobEvent{
						State:    constant.JobProcessing,
						Err:      nil,
						Warnings: ccv3.Warnings{"foo"},
					}
					fakeStream <- ccv3.PollJobEvent{
						State:    constant.JobPolling,
						Err:      nil,
						Warnings: ccv3.Warnings{"bar"},
					}
					fakeStream <- ccv3.PollJobEvent{
						State:    constant.JobFailed,
						Err:      errors.New("bad thing"),
						Warnings: ccv3.Warnings{"baz"},
					}
					close(fakeStream)
				}()

				return fakeStream
			})
		})

		It("converts the types in the stream", func() {
			stream := actor.PollJobToEventStream(fakeJobURL)
			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobProcessing,
				Err:      nil,
				Warnings: Warnings{"foo"},
			})))
			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobPolling,
				Err:      nil,
				Warnings: Warnings{"bar"},
			})))
			Eventually(stream).Should(Receive(Equal(PollJobEvent{
				State:    JobFailed,
				Err:      errors.New("bad thing"),
				Warnings: Warnings{"baz"},
			})))
			Eventually(stream).Should(BeClosed())
		})

		When("the input channel is nil", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobToEventStreamReturns(nil)
			})

			It("returns nil", func() {
				stream := actor.PollJobToEventStream(fakeJobURL)
				Expect(stream).To(BeNil())
			})
		})
	})
})
