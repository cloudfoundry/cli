package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v9/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Account Lifecycle Actions", func() {
	var actor *Actor
	var client *v7actionfakes.FakeCloudControllerClient
	BeforeEach(func() {
		client = new(v7actionfakes.FakeCloudControllerClient)
		actor = NewActor(client, nil, nil, nil, nil, nil)
		client.GetServiceAccountsReturns([]resources.ServiceAccount{{GUID: "account-guid", Name: "workers"}}, ccv3.Warnings{"lookup warning"}, nil)
		client.GetServiceAccountReturns(resources.ServiceAccount{GUID: "account-guid", Name: "workers", Status: "ready"}, ccv3.Warnings{"detail warning"}, nil)
		client.GetApplicationsReturns([]resources.Application{{GUID: "app-guid", Name: "app"}}, ccv3.Warnings{"app warning"}, nil)
		client.DeleteServiceAccountReturns("job-url", ccv3.Warnings{"mutation warning"}, nil)
		client.UpdateServiceAccountEnabledReturns("job-url", ccv3.Warnings{"mutation warning"}, nil)
		client.UpdateApplicationServiceAccountReturns("job-url", ccv3.Warnings{"mutation warning"}, nil)
		client.PollJobReturns(ccv3.Warnings{"poll warning"}, nil)
	})
	It("resolves by name AND owning space, then fetches current detail", func() {
		account, warnings, err := actor.GetServiceAccountByNameAndSpace("workers", "space-guid")
		Expect(err).NotTo(HaveOccurred())
		Expect(account.Status).To(Equal("ready"))
		Expect(warnings).To(Equal(Warnings{"lookup warning", "detail warning"}))
		Expect(client.GetServiceAccountsArgsForCall(0)).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{"workers"}}, {Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}}}))
		Expect(client.GetServiceAccountArgsForCall(0)).To(Equal("account-guid"))
	})
	It("returns a helpful missing-account error without fetching or mutating another space", func() {
		client.GetServiceAccountsReturns(nil, ccv3.Warnings{"lookup warning"}, nil)
		_, warnings, err := actor.GetServiceAccountByNameAndSpace("workers", "space-guid")
		Expect(err).To(MatchError("Service account 'workers' not found in the targeted space."))
		Expect(warnings).To(ConsistOf("lookup warning"))
		Expect(client.GetServiceAccountCallCount()).To(BeZero())
	})
	DescribeTable("waits for lifecycle completion and preserves every warning", func(operation string) {
		var warnings Warnings
		var err error
		switch operation {
		case "delete":
			warnings, err = actor.DeleteServiceAccountByNameAndSpace("workers", "space-guid")
		case "enable":
			warnings, err = actor.SetServiceAccountEnabledByNameAndSpace("workers", "space-guid", true)
		case "disable":
			warnings, err = actor.SetServiceAccountEnabledByNameAndSpace("workers", "space-guid", false)
		case "bind":
			warnings, err = actor.BindServiceAccountByNameAndSpace("app", "workers", "space-guid")
		case "unbind":
			warnings, err = actor.UnbindServiceAccountByAppNameAndSpace("app", "space-guid")
		}
		Expect(err).NotTo(HaveOccurred())
		Expect(client.PollJobCallCount()).To(Equal(1))
		Expect(client.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("job-url")))
		Expect(warnings).To(ContainElements("mutation warning", "poll warning"))
		if operation == "bind" || operation == "unbind" {
			Expect(client.GetApplicationsArgsForCall(0)).To(ContainElements(ccv3.Query{Key: ccv3.NameFilter, Values: []string{"app"}}, ccv3.Query{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}}))
			guid, relationship := client.UpdateApplicationServiceAccountArgsForCall(0)
			Expect(guid).To(Equal("app-guid"))
			if operation == "bind" {
				Expect(relationship.GUID).To(Equal("account-guid"))
			} else {
				Expect(relationship.GUID).To(BeEmpty())
				Expect(client.GetServiceAccountsCallCount()).To(BeZero())
			}
		} else if operation == "delete" {
			Expect(client.DeleteServiceAccountArgsForCall(0)).To(Equal("account-guid"))
		} else {
			guid, enabled := client.UpdateServiceAccountEnabledArgsForCall(0)
			Expect(guid).To(Equal("account-guid"))
			Expect(enabled).To(Equal(operation == "enable"))
		}
	}, Entry("delete", "delete"), Entry("enable", "enable"), Entry("disable", "disable"), Entry("bind", "bind"), Entry("unbind", "unbind"))

	It("does not poll synchronous unbinding", func() {
		client.UpdateApplicationServiceAccountReturns("", ccv3.Warnings{"restart warning"}, nil)
		warnings, err := actor.UnbindServiceAccountByAppNameAndSpace("app", "space-guid")
		Expect(err).NotTo(HaveOccurred())
		Expect(warnings).To(ContainElement("restart warning"))
		Expect(client.PollJobCallCount()).To(BeZero())
	})
	It("stops binding when account lookup fails", func() {
		client.GetServiceAccountsReturns(nil, ccv3.Warnings{"lookup warning"}, errors.New("lookup denied"))
		warnings, err := actor.BindServiceAccountByNameAndSpace("app", "workers", "space-guid")
		Expect(err).To(MatchError("lookup denied"))
		Expect(warnings).To(ContainElement("lookup warning"))
		Expect(client.UpdateApplicationServiceAccountCallCount()).To(BeZero())
	})
	It("stops unbinding when app lookup fails", func() {
		client.GetApplicationsReturns(nil, ccv3.Warnings{"app warning"}, errors.New("app denied"))
		warnings, err := actor.UnbindServiceAccountByAppNameAndSpace("app", "space-guid")
		Expect(err).To(MatchError("app denied"))
		Expect(warnings).To(ContainElement("app warning"))
		Expect(client.UpdateApplicationServiceAccountCallCount()).To(BeZero())
	})
	It("stops deletion on detail failure", func() {
		client.GetServiceAccountReturns(resources.ServiceAccount{}, ccv3.Warnings{"detail warning"}, errors.New("detail denied"))
		warnings, err := actor.DeleteServiceAccountByNameAndSpace("workers", "space-guid")
		Expect(err).To(MatchError("detail denied"))
		Expect(warnings).To(ContainElement("detail warning"))
		Expect(client.DeleteServiceAccountCallCount()).To(BeZero())
	})
	It("does not poll a rejected disable operation", func() {
		client.UpdateServiceAccountEnabledReturns("", ccv3.Warnings{"mutation warning"}, errors.New("operation in progress"))
		warnings, err := actor.SetServiceAccountEnabledByNameAndSpace("workers", "space-guid", false)
		Expect(err).To(MatchError("operation in progress"))
		Expect(warnings).To(ContainElement("mutation warning"))
		Expect(client.PollJobCallCount()).To(BeZero())
	})
	It("reports job failure with poll warnings", func() {
		client.PollJobReturns(ccv3.Warnings{"poll warning"}, errors.New("provisioning failed"))
		warnings, err := actor.BindServiceAccountByNameAndSpace("app", "workers", "space-guid")
		Expect(err).To(MatchError("provisioning failed"))
		Expect(warnings).To(ContainElements("app warning", "lookup warning", "detail warning", "mutation warning", "poll warning"))
	})
})
