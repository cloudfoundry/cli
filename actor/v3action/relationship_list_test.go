package v3action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Relationship List Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v3actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v3actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil, nil)
	})

	Describe("ShareServiceInstanceToSpaces", func() {
		var (
			serviceInstanceGUID string
			shareToSpaceGUID    string

			relationshipList RelationshipList
			warnings         Warnings
			shareErr         error
		)

		BeforeEach(func() {
			serviceInstanceGUID = "some-service-instance-guid"
			shareToSpaceGUID = "some-space-guid"
		})

		JustBeforeEach(func() {
			relationshipList, warnings, shareErr = actor.ShareServiceInstanceToSpaces(
				serviceInstanceGUID,
				[]string{shareToSpaceGUID})
		})

		Context("when no errors occur sharing the service instance", func() {
			var returnedRelationshipList ccv3.RelationshipList

			BeforeEach(func() {
				returnedRelationshipList = ccv3.RelationshipList{
					GUIDs: []string{"some-space-guid"},
				}
				fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(
					returnedRelationshipList,
					ccv3.Warnings{"share-service-instance-warning"},
					nil)
			})

			It("does not return an error and returns warnings", func() {
				Expect(shareErr).ToNot(HaveOccurred())
				Expect(relationshipList).To(Equal(RelationshipList(returnedRelationshipList)))
				Expect(warnings).To(ConsistOf("share-service-instance-warning"))

				Expect(fakeCloudControllerClient.ShareServiceInstanceToSpacesCallCount()).To(Equal(1))
				serviceInstanceGUIDArg, spaceGUIDsArg := fakeCloudControllerClient.ShareServiceInstanceToSpacesArgsForCall(0)
				Expect(serviceInstanceGUIDArg).To(Equal(serviceInstanceGUID))
				Expect(spaceGUIDsArg).To(Equal([]string{shareToSpaceGUID}))
			})
		})

		Context("when an error occurs sharing the service instance", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("share service instance error")
				fakeCloudControllerClient.ShareServiceInstanceToSpacesReturns(
					ccv3.RelationshipList{},
					ccv3.Warnings{"share-service-instance-warning"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(shareErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("share-service-instance-warning"))
			})
		})
	})
})
