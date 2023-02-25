package v7action_test

import (
	"errors"
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Revisions Actions", func() {
	Describe("GetRevisionsByApplicationNameAndSpace", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			fakeConfig                *v7actionfakes.FakeConfig
			appName                   string
			spaceGUID                 string
			fetchedRevisions          []resources.Revision
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			fakeConfig = new(v7actionfakes.FakeConfig)
			actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil, nil)
			appName = "some-app"
			spaceGUID = "space-guid"
			fakeConfig.APIVersionReturns("3.86.0")
		})

		JustBeforeEach(func() {
			fetchedRevisions, warnings, executeErr = actor.GetRevisionsByApplicationNameAndSpace(appName, spaceGUID)
		})

		When("finding the app fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv3.Warnings{"get-application-warning"},
					errors.New("get-application-error"),
				)
			})

			It("returns an executeError", func() {
				Expect(executeErr).To(MatchError("get-application-error"))
				Expect(warnings).To(ConsistOf("get-application-warning"))
			})
		})

		When("finding the app succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]resources.Application{{Name: "some-app", GUID: "some-app-guid"}},
					ccv3.Warnings{"get-application-warning"},
					nil,
				)
			})

			When("getting the app revisions fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns(
						[]resources.Revision{},
						ccv3.Warnings{"get-application-revisisions-warning"},
						errors.New("get-application-revisions-error"),
					)
				})

				It("returns an executeError", func() {
					Expect(executeErr).To(MatchError("get-application-revisions-error"))
					Expect(warnings).To(ConsistOf("get-application-warning", "get-application-revisisions-warning"))
				})
			})

			When("getting the app revisions succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRevisionsReturns(
						[]resources.Revision{
							{GUID: "3", Deployable: true, Droplet: resources.Droplet{GUID: "droplet-guid-3"}},
							{GUID: "2", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-2"}},
							{GUID: "1", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
						},
						ccv3.Warnings{"get-application-revisisions-warning"},
						nil,
					)
				})

				It("makes the API call to get the app revisions and returns all warnings and revisions in descending order", func() {
					Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1), "GetApplications call count")
					Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).To(ConsistOf(
						ccv3.Query{Key: ccv3.NameFilter, Values: []string{appName}},
						ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{spaceGUID}},
					))

					Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1), "GetApplicationRevisions call count")

					appGuidArg, queryArg := fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)
					Expect(appGuidArg).To(Equal("some-app-guid"))
					Expect(queryArg).To(Equal([]ccv3.Query{{Key: ccv3.OrderBy, Values: []string{"-created_at"}}}))

					Expect(fakeConfig.APIVersionCallCount()).To(Equal(1), "APIVersion call count")

					Expect(fetchedRevisions).To(Equal(
						[]resources.Revision{
							{GUID: "3", Deployable: true, Droplet: resources.Droplet{GUID: "droplet-guid-3"}},
							{GUID: "2", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-2"}},
							{GUID: "1", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
						}))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-warning", "get-application-revisisions-warning"))
				})

				When("minimum supported version for CAPI returning deployable field is not met (< 3.86.0)", func() {
					BeforeEach(func() {
						fakeConfig.APIVersionReturns("3.85.0")

						fakeCloudControllerClient.GetApplicationRevisionsReturns(
							[]resources.Revision{
								{GUID: "3", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-3"}},
								{GUID: "2", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-2"}},
								{GUID: "1", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
							},
							ccv3.Warnings{"get-application-revisisions-warning"},
							nil,
						)

						fakeCloudControllerClient.GetDropletsReturns(
							[]resources.Droplet{
								{GUID: "droplet-guid-2", State: constant.DropletExpired},
								{GUID: "droplet-guid-3", State: constant.DropletStaged},
							},
							nil,
							nil,
						)
					})

					It("fetches the deployments", func() {
						Expect(fakeConfig.APIVersionCallCount()).To(Equal(1), "APIVersion call count")
						Expect(fakeCloudControllerClient.GetDropletsCallCount()).To(Equal(1), "GetDroplets call count")
						Expect(fakeCloudControllerClient.GetDropletsArgsForCall(0)).To(Equal(
							[]ccv3.Query{{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}}},
						))
						Expect(fetchedRevisions).To(Equal(
							[]resources.Revision{
								{GUID: "3", Deployable: true, Droplet: resources.Droplet{GUID: "droplet-guid-3"}},
								{GUID: "2", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-2"}},
								{GUID: "1", Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
							}))
						Expect(executeErr).ToNot(HaveOccurred())
					})

					When("fetching droplets fails", func() {
						BeforeEach(func() {
							fakeCloudControllerClient.GetDropletsReturns(
								[]resources.Droplet{},
								ccv3.Warnings{"get-droplets-warning"},
								errors.New("get-droplets-error"),
							)
						})

						It("returns the fetching droplets error", func() {
							Expect(executeErr).To(MatchError("get-droplets-error"))
							Expect(warnings).To(ConsistOf("get-application-warning", "get-application-revisisions-warning", "get-droplets-warning"))
						})
					})
				})
			})
		})
	})

	Describe("GetRevisionByApplicationAndVersion", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			fakeConfig                *v7actionfakes.FakeConfig
			appGUID                   string
			revisionVersion           int
			executeErr                error
			warnings                  Warnings
			revision                  resources.Revision
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			fakeConfig = new(v7actionfakes.FakeConfig)
			actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil, nil)
			appGUID = "some-app-guid"
			revisionVersion = 1
			fakeConfig.APIVersionReturns("3.86.0")
		})

		JustBeforeEach(func() {
			revision, warnings, executeErr = actor.GetRevisionByApplicationAndVersion(appGUID, revisionVersion)
		})

		When("finding the revision succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsReturns(
					[]resources.Revision{
						{GUID: "revision-guid-1", Version: 1, Deployable: true},
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

				Expect(fakeConfig.APIVersionCallCount()).To(Equal(1), "APIVersion call count")
				Expect(fakeCloudControllerClient.GetDropletsCallCount()).To(Equal(0), "GetDroplets call count")

				Expect(revision.Version).To(Equal(1))
				Expect(revision.GUID).To(Equal("revision-guid-1"))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
			})

			When("minimum supported version for CAPI returning deployable field is not met (< 3.86.0)", func() {
				BeforeEach(func() {
					fakeConfig.APIVersionReturns("3.85.0")

					fakeCloudControllerClient.GetApplicationRevisionsReturns(
						[]resources.Revision{
							{GUID: "revision-guid-1", Version: 1, Deployable: false, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
						},
						ccv3.Warnings{"get-revisions-warning-1"},
						nil,
					)

					fakeCloudControllerClient.GetDropletsReturns(
						[]resources.Droplet{
							{GUID: "droplet-guid-1", State: constant.DropletStaged},
						},
						nil,
						nil,
					)
				})

				It("fills in deployable based on droplet status", func() {
					expectedQuery := ccv3.Query{
						Key:    ccv3.VersionsFilter,
						Values: []string{strconv.Itoa(revisionVersion)},
					}
					Expect(fakeCloudControllerClient.GetApplicationRevisionsCallCount()).To(Equal(1), "GetApplicationRevisions call count")
					appGuid, query := fakeCloudControllerClient.GetApplicationRevisionsArgsForCall(0)
					Expect(appGuid).To(Equal("some-app-guid"))
					Expect(query).To(ContainElement(expectedQuery))

					Expect(fakeConfig.APIVersionCallCount()).To(Equal(1), "APIVersion call count")
					Expect(fakeCloudControllerClient.GetDropletsCallCount()).To(Equal(1), "GetDroplets call count")
					Expect(fakeCloudControllerClient.GetDropletsArgsForCall(0)).To(Equal(
						[]ccv3.Query{{Key: ccv3.AppGUIDFilter, Values: []string{"some-app-guid"}}},
					))

					Expect(revision.Version).To(Equal(1))
					Expect(revision.GUID).To(Equal("revision-guid-1"))
					Expect(revision.Deployable).To(Equal(true))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-revisions-warning-1"))
				})

				When("fetching droplets fails", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetDropletsReturns(
							[]resources.Droplet{},
							ccv3.Warnings{"get-droplets-warning"},
							errors.New("get-droplets-error"),
						)
					})

					It("returns the fetching droplets error", func() {
						Expect(executeErr).To(MatchError("get-droplets-error"))
						Expect(warnings).To(ConsistOf("get-revisions-warning-1", "get-droplets-warning"))
					})
				})
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

	Describe("GetApplicationRevisionsDeployed", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			appGUID                   string
			fetchedRevisions          []resources.Revision
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			actor = NewActor(fakeCloudControllerClient, nil, nil, nil, nil, nil)
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			fetchedRevisions, warnings, executeErr = actor.GetApplicationRevisionsDeployed(appGUID)
		})

		When("getting the app deployed revisions fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsDeployedReturns([]resources.Revision{}, ccv3.Warnings{"some-revisions-warnings"}, errors.New("some-revisions-executeError"))
			})

			It("returns an executeError", func() {
				Expect(executeErr).To(MatchError("some-revisions-executeError"))
				Expect(warnings).To(ConsistOf("some-revisions-warnings"))
			})
		})

		When("getting the app revisions succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRevisionsDeployedReturns(
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
				Expect(fakeCloudControllerClient.GetApplicationRevisionsDeployedCallCount()).To(Equal(1))

				appGuidArg := fakeCloudControllerClient.GetApplicationRevisionsDeployedArgsForCall(0)
				Expect(appGuidArg).To(Equal("some-app-guid"))

				Expect(fetchedRevisions).To(Equal(
					[]resources.Revision{
						{GUID: "3"},
						{GUID: "2"},
						{GUID: "1"},
					}))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-revisions-warnings"))
			})
		})
	})
})
