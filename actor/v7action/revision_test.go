package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Revision Actions", func() {
	Describe("GetRevisionByGUID", func() {
		var (
			actor                     *Actor
			fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
			fakeConfig                *v7actionfakes.FakeConfig
			revisionGUID              string
			fetchedRevision           resources.Revision
			executeErr                error
			warnings                  Warnings
		)

		BeforeEach(func() {
			fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
			fakeConfig = new(v7actionfakes.FakeConfig)
			actor = NewActor(fakeCloudControllerClient, fakeConfig, nil, nil, nil, nil)
			revisionGUID = "revision-guid"
			fakeConfig.APIVersionReturns("3.86.0")
		})

		JustBeforeEach(func() {
			fetchedRevision, warnings, executeErr = actor.GetRevisionByGUID(revisionGUID)
		})

		When("finding the revision fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRevisionReturns(
					resources.Revision{},
					ccv3.Warnings{"get-revision-warning"},
					errors.New("get-revision-error"),
				)
			})

			It("returns an executeError", func() {
				Expect(executeErr).To(MatchError("get-revision-error"))
				Expect(warnings).To(ConsistOf("get-revision-warning"))
			})
		})

		When("finding the revision succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRevisionReturns(
					resources.Revision{GUID: "1", Deployable: true, Droplet: resources.Droplet{GUID: "droplet-guid-1"}},
					nil,
					nil,
				)
			})

			It("returns the revision", func() {
				Expect(fakeCloudControllerClient.GetRevisionCallCount()).To(Equal(1), "GetRevision call count")
				Expect(fakeCloudControllerClient.GetRevisionArgsForCall(0)).To(Equal(revisionGUID))

				Expect(fetchedRevision).To(Equal(
					resources.Revision{GUID: "1", Deployable: true, Droplet: resources.Droplet{GUID: "droplet-guid-1"}}))
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})
	})
})
