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
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
	})

	Describe("SetSpaceManifest", func() {
		var (
			spaceGUID   string
			rawManifest []byte

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			spaceGUID = "some-space-guid"
			rawManifest = []byte("---\n- applications:\n name: my-app")
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.SetSpaceManifest(spaceGUID, rawManifest)
		})

		When("applying the manifest succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateSpaceApplyManifestReturns(
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
					Expect(warnings).To(ConsistOf("apply-manifest-1-warning", "poll-1-warning"))

					Expect(fakeCloudControllerClient.UpdateSpaceApplyManifestCallCount()).To(Equal(1))
					guidInCall, appManifest := fakeCloudControllerClient.UpdateSpaceApplyManifestArgsForCall(0)
					Expect(guidInCall).To(Equal("some-space-guid"))
					Expect(appManifest).To(Equal(rawManifest))

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
					Expect(warnings).To(ConsistOf("apply-manifest-1-warning", "poll-1-warning"))
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
					Expect(warnings).To(ConsistOf("apply-manifest-1-warning", "poll-1-warning"))
				})
			})
		})

		When("applying the manifest errors", func() {
			var applyErr error

			BeforeEach(func() {
				applyErr = errors.New("some-apply-manifest-error")
				fakeCloudControllerClient.UpdateSpaceApplyManifestReturns(
					"",
					ccv3.Warnings{"apply-manifest-1-warning"},
					applyErr,
				)
			})

			It("reports a error trying to apply the manifest", func() {
				Expect(executeErr).To(Equal(applyErr))
				Expect(warnings).To(ConsistOf("apply-manifest-1-warning"))
			})
		})
	})

	Describe("GetSpaceManifestDiff", func() {
		var (
			spaceGUID        string
			newManifest      []byte
			manifestDiff     []byte
			expectedManifest []byte

			returnedManifest string
			warnings         Warnings
			executeErr       error
		)

		BeforeEach(func() {
			spaceGUID = "some-space-guid"
			newManifest = []byte("---\n- applications:\n name: my-app")
			manifestDiff = []byte(`[{"op": "replace", "path": "/applications/2/processes/1/memory", "was": "256M", "value": "512M"}]`)
			expectedManifest = []byte("---\n- applications:\n name: my-app")
		})

		JustBeforeEach(func() {
			returnedManifest, warnings, executeErr = actor.GetSpaceManifestDiff(spaceGUID, newManifest)
		})

		When("getting the manifest diff succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceManifestDiffReturns(
					manifestDiff,
					ccv3.Warnings{"get-manifest-diff-warning"},
					nil,
				)
			})

			It("returns the manifest diff and warnings", func() {
				Expect(returnedManifest).To(Equal(expectedManifest))
				Expect(warnings).To(ConsistOf("get-manifest-diff-warning"))
				Expect(executeErr).NotTo(HaveOccurred())
			})
		})

		When("getting the manifest diff errors", func() {
			var getDiffErr error

			BeforeEach(func() {
				getDiffErr = errors.New("some-get-manifest-diff-err")
				fakeCloudControllerClient.GetSpaceManifestDiffReturns(
					[]byte{},
					ccv3.Warnings{"get-manifest-diff-warning"},
					getDiffErr,
				)
			})

			It("reports a error trying to apply the manifest", func() {
				Expect(warnings).To(ConsistOf("get-manifest-diff-warning"))
				Expect(executeErr).To(Equal(getDiffErr))
			})
		})

	})
})
