package v3action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Manifest Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ApplyApplicationManifest", func() {
		var (
			fakeParser *v3actionfakes.FakeManifestParser
			spaceGUID  string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeParser = new(v3actionfakes.FakeManifestParser)
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
								Expect(warnings).To(ConsistOf("app-1-warning", "apply-manifest-1-warning", "poll-1-warning"))

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
								Expect(warnings).To(ConsistOf("app-1-warning", "apply-manifest-1-warning", "poll-1-warning"))
							})
						})

						When("polling returns an job failed error", func() {
							var expectedErr error

							BeforeEach(func() {
								expectedErr = ccerror.V3JobFailedError{Detail: "some-job-failed"}
								fakeCloudControllerClient.PollJobReturns(
									ccv3.Warnings{"poll-1-warning"},
									expectedErr,
								)
							})

							It("reports a polling error", func() {
								Expect(executeErr).To(Equal(actionerror.ApplicationManifestError{Message: "some-job-failed"}))
								Expect(warnings).To(ConsistOf("app-1-warning", "apply-manifest-1-warning", "poll-1-warning"))
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
							Expect(warnings).To(ConsistOf("app-1-warning", "apply-manifest-1-warning"))
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
})
