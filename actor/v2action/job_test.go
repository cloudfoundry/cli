package v2action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Job Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		fakeConfig                *v2actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v2actionfakes.FakeConfig)
		fakeConfig.OverallPollingTimeoutReturns(5 * time.Second)
		actor = NewActor(fakeCloudControllerClient, nil, fakeConfig)
	})

	Describe("PollJob", func() {
		Context("when the job starts queued and then finishes successfully", func() {
			BeforeEach(func() {
				counter := 0
				fakeCloudControllerClient.GetJobStub = func(guid string) (ccv2.Job, ccv2.Warnings, error) {
					if counter == 0 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusQueued,
							GUID:   guid,
						}, ccv2.Warnings{"warning-1"}, nil
					} else if counter == 1 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusRunning,
							GUID:   guid,
						}, ccv2.Warnings{"warning-2", "warning-3"}, nil
					} else {
						return ccv2.Job{
							Status: ccv2.JobStatusFinished,
							GUID:   guid,
						}, ccv2.Warnings{"warning-4"}, nil
					}
				}
			})

			It("should poll until completion", func() {
				warnings, err := actor.PollJob(ccv2.Job{GUID: "some-job-guid"})
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetJobCallCount()).To(Equal(3))
				jobGUID := fakeCloudControllerClient.GetJobArgsForCall(0)
				Expect(jobGUID).To(Equal("some-job-guid"))
				jobGUID = fakeCloudControllerClient.GetJobArgsForCall(1)
				Expect(jobGUID).To(Equal("some-job-guid"))
				jobGUID = fakeCloudControllerClient.GetJobArgsForCall(2)
				Expect(jobGUID).To(Equal("some-job-guid"))
			})
		})

		Context("when the job starts queued and then fails", func() {
			var jobFailureMessage string
			BeforeEach(func() {
				jobFailureMessage = "I am a banana!!!"

				counter := 0
				fakeCloudControllerClient.GetJobStub = func(guid string) (ccv2.Job, ccv2.Warnings, error) {
					if counter == 0 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusQueued,
							GUID:   guid,
						}, ccv2.Warnings{"warning-1"}, nil
					} else if counter == 1 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusRunning,
							GUID:   guid,
						}, ccv2.Warnings{"warning-2", "warning-3"}, nil
					} else {
						return ccv2.Job{
							Status: ccv2.JobStatusFailed,
							GUID:   guid,
							Error:  jobFailureMessage,
						}, ccv2.Warnings{"warning-4"}, nil
					}
				}
			})

			It("returns a JobFailedError", func() {
				warnings, err := actor.PollJob(ccv2.Job{GUID: "some-job-guid"})
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				Expect(err).To(MatchError(JobFailedError{
					JobGUID: "some-job-guid",
					Message: jobFailureMessage,
				}))

				Expect(fakeCloudControllerClient.GetJobCallCount()).To(Equal(3))
				jobGUID := fakeCloudControllerClient.GetJobArgsForCall(0)
				Expect(jobGUID).To(Equal("some-job-guid"))
				jobGUID = fakeCloudControllerClient.GetJobArgsForCall(1)
				Expect(jobGUID).To(Equal("some-job-guid"))
				jobGUID = fakeCloudControllerClient.GetJobArgsForCall(2)
				Expect(jobGUID).To(Equal("some-job-guid"))
			})
		})

		Context("when the CC returns an error", func() {
			var expectedError error
			BeforeEach(func() {
				expectedError = errors.New("I am a banana")
				fakeCloudControllerClient.GetJobReturns(
					ccv2.Job{}, ccv2.Warnings{"warning-1", "warning-2"}, expectedError)
			})

			It("returns the CC error", func() {
				warnings, err := actor.PollJob(ccv2.Job{GUID: "some-job-guid"})
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(err).To(MatchError(expectedError))
			})
		})

		Describe("Sleeps", func() {
			BeforeEach(func() {
				counter := 0
				fakeCloudControllerClient.GetJobStub = func(guid string) (ccv2.Job, ccv2.Warnings, error) {
					if counter == 0 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusQueued,
							GUID:   guid,
						}, ccv2.Warnings{"warning-1"}, nil
					} else if counter == 1 {
						counter += 1
						return ccv2.Job{
							Status: ccv2.JobStatusRunning,
							GUID:   guid,
						}, ccv2.Warnings{"warning-2"}, nil
					} else {
						return ccv2.Job{
							Status: ccv2.JobStatusFinished,
							GUID:   guid,
						}, ccv2.Warnings{"warning-3"}, nil
					}
				}
			})

			// This test makes an assumption that time.Sleep is called around
			// config.PollingInterval()
			It("sleep for the pre-defined time", func() {
				warnings, err := actor.PollJob(ccv2.Job{GUID: "some-job-guid"})
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3"))
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetJobCallCount()).To(Equal(3))
				Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(2))
			})
		})

		Describe("OverallPollingTimeout", func() {
			Context("when the job runs longer than the OverallPollingTimeout", func() {
				var overallPollingTimeout time.Duration

				BeforeEach(func() {
					counter := 0
					fakeCloudControllerClient.GetJobStub = func(guid string) (ccv2.Job, ccv2.Warnings, error) {
						if counter == 0 {
							counter += 1
							return ccv2.Job{
								Status: ccv2.JobStatusQueued,
								GUID:   guid,
							}, ccv2.Warnings{"warning-1"}, nil
						} else if counter == 1 {
							counter += 1
							return ccv2.Job{
								Status: ccv2.JobStatusRunning,
								GUID:   guid,
							}, ccv2.Warnings{"warning-2", "warning-3"}, nil
						} else {
							// Should never get here
							return ccv2.Job{
								Status: ccv2.JobStatusFinished,
								GUID:   guid,
							}, ccv2.Warnings{"warning-4"}, nil
						}
					}

					overallPollingTimeout = 100 * time.Millisecond
					fakeConfig.OverallPollingTimeoutReturns(overallPollingTimeout)
					fakeConfig.PollingIntervalReturns(60 * time.Millisecond)
				})

				It("raises a JobTimeoutError", func() {
					warnings, err := actor.PollJob(ccv2.Job{GUID: "some-job-guid"})

					Expect(warnings).To(ContainElement("warning-1"))
					Expect(err).To(MatchError(JobTimeoutError{
						Timeout: overallPollingTimeout,
						JobGUID: "some-job-guid",
					}))
				})

				// Fuzzy test to ensure that the overall function time isn't [far]
				// greater than the OverallPollingTimeout. Since this is partially
				// dependant on the speed of the system, the expectation is that the
				// function *should* never exceed twice the timeout.
				It("does not run [too much] longer than the timeout", func() {
					startTime := time.Now()
					actor.PollJob(ccv2.Job{GUID: "some-job-guid"})
					endTime := time.Now()

					Expect(endTime).To(BeTemporally("~", startTime, 2*overallPollingTimeout))
				})
			})
		})
	})
})
