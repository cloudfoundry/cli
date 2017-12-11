package v3action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
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
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil)
	})

	Describe("StagePackage", func() {
		var (
			dropletStream  <-chan Droplet
			warningsStream <-chan Warnings
			errorStream    <-chan error

			buildGUID   string
			dropletGUID string
		)

		AfterEach(func() {
			Eventually(errorStream).Should(BeClosed())
			Eventually(warningsStream).Should(BeClosed())
			Eventually(dropletStream).Should(BeClosed())
		})

		JustBeforeEach(func() {
			dropletStream, warningsStream, errorStream = actor.StagePackage("some-package-guid", "some-app")
		})

		Context("when the creation is successful", func() {
			BeforeEach(func() {
				buildGUID = "some-build-guid"
				dropletGUID = "some-droplet-guid"
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, nil)
				fakeConfig.StagingTimeoutReturns(time.Minute)
			})

			Context("when the polling is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
					fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{CreatedAt: "some-time", GUID: buildGUID, State: constant.BuildStaged, DropletGUID: "some-droplet-guid"}, ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, nil)
				})

				//TODO: uncommend after #150569020
				// FContext("when looking up the droplet fails", func() {
				// 	BeforeEach(func() {
				// 		fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{}, ccv3.Warnings{"droplet-warnings-1", "droplet-warnings-2"}, errors.New("some-droplet-error"))
				// 	})

				// 	It("returns the warnings and the droplet error", func() {
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("droplet-warnings-1", "droplet-warnings-2")))

				// 		Eventually(errorStream).Should(Receive(MatchError("some-droplet-error")))
				// 	})
				// })

				// Context("when looking up the droplet succeeds", func() {
				// 	BeforeEach(func() {
				// 		fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{GUID: dropletGUID, State: ccv3.DropletStateStaged}, ccv3.Warnings{"droplet-warnings-1", "droplet-warnings-2"}, nil)
				// 	})

				It("polls until build is finished and returns the final droplet", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					// Eventually(warningsStream).Should(Receive(ConsistOf("droplet-warnings-1", "droplet-warnings-2")))

					Eventually(dropletStream).Should(Receive(Equal(Droplet{GUID: dropletGUID, State: constant.DropletStaged, CreatedAt: "some-time"})))
					Consistently(errorStream).ShouldNot(Receive())

					Expect(fakeCloudControllerClient.CreateBuildCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateBuildArgsForCall(0)).To(Equal(ccv3.Build{
						PackageGUID: "some-package-guid",
					}))

					Expect(fakeCloudControllerClient.GetBuildCallCount()).To(Equal(2))
					Expect(fakeCloudControllerClient.GetBuildArgsForCall(0)).To(Equal(buildGUID))
					Expect(fakeCloudControllerClient.GetBuildArgsForCall(1)).To(Equal(buildGUID))

					Expect(fakeConfig.PollingIntervalCallCount()).To(Equal(1))
				})
				// })

				Context("when polling returns a failed build", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetBuildReturnsOnCall(
							1,
							ccv3.Build{
								GUID:  buildGUID,
								State: constant.BuildFailed,
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
						Eventually(dropletStream).ShouldNot(Receive())

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
					expectedErr = actionerror.StagingTimeoutError{AppName: "some-app", Timeout: 0}
					fakeConfig.StagingTimeoutReturns(0)
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
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
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
