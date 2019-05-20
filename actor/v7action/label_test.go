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

var _ = Describe("Labels", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig
		warnings                  Warnings
		executeErr                error
		resourceName              string
		spaceGUID                 string
		orgGUID                   string
		labels                    map[string]types.NullString
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		fakeConfig = new(v7actionfakes.FakeConfig)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil)
	})

	Context("UpdateApplicationLabelsByApplicationName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateApplicationLabelsByApplicationName(resourceName, spaceGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					ccv3.ResourceMetadata{},
					ccv3.Warnings{"set-app-labels-warnings"},
					nil,
				)
			})

			It("sets the app labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, appGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("app"))
				Expect(appGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-app-labels-warnings"))
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

			When("UpdateApplication fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationsReturns(
						[]ccv3.Application{ccv3.Application{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						ccv3.ResourceMetadata{},
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

	Context("UpdateOrganizationLabelsByOrganizationName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateOrganizationLabelsByOrganizationName(resourceName, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]ccv3.Organization{ccv3.Organization{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					ccv3.ResourceMetadata{},
					ccv3.Warnings{"set-org"},
					nil,
				)
			})

			It("sets the org labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, orgGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("org"))
				Expect(orgGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-org"))
			})
		})

		When("there are client errors", func() {
			When("fetching the organization fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{ccv3.Organization{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-orgs-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-orgs-error"))
				})
			})

			When("updating the organization fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{ccv3.Organization{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						ccv3.ResourceMetadata{},
						ccv3.Warnings{"set-org"},
						errors.New("update-orgs-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-org"))
					Expect(executeErr).To(MatchError("update-orgs-error"))
				})
			})
		})
	})

	Context("UpdateSpaceLabelsBySpaceName", func() {
		JustBeforeEach(func() {
			warnings, executeErr = actor.UpdateSpaceLabelsBySpaceName(resourceName, orgGUID, labels)
		})

		When("there are no client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{ccv3.Space{GUID: "some-guid"}},
					ccv3.Warnings([]string{"warning-1", "warning-2"}),
					nil,
				)
				fakeCloudControllerClient.UpdateResourceMetadataReturns(
					ccv3.ResourceMetadata{},
					ccv3.Warnings{"set-space-metadata"},
					nil,
				)
			})

			It("sets the space labels", func() {
				Expect(fakeCloudControllerClient.UpdateResourceMetadataCallCount()).To(Equal(1))
				resourceType, spaceGUID, sentMetadata := fakeCloudControllerClient.UpdateResourceMetadataArgsForCall(0)
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(resourceType).To(BeEquivalentTo("space"))
				Expect(spaceGUID).To(BeEquivalentTo("some-guid"))
				Expect(sentMetadata.Labels).To(BeEquivalentTo(labels))
			})

			It("aggregates warnings", func() {
				Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-space-metadata"))
			})
		})

		When("there are client errors", func() {
			When("fetching the space fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{ccv3.Space{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-failure-1", "warning-failure-2"}),
						errors.New("get-spaces-error"),
					)
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-failure-1", "warning-failure-2"))
					Expect(executeErr).To(MatchError("get-spaces-error"))
				})
			})

			When("updating the space fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{ccv3.Space{GUID: "some-guid"}},
						ccv3.Warnings([]string{"warning-1", "warning-2"}),
						nil,
					)
					fakeCloudControllerClient.UpdateResourceMetadataReturns(
						ccv3.ResourceMetadata{},
						ccv3.Warnings{"set-space"},
						errors.New("update-space-error"),
					)
				})
				It("returns the error and all warnings", func() {
					Expect(executeErr).To(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "set-space"))
					Expect(executeErr).To(MatchError("update-space-error"))
				})
			})
		})
	})
})
