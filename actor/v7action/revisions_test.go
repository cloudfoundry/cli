package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Revisions Actions", func() {

	Describe("GetRevisionsByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appName                   string
			spaceGUID                 string
			fetchedRevisions          []resources.Revision
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
		})

		JustBeforeEach(func() {
			fetchedRevisions, warnings, executeErr = actor.GetRevisionsByApplicationNameAndSpace(appName, spaceGUID)
		})

		When("finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			When("getting the app revisions fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns([]resources.Revision{}, ccv3.Warnings{"some-revisions-warnings"}, errors.New("some-revisions-error"))
				})
				It("returns an error", func() {
					Expect(executeErr).To(MatchError("some-revisions-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-revisions-warnings"))
				})
			})

			When("getting the app revisions succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns(
						[]resources.Revision{
							{GUID: "1"},
							{GUID: "2"},
							{GUID: "3"},
						},
						ccv3.Warnings{"some-revisions-warnings"},
						nil,
					)
				})

				It("makes the API call to get the app revisions and returns all warnings", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					))

					Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)).To(Equal("some-app-guid"))

					Expect(fetchedRevisions).To(Equal(
						[]resources.Revision{
							{GUID: "1"},
							{GUID: "2"},
							{GUID: "3"},
						}))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-warning", "some-revisions-warnings"))
				})
			})
		})
	})

	Describe("GetRevisionByApplicationAndVersion", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appGUID                   string
			revisionVersion           int
			err                       error
			warnings                  Warnings
			revision                  resources.Revision
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
			appGUID = "some-app-guid"
			revisionVersion = 1
		})

		JustBeforeEach(func() {
			revision, warnings, err = actor.GetRevisionByApplicationAndVersion(appGUID, revisionVersion)
		})

		When("finding the revision succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(
					[]resources.Revision{
						{GUID: "revision-guid-1", Version: 1},
						{GUID: "revision-guid-2", Version: 2},
					},
					ccv3.Warnings{"get-revisions-warning-1"},
					nil,
				)
			})

			It("returns the revision", func() {
				Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1), "GetApplicationRevisions call count")
				Expect(fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)).To(Equal("some-app-guid"))

				Expect(revision.Version).To(Equal(1))
				Expect(revision.GUID).To(Equal("revision-guid-1"))
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
			})
		})
	})
})
