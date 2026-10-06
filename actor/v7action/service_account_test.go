package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Account Actions", func() {
	var actor *Actor
	var client *v7actionfakes.FakeCloudControllerClient
	BeforeEach(func() {
		client = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(client, nil, nil, nil, nil, nil)
	})
	It("creates in the specified space and preserves platform state and warnings", func() {
		created := resources.ServiceAccount{GUID: "account-guid", Status: "unprovisioned"}
		client.CreateServiceAccountReturns(created, ccv3.Warnings{"warning"}, nil)
		account, warnings, err := actor.CreateServiceAccountInSpace("shared-worker", "Workers", "space-guid")
		Expect(err).NotTo(HaveOccurred())
		Expect(account).To(Equal(created))
		Expect(warnings).To(ConsistOf("warning"))
		Expect(client.CreateServiceAccountCallCount()).To(Equal(1))
		Expect(client.CreateServiceAccountArgsForCall(0)).To(Equal(resources.ServiceAccount{
			Name: "shared-worker", Description: "Workers", Relationships: resources.Relationships{constant.RelationshipTypeSpace: {GUID: "space-guid"}},
		}))
	})
	It("propagates errors and warnings without adopting an existing account", func() {
		denied := errors.New("name permanently reserved")
		client.CreateServiceAccountReturns(resources.ServiceAccount{}, ccv3.Warnings{"warning"}, denied)
		_, warnings, err := actor.CreateServiceAccountInSpace("shared-worker", "", "space-guid")
		Expect(err).To(MatchError(denied))
		Expect(warnings).To(ConsistOf("warning"))
	})
})
