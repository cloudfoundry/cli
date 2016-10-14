package v2actions_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/actors/v2actions/v2actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionsfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionsfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("GetDomainByGUID", func() {
		Context("when the domain exists and is a shared domain", func() {
			var expectedDomain ccv2.Domain

			BeforeEach(func() {
				expectedDomain = ccv2.Domain{
					GUID: "shared-domain-guid",
					Name: "shared-domain",
				}
				fakeCloudControllerClient.GetSharedDomainReturns(expectedDomain, nil, nil)
			})

			It("returns the shared domain", func() {
				domain, _, err := actor.GetDomainByGUID("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain(expectedDomain)))

				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("shared-domain-guid"))
			})
		})

		Context("when the domain exists and is a private domain", func() {
			var expectedDomain ccv2.Domain

			BeforeEach(func() {
				expectedDomain = ccv2.Domain{
					GUID: "private-domain-guid",
					Name: "private-domain",
				}
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, nil, ccv2.ResourceNotFoundError{})
				fakeCloudControllerClient.GetPrivateDomainReturns(expectedDomain, nil, nil)
			})

			It("returns the private domain", func() {
				domain, _, err := actor.GetDomainByGUID("private-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain(expectedDomain)))

				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("private-domain-guid"))
				Expect(fakeCloudControllerClient.GetPrivateDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetPrivateDomainArgsForCall(0)).To(Equal("private-domain-guid"))
			})
		})

		Context("when the domain does not exist", func() {
			var expectedErr DomainNotFoundError

			BeforeEach(func() {
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, nil, ccv2.ResourceNotFoundError{})
				fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, nil, ccv2.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError", func() {
				domain, _, err := actor.GetDomainByGUID("private-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain(ccv2.Domain{})))
			})
		})

		DescribeTable("when there are warnings and errors", func(
			stubGetSharedDomain func(),
			stubGetPrivateDomain func(),
			expectedDomain Domain,
			expectedWarnings Warnings,
			expectingError bool,
			expectedErr error,
		) {
			stubGetSharedDomain()
			stubGetPrivateDomain()
			domain, warnings, err := actor.GetDomainByGUID("some-domain-guid")
			Expect(domain).To(Equal(expectedDomain))
			Expect(warnings).To(ConsistOf(expectedWarnings))
			if expectingError {
				Expect(err).To(MatchError(expectedErr))
			} else {
				Expect(err).To(Not(HaveOccurred()))
			}
		},

			Entry(
				"shared domain warning and error",
				func() {
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, []string{"shared-domain-warning"}, errors.New("shared domain error"))
				},
				func() { fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, nil, nil) },
				Domain{},
				Warnings{"shared-domain-warning"},
				true,
				errors.New("shared domain error"),
			),

			Entry(
				"shared domain warning and resource not found; private domain warning & error",
				func() {
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, []string{"shared-domain-warning"}, ccv2.ResourceNotFoundError{})
				},
				func() {
					fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, []string{"private-domain-warning"}, errors.New("private domain error"))
				},
				Domain{},
				Warnings{"shared-domain-warning", "private-domain-warning"},
				true,
				errors.New("private domain error"),
			),
		)
	})
})
