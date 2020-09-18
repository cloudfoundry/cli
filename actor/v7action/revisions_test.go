package v7action_test

import (
	"errors"
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
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
				fakeCloudControllerClient.GetApplicationsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-executeError"))
			})

			It("returns an executeError", func() {
				Expect(executeErr).To(MatchError("get-application-executeError"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns([]resources.Application{{Name: "some-app", GUID: "some-app-guid"}}, ccv3.Warnings{"get-application-warning"}, nil)
			})

			When("getting the app revisions fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns([]resources.Revision{}, ccv3.Warnings{"some-revisions-warnings"}, errors.New("some-revisions-executeError"))
				})

				It("returns an executeError", func() {
					Expect(executeErr).To(MatchError("some-revisions-executeError"))
					Expect(warnings).To(ConsistOf("get-application-warning", "some-revisions-warnings"))
				})
			})

			When("getting the app revisions succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns(
						[]resources.Revision{
							{GUID: "3"},
							{GUID: "2"},
							{GUID: "1"},
						},
						ccv3.Warnings{"some-revisions-warnings"},
						nil,
					)
				})

				It("makes the API call to get the app revisions and returns all warnings and revisions in descending order", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					))

					Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1))

					appGuidArg, queryArg := fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)
					Expect(appGuidArg).To(Equal("some-app-guid"))
					Expect(queryArg).To(Equal([]ccv3.Query{{Key: ccv3.OrderBy, Values: []string{"-created_at"}}}))

					Expect(fetchedRevisions).To(Equal(
						[]resources.Revision{
							{GUID: "3"},
							{GUID: "2"},
							{GUID: "1"},
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
			executeErr                error
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
			revision, warnings, executeErr = actor.GetRevisionByApplicationAndVersion(appGUID, revisionVersion)
		})

		When("finding the revision succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(
					[]resources.Revision{
						{GUID: "revision-guid-1", Version: 1},
					},
					ccv3.Warnings{"get-revisions-warning-1"},
					nil,
				)
			})

			It("returns the revision", func() {
				expectedQuery := ccv3.Query{
					Key:    ccv3.VersionsFilter,
					Values: []string{strconv.Itoa(revisionVersion)},
				}
				Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1), "GetApplicationRevisions call count")
				appGuid, query := fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)
				Expect(appGuid).To(Equal("some-app-guid"))
				Expect(query).To(ContainElement(expectedQuery))

				Expect(revision.Version).To(Equal(1))
				Expect(revision.GUID).To(Equal("revision-guid-1"))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
			})
		})
		When("no matching revision found", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(
					[]resources.Revision{},
					ccv3.Warnings{"get-revisions-warning-1"},
					nil,
				)
			})
			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.RevisionNotFoundError{Version: 1}))
				Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
			})
		})
		When("more than one revision found", func() {

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(
					[]resources.Revision{
						{GUID: "revision-guid-1", Version: 1},
						{GUID: "revision-guid-2", Version: 22},
					},
					ccv3.Warnings{"get-revisions-warning-1"},
					nil,
				)
			})
			It("returns an error", func() {
				Expect(executeErr).To(MatchError(actionerror.RevisionAmbiguousError{Version: 1}))
				Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
			})
		})
		When("finding the revision fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(nil, ccv3.Warnings{"get-application-warning"}, errors.New("get-application-error"))
			})

			It("returns an error", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})
	})
})
