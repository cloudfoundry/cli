package pushaction_test

import (
	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/manifest"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application Config", func() {
	var (
		actor  *Actor
		v2fake *pushactionfakes.FakeV2Actor
	)

	Describe("ConvertToApplicationConfig", func() {
		var (
			spaceGUID    string
			manifestApps []manifest.Application

			configs    []ApplicationConfig
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			v2fake = new(pushactionfakes.FakeV2Actor)
			actor = NewActor(v2fake)

			spaceGUID = "some-space-guid"
			manifestApps = []manifest.Application{{
				Name: "some-app",
				Path: "some-path",
			}}
		})

		JustBeforeEach(func() {
			configs, warnings, executeErr = actor.ConvertToApplicationConfig(spaceGUID, manifestApps)
		})

		Context("when the application exists", func() {
			var app v2action.Application

			BeforeEach(func() {
				app = v2action.Application{
					Name: "some-app",
					GUID: "some-app-guid",
				}

				v2fake.GetApplicationByNameAndSpaceReturns(app, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, nil)
			})

			It("sets the current and desired application to the current", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
				Expect(configs).To(Equal([]ApplicationConfig{{
					CurrentApplication: app,
					DesiredApplication: app,
					TargetedSpaceGUID:  spaceGUID,
					Path:               "some-path",
				}}))
			})
		})

		Describe("when the application does not exist", func() {
			BeforeEach(func() {
				v2fake.GetApplicationByNameAndSpaceReturns(v2action.Application{}, v2action.Warnings{"some-app-warning-1", "some-app-warning-2"}, v2action.ApplicationNotFoundError{})
			})

			It("creates a new application and sets it to the desired application", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-app-warning-1", "some-app-warning-2"))
				Expect(configs).To(Equal([]ApplicationConfig{{
					DesiredApplication: v2action.Application{
						Name: "some-app",
					},
					TargetedSpaceGUID: spaceGUID,
					Path:              "some-path",
				}}))
			})
		})
	})
})
