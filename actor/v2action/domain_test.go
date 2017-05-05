package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetDomain", func() {
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
				domain, _, err := actor.GetDomain("shared-domain-guid")
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
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, nil, ccerror.ResourceNotFoundError{})
				fakeCloudControllerClient.GetPrivateDomainReturns(expectedDomain, nil, nil)
			})

			It("returns the private domain", func() {
				domain, _, err := actor.GetDomain("private-domain-guid")
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
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, nil, ccerror.ResourceNotFoundError{})
				fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, nil, ccerror.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError", func() {
				domain, _, err := actor.GetDomain("private-domain-guid")
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
			domain, warnings, err := actor.GetDomain("some-domain-guid")
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
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, []string{"shared-domain-warning"}, ccerror.ResourceNotFoundError{})
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

	Describe("GetSharedDomain", func() {
		Context("when the shared domain exists", func() {
			var expectedDomain ccv2.Domain

			BeforeEach(func() {
				expectedDomain = ccv2.Domain{
					GUID: "shared-domain-guid",
					Name: "shared-domain",
				}
				fakeCloudControllerClient.GetSharedDomainReturns(expectedDomain, ccv2.Warnings{"shared domain warning"}, nil)
			})

			It("returns the shared domain and all warnings", func() {
				domain, warnings, err := actor.GetSharedDomain("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain(expectedDomain)))
				Expect(warnings).To(ConsistOf("shared domain warning"))
			})
		})

		Context("when the API returns a not found error", func() {
			var expectedErr DomainNotFoundError

			BeforeEach(func() {
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"shared domain warning"}, ccerror.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError and all warnings", func() {
				domain, warnings, err := actor.GetSharedDomain("shared-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("shared domain warning"))
			})
		})

		Context("when the API returns any other error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("shared domain error")
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"shared domain warning"}, expectedErr)
			})

			It("returns the same error and all warnings", func() {
				domain, warnings, err := actor.GetSharedDomain("shared-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("shared domain warning"))
			})
		})
	})

	Describe("GetPrivateDomain", func() {
		Context("when the private domain exists", func() {
			var expectedDomain ccv2.Domain

			BeforeEach(func() {
				expectedDomain = ccv2.Domain{
					GUID: "private-domain-guid",
					Name: "private-domain",
				}
				fakeCloudControllerClient.GetPrivateDomainReturns(expectedDomain, ccv2.Warnings{"private domain warning"}, nil)
			})

			It("returns the private domain and all warnings", func() {
				domain, warnings, err := actor.GetPrivateDomain("private-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domain).To(Equal(Domain(expectedDomain)))
				Expect(warnings).To(ConsistOf("private domain warning"))
			})
		})

		Context("when the API returns a not found error", func() {
			var expectedErr DomainNotFoundError

			BeforeEach(func() {
				fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, ccv2.Warnings{"private domain warning"}, ccerror.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError and all warnings", func() {
				domain, warnings, err := actor.GetPrivateDomain("private-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("private domain warning"))
			})
		})

		Context("when the API returns any other error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("private domain error")
				fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, ccv2.Warnings{"private domain warning"}, expectedErr)
			})

			It("returns the same error and all warnings", func() {
				domain, warnings, err := actor.GetPrivateDomain("private-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("private domain warning"))
			})
		})
	})

	Describe("GetOrganizationDomains", func() {
		Context("when the organization has both shared and private domains", func() {
			BeforeEach(func() {
				sharedDomain := ccv2.Domain{
					Name: "some-shared-domain",
				}
				privateDomain := ccv2.Domain{
					Name: "some-private-domain",
				}
				otherPrivateDomain := ccv2.Domain{
					Name: "some-other-private-domain",
				}

				fakeCloudControllerClient.GetSharedDomainsReturns([]ccv2.Domain{sharedDomain}, ccv2.Warnings{"shared domains warning"}, nil)
				fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns([]ccv2.Domain{privateDomain, otherPrivateDomain}, ccv2.Warnings{"private domains warning"}, nil)
			})

			It("returns a concatenated slice with private then shared domains", func() {
				domains, warnings, err := actor.GetOrganizationDomains("some-org-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(domains).To(Equal([]Domain{
					{Name: "some-private-domain"},
					{Name: "some-other-private-domain"},
					{Name: "some-shared-domain"},
				}))
				Expect(warnings).To(ConsistOf("shared domains warning", "private domains warning"))

				Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
				orgGUID, query := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(query).To(BeEmpty())
			})
		})
		Context("when get shared domains returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("shared domains error")
				fakeCloudControllerClient.GetSharedDomainsReturns([]ccv2.Domain{}, ccv2.Warnings{"shared domains warning"}, expectedErr)
			})

			It("returns that error", func() {
				domains, warnings, err := actor.GetOrganizationDomains("some-org-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domains).To(Equal([]Domain{}))
				Expect(warnings).To(ConsistOf("shared domains warning"))
			})
		})

		Context("when get organization private domains returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("private domains error")
				fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns([]ccv2.Domain{}, ccv2.Warnings{"private domains warning"}, expectedErr)
			})

			It("returns that error", func() {
				domains, warnings, err := actor.GetOrganizationDomains("some-org-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domains).To(Equal([]Domain{}))
				Expect(warnings).To(ConsistOf("private domains warning"))
			})
		})
	})
})
