package v3action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Build Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
		fakeConfig                *v3actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v3actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig)
	})

	Describe("StagePackage", func() {
		var (
			buildStream    <-chan Build
			warningsStream <-chan Warnings
			errorStream    <-chan error

			buildGUID string
		)

		AfterEach(func() {
			Eventually(errorStream).Should(BeClosed())
			Eventually(warningsStream).Should(BeClosed())
			Eventually(buildStream).Should(BeClosed())
		})

		JustBeforeEach(func() {
			buildStream, warningsStream, errorStream = actor.StagePackage("some-package-guid", "some-app")
		})

		Context("when the creation is successful", func() {
			BeforeEach(func() {
				buildGUID = "some-build-guid"
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{GUID: buildGUID, State: ccv3.BuildStateStaging}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, nil)
				fakeConfig.StagingTimeoutReturns(time.Minute)
			})

			Context("when the polling is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: ccv3.BuildStateStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
					fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{GUID: buildGUID, State: ccv3.BuildStateStaged}, ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, nil)
				})

				It("creates the build", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					Eventually(buildStream).Should(Receive(Equal(Build{GUID: buildGUID, State: ccv3.BuildStateStaged})))

					Expect(fakeCloudControllerClient.CreateBuildCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateBuildArgsForCall(0)).To(Equal(ccv3.Build{
						Package: ccv3.Package{
							GUID: "some-package-guid",
						},
					}))
				})

				It("polls until the build status is not 'STAGING'", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					Eventually(buildStream).Should(Receive())

					Expect(fakeCloudControllerClient.GetBuildCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetBuildArgsForCall(0)).To(Equal(buildGUID))
					Expect(fakeCloudControllerClient.GetBuildArgsForCall(1)).To(Equal(buildGUID))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
				})

				Context("when polling returns a failed build", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetBuildReturnsOnCall(
							1,
							ccv3.Build{
								GUID:  buildGUID,
								State: ccv3.BuildStateFailed,
								Error: "some staging error",
							},
							ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, nil)
					})

					It("returns an error and all warnings", func() {
						Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
						Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
						Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
						stagingErr := errors.New("some staging error")
						Eventually(errorStream).Should(Receive(&stagingErr))
						Eventually(buildStream).ShouldNot(Receive())

						Expect(fakeCloudControllerClient.GetBuildCallCount()).To(Equal(2))
						Expect(fakeCloudControllerClient.GetBuildArgsForCall(0)).To(Equal(buildGUID))
						Expect(fakeCloudControllerClient.GetBuildArgsForCall(1)).To(Equal(buildGUID))

						Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
					})
				})
			})

			Context("when polling times out", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = StagingTimeoutError{AppName: "some-app", Timeout: time.Nanosecond}
					fakeConfig.StagingTimeoutReturns(time.Nanosecond)
				})

				It("returns the error and warnings", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				})
			})

			Context("when the polling errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("I am a banana")
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: ccv3.BuildStateStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
					fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{}, ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				})
			})
		})

		Context("when creation errors", func() {
			var expectedErr error
			BeforeEach(func() {
				expectedErr = errors.New("I am a banana")
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			})
		})
	})
})
