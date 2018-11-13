package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Manifest Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ApplyApplicationManifest", func() {
		var (
			fakeParser *v7actionfakes.FakeManifestParser
			spaceGUID  string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeParser = new(v7actionfakes.FakeManifestParser)
			spaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.ApplyApplicationManifest(fakeParser, spaceGUID)
		})

		When("given at least one application", func() {
			BeforeEach(func() {
				fakeParser.AppNamesReturns([]string{"app-1"})
			})

			When("getting the raw manifest bytes is successful", func() {
				var manifestContent []byte

				BeforeEach(func() {
					manifestContent = []byte("some-manifest-contents")
					fakeParser.RawManifestReturns(manifestContent, nil)
				})

				When("the app exists", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv3.Application{{GUID: "app-1-guid"}},
							ccv3.Warnings{"app-1-warning"},
							nil,
						)
					})

					When("applying the manifest succeeds", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.UpdateApplicationApplyManifestReturns(
								"some-job-url",
								ccv3.Warnings{"apply-manifest-1-warning"},
								nil,
							)
						})

						When("polling finishes successfully", func() {
							BeforeEach(func() {
								fakeCloudControllerClient.PollJobReturns(
									ccv3.Warnings{"poll-1-warning"},
									nil,
								)
							})

							It("uploads the app manifest", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(Equal(Warnings{"app-1-warning", "apply-manifest-1-warning", "poll-1-warning"}))

								Expect(fakeParser.RawManifestCallCount()).To(Equal(1))
								appName := fakeParser.RawManifestArgsForCall(0)
								Expect(appName).To(Equal("app-1"))

								Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
								queries := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
								Expect(queries).To(ConsistOf(
									ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
									ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
								))

								Expect(fakeCloudControllerClient.UpdateApplicationApplyManifestCallCount()).To(Equal(1))
								guidInCall, appManifest := fakeCloudControllerClient.UpdateApplicationApplyManifestArgsForCall(0)
								Expect(guidInCall).To(Equal("app-1-guid"))
								Expect(appManifest).To(Equal(manifestContent))

								Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
								jobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
								Expect(jobURL).To(Equal(ccv3.JobURL("some-job-url")))
							})
						})

						When("polling returns a generic error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = errors.New("poll-1-error")
								fakeCloudControllerClient.PollJobReturns(
									ccv3.Warnings{"poll-1-warning"},
									expectedErr,
								)
							})

							It("reports a polling error", func() {
								Expect(executeErr).To(Equal(expectedErr))
								Expect(warnings).To(Equal(Warnings{"app-1-warning", "apply-manifest-1-warning", "poll-1-warning"}))
							})
						})

						When("polling returns an job failed error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = ccerror.JobFailedError{Message: "some-job-failed"}
								fakeCloudControllerClient.PollJobReturns(
									ccv3.Warnings{"poll-1-warning"},
									expectedErr,
								)
							})

							It("reports a polling error", func() {
								Expect(executeErr).To(Equal(actionerror.ApplicationManifestError{Message: "some-job-failed"}))
								Expect(warnings).To(Equal(Warnings{"app-1-warning", "apply-manifest-1-warning", "poll-1-warning"}))
							})
						})
					})

					When("applying the manifest errors", func() {
						var applyErr error

						BeforeEach(func() {
							applyErr = errors.New("some-apply-manifest-error")
							fakeCloudControllerClient.UpdateApplicationApplyManifestReturns(
								"",
								ccv3.Warnings{"apply-manifest-1-warning"},
								applyErr,
							)
						})

						It("reports a error trying to apply the manifest", func() {
							Expect(executeErr).To(Equal(applyErr))
							Expect(warnings).To(Equal(Warnings{"app-1-warning", "apply-manifest-1-warning"}))
						})
					})
				})

				When("there's an error retrieving the application", func() {
					var getAppErr error

					BeforeEach(func() {
						getAppErr = errors.New("get-application-error")

						fakeCloudControllerClient.GetApplicationsReturns(
							nil,
							ccv3.Warnings{"app-1-warning"},
							getAppErr,
						)
					})

					It("returns error and warnings", func() {
						Expect(executeErr).To(Equal(getAppErr))
						Expect(warnings).To(ConsistOf("app-1-warning"))
					})
				})
			})

			When("generating the raw manifest errors", func() {
				getManifestErr := errors.New("get-manifest-error")
				BeforeEach(func() {
					fakeParser.RawManifestReturns(nil, getManifestErr)
				})

				It("returns error", func() {
					Expect(executeErr).To(Equal(getManifestErr))
					Expect(warnings).To(BeEmpty())
				})
			})
		})

		Context("no applications", func() {
			It("does nothing", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())
			})
		})
	})

	Describe("GetRawApplicationManifestByNameAndSpace", func() {
		var (
			appName   string
			spaceGUID string

			manifestBytes []byte
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			appName = "some-app-name"
			spaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			manifestBytes, warnings, executeErr = actor.GetRawApplicationManifestByNameAndSpace(appName, spaceGUID)
		})

		When("getting the application is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{
						{Name: appName, GUID: "some-app-guid"},
					},
					ccv3.Warnings{"get-application-warning"},
					nil,
				)
			})

			It("gets the application from the Cloud Controller", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ContainElement("get-application-warning"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
					ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
					ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
				))
			})

			When("getting the manifest is successful", func() {
				var rawManifest []byte

				BeforeEach(func() {
					rawManifest = []byte("---\n- potato")
					fakeCloudControllerClient.GetApplicationManifestReturns(
						rawManifest,
						ccv3.Warnings{"get-manifest-warnings"},
						nil,
					)
				})

				It("returns the raw manifest bytes and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-warning", "get-manifest-warnings"))
					Expect(manifestBytes).To(Equal(rawManifest))
				})

				It("gets the manifest for the application", func() {
					Expect(fakeCloudControllerClient.GetApplicationManifestCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationManifestArgsForCall(0)).To(Equal("some-app-guid"))
				})
			})

			When("getting the manifest returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("some error")
					fakeCloudControllerClient.GetApplicationManifestReturns(
						nil,
						ccv3.Warnings{"get-manifest-warnings"},
						expectedErr,
					)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(Equal(expectedErr))
					Expect(warnings).To(ConsistOf("get-application-warning", "get-manifest-warnings"))
				})
			})
		})

		When("getting the application returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv3.Warnings{"get-application-warning"},
					expectedErr,
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(Equal(expectedErr))
				Expect(warnings).To(ConsistOf("get-application-warning"))
				_ = manifestBytes //TODO DELETE ME
			})
		})
	})
})
