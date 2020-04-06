package v7action_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/clock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Build Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeConfig                *v7actionfakes.FakeConfig
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, clock.NewClock())
	})

	Describe("StagePackage", func() {
		var (
			dropletStream  <-chan Droplet
			warningsStream <-chan Warnings
			errorStream    <-chan error

			appName     string
			appGUID     string
			buildGUID   string
			dropletGUID string
			spaceGUID   string
			packageGUID string
		)

		BeforeEach(func() {
			appName = "some-app"
			appGUID = "app-guid"
			spaceGUID = "space-guid"
			packageGUID = "some-package-guid"
		})

		AfterEach(func() {
			Eventually(errorStream).Should(BeClosed())
			Eventually(warningsStream).Should(BeClosed())
			Eventually(dropletStream).Should(BeClosed())
		})

		JustBeforeEach(func() {
			dropletStream, warningsStream, errorStream = actor.StagePackage(packageGUID, appName, spaceGUID)
		})

		When("finding the app fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a tomato")
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-apps-warning"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			})
		})

		When("app is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{}, ccv3.Warnings{"get-apps-warning"}, nil)
			})

			It("returns the error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				Eventually(errorStream).Should(Receive(MatchError(actionerror.ApplicationNotFoundError{Name: appName})))
			})
		})

		When("finding the package fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("I am a passion fruit")
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{GUID: appGUID}}, ccv3.Warnings{"get-apps-warning"}, nil)
				fakeCloudControllerClient.GetPackagesReturns([]ccv3.Package{}, ccv3.Warnings{"get-packages-warning"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			})
		})

		When("the package does not belong to the app", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{GUID: appGUID}}, ccv3.Warnings{"get-apps-warning"}, nil)
				fakeCloudControllerClient.GetPackagesReturns(
					[]ccv3.Package{{GUID: "some-other-package-guid"}},
					ccv3.Warnings{"get-packages-warning"},
					nil,
				)
			})

			It("returns a not found error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
				Eventually(errorStream).Should(Receive(MatchError(actionerror.PackageNotFoundInAppError{GUID: packageGUID, AppName: appName})))
			})
		})

		When("the creation is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{GUID: appGUID}}, ccv3.Warnings{"get-apps-warning"}, nil)
				fakeCloudControllerClient.GetPackagesReturns(
					[]ccv3.Package{{GUID: packageGUID}},
					ccv3.Warnings{"get-packages-warning"},
					nil,
				)

				buildGUID = "some-build-guid"
				dropletGUID = "some-droplet-guid"
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, nil)
				fakeConfig.StagingTimeoutReturns(time.Minute)
			})

			When("the polling is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
					fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{CreatedAt: "some-time", GUID: buildGUID, State: constant.BuildStaged, DropletGUID: "some-droplet-guid"}, ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, nil)
				})

				//TODO: uncommend after #150569020
				// FWhen("looking up the droplet fails", func() {
				// 	BeforeEach(func() {
				// 		fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{}, ccv3.Warnings{"droplet-warnings-1", "droplet-warnings-2"}, errors.New("some-droplet-error"))
				// 	})

				// 	It("returns the warnings and the droplet error", func() {
				//    Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				//    Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
				// 		Eventually(warningsStream).Should(Receive(ConsistOf("droplet-warnings-1", "droplet-warnings-2")))

				// 		Eventually(errorStream).Should(Receive(MatchError("some-droplet-error")))
				// 	})
				// })

				// When("looking up the droplet succeeds", func() {
				// 	BeforeEach(func() {
				// 		fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{GUID: dropletGUID, State: ccv3.DropletStateStaged}, ccv3.Warnings{"droplet-warnings-1", "droplet-warnings-2"}, nil)
				// 	})

				It("polls until build is finished and returns the final droplet", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					// Eventually(warningsStream).Should(Receive(ConsistOf("droplet-warnings-1", "droplet-warnings-2")))

					Eventually(dropletStream).Should(Receive(Equal(Droplet{GUID: dropletGUID, State: constant.DropletStaged, CreatedAt: "some-time"})))
					Consistently(errorStream).ShouldNot(Receive())

					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(Equal([]ccv3.Query{
						{Key: ccv3.NameFilter, Values: []string{appName}},
						{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					}))

					Expect(fakeCloudControllerClient.GetPackagesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetPackagesArgsForCall(0)).To(Equal([]ccv3.Query{
						{Key: ccv3.AppGUIDFilter, Values: []string{appGUID}},
					}))

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

				When("polling returns a failed build", func() {
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
						Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
						Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
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

			When("polling times out", func() {
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

			When("the polling errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("I am a banana")
					fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{GUID: buildGUID, State: constant.BuildStaging}, ccv3.Warnings{"get-warnings-1", "get-warnings-2"}, nil)
					fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{}, ccv3.Warnings{"get-warnings-3", "get-warnings-4"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
					Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-1", "get-warnings-2")))
					Eventually(warningsStream).Should(Receive(ConsistOf("get-warnings-3", "get-warnings-4")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				})
			})
		})

		When("creation errors", func() {
			var expectedErr error

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]ccv3.Application{{GUID: appGUID}}, ccv3.Warnings{"get-apps-warning"}, nil)
				fakeCloudControllerClient.GetPackagesReturns(
					[]ccv3.Package{{GUID: packageGUID}},
					ccv3.Warnings{"get-packages-warning"},
					nil,
				)

				expectedErr = errors.New("I am a banana")
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Eventually(warningsStream).Should(Receive(ConsistOf("get-apps-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("get-packages-warning")))
				Eventually(warningsStream).Should(Receive(ConsistOf("create-warnings-1", "create-warnings-2")))
				Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
			})
		})
	})

	Describe("StageApplicationPackage", func() {
		var (
			build      Build
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			build, warnings, executeErr = actor.StageApplicationPackage("some-package-guid")
		})

		When("the creation is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{GUID: "some-build-guid"}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, nil)
			})

			It("returns the build and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(build).To(Equal(Build{GUID: "some-build-guid"}))
				Expect(warnings).To(ConsistOf("create-warnings-1", "create-warnings-2"))
			})
		})

		When("the creation fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateBuildReturns(ccv3.Build{}, ccv3.Warnings{"create-warnings-1", "create-warnings-2"}, errors.New("blurp"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("blurp"))
				Expect(warnings).To(ConsistOf("create-warnings-1", "create-warnings-2"))
			})
		})
	})

	Describe("PollBuild", func() {
		var (
			droplet    Droplet
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			droplet, warnings, executeErr = actor.PollBuild("some-build-guid", "some-app")
		})

		When("getting the build yields a 'Staged' build", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{State: constant.BuildStaging}, ccv3.Warnings{"some-get-build-warnings"}, nil)
				fakeCloudControllerClient.GetBuildReturnsOnCall(1, ccv3.Build{GUID: "some-build-guid", DropletGUID: "some-droplet-guid", State: constant.BuildStaged}, ccv3.Warnings{"some-get-build-warnings"}, nil)
				fakeConfig.StagingTimeoutReturns(500 * time.Millisecond)
			})

			It("gets the droplet", func() {
				Expect(fakeCloudControllerClient.GetBuildCallCount()).To(Equal(2))

				Expect(fakeCloudControllerClient.GetDropletCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetDropletArgsForCall(0)).To(Equal("some-droplet-guid"))
			})

			When("getting the droplet is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{GUID: "some-droplet-guid", CreatedAt: "some-droplet-time", State: constant.DropletStaged}, ccv3.Warnings{"some-get-droplet-warnings"}, nil)
				})

				It("returns the droplet and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(warnings).To(ConsistOf("some-get-build-warnings", "some-get-build-warnings", "some-get-droplet-warnings"))
					Expect(droplet).To(Equal(Droplet{
						GUID:      "some-droplet-guid",
						CreatedAt: "some-droplet-time",
						State:     constant.DropletStaged,
					}))
				})
			})

			When("getting the droplet fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDropletReturns(ccv3.Droplet{}, ccv3.Warnings{"some-get-droplet-warnings"}, errors.New("no rain"))
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("no rain"))
					Expect(warnings).To(ConsistOf("some-get-build-warnings", "some-get-build-warnings", "some-get-droplet-warnings"))
				})
			})
		})

		When("getting the build yields a 'Failed' build", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{State: constant.BuildFailed, Error: "ded build"}, ccv3.Warnings{"some-get-build-warnings"}, nil)
				fakeConfig.StagingTimeoutReturns(500 * time.Millisecond)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("ded build"))
				Expect(warnings).To(ConsistOf("some-get-build-warnings"))
			})
		})

		When("getting the build fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{}, ccv3.Warnings{"some-get-build-warnings"}, errors.New("some-poll-build-error"))
				fakeConfig.StagingTimeoutReturns(500 * time.Millisecond)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("some-poll-build-error"))
				Expect(warnings).To(ConsistOf("some-get-build-warnings"))
			})
		})

		When("polling the build times out", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildReturnsOnCall(0, ccv3.Build{}, ccv3.Warnings{"some-get-build-warnings"}, nil)
				fakeConfig.StagingTimeoutReturns(500 * time.Millisecond)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.StagingTimeoutError{AppName: "some-app", Timeout: 500 * time.Millisecond}))
				Expect(warnings).To(ConsistOf("some-get-build-warnings"))
			})
		})
	})
})
