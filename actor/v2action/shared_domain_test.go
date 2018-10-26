package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Shared Domain Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
		executeErr                error
		domainName                string
		routerGroup               RouterGroup
		warnings                  Warnings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("CreateSharedDomain", func() {
		BeforeEach(func() {
			domainName = "some-domain-name"
			routerGroup = RouterGroup{
				GUID: "some-guid",
			}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateSharedDomain(domainName, routerGroup)
		})

		It("should call the appropriate method on the client", func() {
			Expect(fakeCloudControllerClient.CreateSharedDomainCallCount()).To(Equal(1))
			domain, routerGrouId := fakeCloudControllerClient.CreateSharedDomainArgsForCall(0)
			Expect(domain).To(Equal(domainName))
			Expect(routerGrouId).To(Equal(routerGroup.GUID))
		})

		When("the call fails", func() {
			var expectedError error
			BeforeEach(func() {
				expectedError = errors.New("something terrible has happened")
				fakeCloudControllerClient.CreateSharedDomainReturns(
					ccv2.Warnings{"some warning", "another warning"},
					expectedError,
				)
			})
			It("should return the appropriate error", func() {
				Expect(executeErr).To(MatchError(expectedError))
			})

			It("should return all warnings", func() {
				Expect(warnings).To(ConsistOf("some warning", "another warning"))
			})
		})

		When("the call succeeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateSharedDomainReturns(
					ccv2.Warnings{"some warning", "another warning"},
					nil,
				)
			})

			It("returns the all warnings and no error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some warning", "another warning"))
			})
		})
	})
})
