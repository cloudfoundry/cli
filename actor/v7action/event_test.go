package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/resources"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("GetRecentEventsForApp", func() {
		var (
			appName   string
			spaceGUID string
			events    []Event
			warnings  Warnings
			err       error
		)

		BeforeEach(func() {
			appName = "some-app-name"
			spaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			events, warnings, err = actor.GetRecentEventsByApplicationNameAndSpace(appName, spaceGUID)
		})

		When("getting the app succeeds", func() {
			BeforeEach(func() {
				apps := []resources.Application{
					{GUID: "some-app-guid"},
				}

				ccWarnings := ccv3.Warnings{
					"some-app-warnings",
				}

				fakeCloudControllerClient.GetApplicationsReturns(
					apps,
					ccWarnings,
					nil,
				)
			})

			When("the cc client returns the list of events", func() {
				BeforeEach(func() {
					ccEvents := []ccv3.Event{
						{GUID: "event-1", Type: "audit.app.wow", Data: map[string]interface{}{"index": "17"}},
						{GUID: "event-2", Type: "audit.app.cool", Data: map[string]interface{}{"unimportant_key": "23"}},
					}

					ccWarnings := ccv3.Warnings{
						"some-event-warnings",
					}

					fakeCloudControllerClient.GetEventsReturns(
						ccEvents,
						ccWarnings,
						nil,
					)
				})

				It("returns the events and warnings", func() {
					Expect(events).To(Equal([]Event{
						{GUID: "event-1", Type: "audit.app.wow", Description: "index: 17"},
						{GUID: "event-2", Type: "audit.app.cool", Description: ""},
					}))
					Expect(warnings).To(ConsistOf("some-app-warnings", "some-event-warnings"))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			When("the cc client returns an error", func() {
				BeforeEach(func() {
					ccWarnings := ccv3.Warnings{
						"some-event-warnings",
					}

					fakeCloudControllerClient.GetEventsReturns(
						nil,
						ccWarnings,
						errors.New("failed to get events"),
					)
				})

				It("returns the err and warnings", func() {
					Expect(warnings).To(ConsistOf("some-app-warnings", "some-event-warnings"))
					Expect(err).To(MatchError("failed to get events"))
				})
			})
		})

		When("getting the app fails", func() {
			BeforeEach(func() {
				ccWarnings := ccv3.Warnings{
					"some-app-warnings",
				}

				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccWarnings,
					errors.New("failed to get app"),
				)
			})

			It("returns the err and warnings", func() {
				Expect(warnings).To(ConsistOf("some-app-warnings"))
				Expect(err).To(MatchError("failed to get app"))
			})
		})
	})
})
