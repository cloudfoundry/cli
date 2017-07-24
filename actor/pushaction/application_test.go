package pushaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Applications", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)
	})

	Describe("CreateOrUpdateApp", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			event          Event
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				DesiredApplication: Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					},
				},
				Path: "some-path",
			}
		})

		JustBeforeEach(func() {
			returnedConfig, event, warnings, executeErr = actor.CreateOrUpdateApp(config)
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				config.CurrentApplication = Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: "java",
					},
				}
				config.DesiredApplication = Application{
					Application: v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: "ruby",
					},
				}
			})

			Context("when the update is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.UpdateApplicationReturns(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: "ruby",
					}, v2action.Warnings{"update-warning"}, nil)
				})

				It("updates the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("update-warning"))
					Expect(event).To(Equal(UpdatedApplication))

					Expect(returnedConfig.DesiredApplication).To(Equal(Application{
						Application: v2action.Application{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
							Buildpack: "ruby",
						}}))
					Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

					Expect(fakeV2Actor.UpdateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.UpdateApplicationArgsForCall(0)).To(Equal(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: "ruby",
					}))

				})
			})

			Context("when the update errors", func() {
				var expectedErr error
				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.UpdateApplicationReturns(v2action.Application{}, v2action.Warnings{"update-warning"}, expectedErr)
				})

				It("returns warnings and error and stops", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("update-warning"))
				})
			})
		})

		Context("when the app does not exist", func() {
			Context("when the creation is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.CreateApplicationReturns(v2action.Application{
						Name:      "some-app-name",
						GUID:      "some-app-guid",
						SpaceGUID: "some-space-guid",
						Buildpack: "ruby",
					}, v2action.Warnings{"create-warning"}, nil)
				})

				It("creates the application", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("create-warning"))
					Expect(event).To(Equal(CreatedApplication))

					Expect(returnedConfig.DesiredApplication).To(Equal(Application{
						Application: v2action.Application{
							Name:      "some-app-name",
							GUID:      "some-app-guid",
							SpaceGUID: "some-space-guid",
							Buildpack: "ruby",
						}}))
					Expect(returnedConfig.CurrentApplication).To(Equal(returnedConfig.DesiredApplication))

					Expect(fakeV2Actor.CreateApplicationCallCount()).To(Equal(1))
					Expect(fakeV2Actor.CreateApplicationArgsForCall(0)).To(Equal(v2action.Application{
						Name:      "some-app-name",
						SpaceGUID: "some-space-guid",
					}))
				})
			})

			Context("when the creation errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.CreateApplicationReturns(v2action.Application{}, v2action.Warnings{"create-warning"}, expectedErr)
				})

				It("sends the warnings and errors and returns true", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("create-warning"))
				})
			})
		})
	})
})
