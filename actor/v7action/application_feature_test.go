package v7action_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Feature Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeConfig                *v7actionfakes.FakeConfig
		fakeClock                 *fakeclock.FakeClock
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		fakeClock = fakeclock.NewFakeClock(time.Now())
		actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, fakeClock)
	})

	Describe("GetAppFeature", func() {
		var (
			appGUID    = "some-app-guid"
			warnings   Warnings
			executeErr error
			appFeature ccv3.ApplicationFeature
		)

		JustBeforeEach(func() {
			appFeature, warnings, executeErr = actor.GetAppFeature(appGUID, "ssh")
		})
		Context("Getting SSH", func() {
			When("it succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetAppFeatureReturns(
						ccv3.ApplicationFeature{Name: "ssh", Enabled: true},
						ccv3.Warnings{},
						nil,
					)
				})

				It("calls ccv3 to check current ssh ability", func() {
					Expect(fakeCloudControllerClient.GetAppFeatureCallCount()).To(Equal(1))
					appGuid, featureName := fakeCloudControllerClient.GetAppFeatureArgsForCall(0)
					Expect(appGuid).To(Equal(appGUID))
					Expect(featureName).To(Equal("ssh"))
				})

				It("returns an app feature", func() {
					Expect(appFeature.Name).To(Equal("ssh"))
					Expect(appFeature.Enabled).To(BeTrue())
				})

				When("desired enabled state is already, the current state", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetAppFeatureReturns(
							ccv3.ApplicationFeature{Enabled: true},
							ccv3.Warnings{"some-ccv3-warning"},
							nil,
						)
					})

					It("returns a waring", func() {
						Expect(warnings).To(ConsistOf("some-ccv3-warning"))
					})
				})
			})

			When("when the API layer call returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetAppFeatureReturns(
						ccv3.ApplicationFeature{Enabled: false},
						ccv3.Warnings{"some-get-ssh-warning"},
						errors.New("some-get-ssh-error"),
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(executeErr).To(MatchError("some-get-ssh-error"))
					Expect(warnings).To(ConsistOf("some-get-ssh-warning"))

					Expect(fakeCloudControllerClient.GetAppFeatureCallCount()).To(Equal(1))
				})
			})
		})

	})

	Describe("UpdateAppFeature", func() {
		var (
			app        = Application{Name: "some-app", GUID: "some-app-guid"}
			enabled    = true
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateAppFeature(app, true, "ssh")
		})
		Context("Getting SSH", func() {
			When("it succeeds", func() {
				It("calls ccv3 to enable ssh", func() {
					Expect(fakeCloudControllerClient.UpdateAppFeatureCallCount()).To(Equal(1))
					actualApp, actualEnabled, featureName := fakeCloudControllerClient.UpdateAppFeatureArgsForCall(0)
					Expect(actualApp).To(Equal(app.GUID))
					Expect(actualEnabled).To(Equal(enabled))
					Expect(featureName).To(Equal("ssh"))
				})

				When("the API layer call is successful", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateAppFeatureReturns(ccv3.Warnings{"some-update-ssh-warning"}, nil)
					})

					It("does not error", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("some-update-ssh-warning"))
					})
				})
			})

			When("when the API layer call returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateAppFeatureReturns(
						ccv3.Warnings{"some-update-ssh-warning"},
						errors.New("some-update-ssh-error"),
					)
				})

				It("returns the error and prints warnings", func() {
					Expect(executeErr).To(MatchError("some-update-ssh-error"))
					Expect(warnings).To(ConsistOf("some-update-ssh-warning"))

					Expect(fakeCloudControllerClient.UpdateAppFeatureCallCount()).To(Equal(1))
				})
			})
		})
	})
})
