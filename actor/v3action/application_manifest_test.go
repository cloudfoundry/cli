package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
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

		Context("when given at least one application", func() {
			BeforeEach(func() {
				fakeParser.AppNamesReturns([]string{"app-1"})
			})

			Context("when getting the raw manifest bytes is successful", func() {
				var manifestContent []byte

				BeforeEach(func() {
					manifestContent = []byte("some-manifest-contents")
					fakeParser.RawManifestReturns(manifestContent, nil)
				})

				Context("when the app exists", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv3.Application{{GUID: "app-1-guid"}},
							ccv3.Warnings{"app-1-warning"},
							nil,
						)
					})

					Context("when applying the manifest succeeds", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.CreateApplicationActionsApplyManifestByApplicationReturns(
								"some-job-url",
								ccv3.Warnings{"apply-manifest-1-warning"},
								nil,
							)
						})

						Context("when polling finishes successfully", func() {
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

								Expect(fakeCloudControllerClient.CreateApplicationActionsApplyManifestByApplicationCallCount()).To(Equal(1))
								appManifest, guidInCall := fakeCloudControllerClient.CreateApplicationActionsApplyManifestByApplicationArgsForCall(0)
								Expect(appManifest).To(Equal(manifestContent))
								Expect(guidInCall).To(Equal("app-1-guid"))

								Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
								jobURL := fakeCloudControllerClient.PollJobArgsForCall(0)
								Expect(jobURL).To(Equal("some-job-url"))
							})
						})

						Context("when polling returns an error", func() {
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
					})

					Context("when applying the manifest errors", func() {
						var applyErr error

						BeforeEach(func() {
							applyErr = errors.New("some-apply-manifest-error")
							fakeCloudControllerClient.CreateApplicationActionsApplyManifestByApplicationReturns(
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

				Context("when there's an error retrieving the application", func() {
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
						Expect(warnings).To(Equal(Warnings{"app-1-warning"}))
					})
				})
			})

			Context("when generating the raw manifest errors", func() {
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
