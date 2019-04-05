package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Label actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig
		warnings                  Warnings
		executeErr                error
		appName                   string
		spaceGUID                 string
		labels                    map[string]types.NullString
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil)
	})

	Describe("GetApplicationLabels", func() {
		BeforeEach(func() {

			fakeCloudControllerClient.GetApplicationsReturns(
				[]ccv3.Application{
					ccv3.Application{
						GUID: "some-other-guid",
						Metadata: struct {
							Labels map[string]types.NullString `json:"labels,omitempty"`
						}{
							Labels: map[string]types.NullString{
								"some-key": types.NewNullString("some-value"),
							},
						},
					},
				},
				nil,
				nil,
			)
		})

		JustBeforeEach(func() {
			labels, warnings, executeErr = actor.GetApplicationLabels(appName, spaceGUID)
		})

		It("gets the app labels", func() {
			Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
			Expect(labels).To(Equal(map[string]types.NullString{"some-key": types.NewNullString("some-value")}))
			Expect(warnings).To(BeEmpty())
			Expect(executeErr).ToNot(HaveOccurred())
		})

		When("warnings are returned by the API", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)

			})

			It("returns the warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})
		})

		When("there are client errors", func() {
			When("GetApplications fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-apps-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-apps-error"))
				})
			})
		})

	})

	Describe("UpdateApplicationLabelsByApplicationName", func() {

		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateApplicationLabelsByApplicationName(appName, spaceGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateApplicationReturns(
					ccv3.Application{},
					ccv3.Warnings{"set-app-labels-warnings"},
					nil,
				)
			})

			It("sets the app labels", func() {
				Expect(fakeCloudControllerClient.UpdateApplicationCallCount()).To(Equal(1))
				sentApp := fakeCloudControllerClient.UpdateApplicationArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(sentApp.Metadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
			})

			When("there are client errors", func() {
				When("GetApplications fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetApplicationsReturns(
							[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
							ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
							errors.New("get-apps-error"),
						)
					})

					It("returns the error and all warnings", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
						Expect(executeErr).To(MatchError("get-apps-error"))
					})
				})

				When("UpdateApplication fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.UpdateApplicationReturns(
							ccv3.Application{},
							ccv3.Warnings{"set-app-labels-warnings"},
							errors.New("update-application-error"),
						)
					})
					It("returns the error and all warnings", func() {
						Expect(executeErr).To(HaveOccurred())
						Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
						Expect(executeErr).To(MatchError("update-application-error"))
					})
				})
			})
		})
	})
})
