package v7action_test

import (
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployment Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("CreateDeployment", func() {
		It("delegates to the cloud controller client", func() {
			fakeCloudControllerClient.CreateApplicationDeploymentReturns("some-deployment-guid", ccv3.Warnings{"create-warning-1", "create-warning-2"}, errors.New("create-error"))

			returnedDeploymentGUID, warnings, executeErr := actor.CreateDeployment("some-app-guid", "some-droplet-guid")

			Expect(fakeCloudControllerClient.CreateApplicationDeploymentCallCount()).To(Equal(1))
			givenAppGUID, givenDropletGUID := fakeCloudControllerClient.CreateApplicationDeploymentArgsForCall(0)

			Expect(givenAppGUID).To(Equal("some-app-guid"))
			Expect(givenDropletGUID).To(Equal("some-droplet-guid"))

			Expect(returnedDeploymentGUID).To(Equal("some-deployment-guid"))
			Expect(warnings).To(Equal(Warnings{"create-warning-1", "create-warning-2"}))
			Expect(executeErr).To(MatchError("create-error"))
		})
	})
})
