package v2action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("Domain", func() {
		var domain Domain

		Describe("IsHTTP", func() {
			When("the RouterGroupType = 'http'", func() {
				BeforeEach(func() {
					domain.RouterGroupType = constant.HTTPRouterGroup
				})

				It("returns true", func() {
					Expect(domain.IsHTTP()).To(BeTrue())
				})
			})

			When("the RouterGroupType is anything other than 'tcp'", func() {
				BeforeEach(func() {
					domain.RouterGroupType = ""
				})

				It("returns true", func() {
					Expect(domain.IsHTTP()).To(BeTrue())
				})
			})

			When("the RouterGroupType = 'http'", func() {
				BeforeEach(func() {
					domain.RouterGroupType = constant.TCPRouterGroup
				})

				It("returns false", func() {
					Expect(domain.IsHTTP()).To(BeFalse())
				})
			})
		})

		Describe("IsTCP", func() {
			When("the RouterGroupType = 'tcp'", func() {
				BeforeEach(func() {
					domain.RouterGroupType = constant.TCPRouterGroup
				})

				It("returns true", func() {
					Expect(domain.IsTCP()).To(BeTrue())
				})
			})

			When("the RouterGroupType is anything else", func() {
				BeforeEach(func() {
					domain.RouterGroupType = constant.HTTPRouterGroup
				})

				It("returns false", func() {
					Expect(domain.IsTCP()).To(BeFalse())
				})
			})
		})

		Describe("IsShared", func() {
			When("the the type is shared", func() {
				BeforeEach(func() {
					domain.Type = constant.SharedDomain
				})

				It("returns true", func() {
					Expect(domain.IsShared()).To(BeTrue())
				})
			})

			When("the RouterGroupType is anything else", func() {
				BeforeEach(func() {
					domain.Type = constant.PrivateDomain
				})

				It("returns false", func() {
					Expect(domain.IsShared()).To(BeFalse())
				})
			})
		})

		Describe("IsPrivate", func() {
			When("the the type is shared", func() {
				BeforeEach(func() {
					domain.Type = constant.PrivateDomain
				})

				It("returns true", func() {
					Expect(domain.IsPrivate()).To(BeTrue())
				})
			})

			When("the RouterGroupType is anything else", func() {
				BeforeEach(func() {
					domain.Type = constant.SharedDomain
				})

				It("returns false", func() {
					Expect(domain.IsPrivate()).To(BeFalse())
				})
			})
		})
	})

	Describe("DomainNotFoundError", func() {
		var err actionerror.DomainNotFoundError
		When("the name is provided", func() {
			BeforeEach(func() {
				err = actionerror.DomainNotFoundError{Name: "some-domain-name"}
			})

			It("returns the correct message", func() {
				Expect(err.Error()).To(Equal("Domain 'some-domain-name' not found."))
			})
		})

		When("the name is not provided but the guid is", func() {
			BeforeEach(func() {
				err = actionerror.DomainNotFoundError{GUID: "some-domain-guid"}
			})

			It("returns the correct message", func() {
				Expect(err.Error()).To(Equal("Domain with GUID 'some-domain-guid' not found."))
			})
		})

		When("neither the name nor the guid is provided", func() {
			BeforeEach(func() {
				err = actionerror.DomainNotFoundError{}
			})

			It("returns the correct message", func() {
				Expect(err.Error()).To(Equal("Domain not found."))
			})
		})
	})

	Describe("GetDomain", func() {
		When("the domain exists and is a shared domain", func() {
			var expectedDomain ccv2.Domain

			BeforeEach(func() {
				expectedDomain = ccv2.Domain{
					GUID: "shared-domain-guid",
					Name: "shared-domain",
				}
				fakeCloudControllerClient.GetSharedDomainReturns(expectedDomain, ccv2.Warnings{"get-domain-warning"}, nil)
			})

			It("returns the shared domain", func() {
				domain, warnings, err := actor.GetDomain("shared-domain-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-domain-warning"))
				Expect(domain).To(Equal(Domain(expectedDomain)))

				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("shared-domain-guid"))
			})
		})

		When("the domain exists and is a private domain", func() {
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

		When("the domain does not exist", func() {
			var expectedErr actionerror.DomainNotFoundError

			BeforeEach(func() {
				expectedErr = actionerror.DomainNotFoundError{GUID: "private-domain-guid"}
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

	Describe("GetDomains", func() {
		var (
			domains    []Domain
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			domains, warnings, executeErr = actor.GetDomains("some-org")
		})

		It("calls GetSharedDomains", func() {
			Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
		})

		When("GetSharedDomains returns successfully", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSharedDomainsReturns(
					[]ccv2.Domain{
						{
							Name: "some-domain.com",
							Type: constant.SharedDomain,
						},
						{
							Name: "some-other-domain.com",
							Type: constant.SharedDomain,
						},
					},
					ccv2.Warnings{"warning-1", "warning-2"},
					nil)
			})

			It("returns all shared domains and prints all warnings", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
				Expect(domains).To(ConsistOf(
					Domain{Name: "some-domain.com", Type: constant.SharedDomain},
					Domain{Name: "some-other-domain.com", Type: constant.SharedDomain},
				))
			})

			It("calls GetOrganizationPrivateDomains with the given org guid", func() {
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)).To(Equal("some-org"))
			})

			When("GetOrganizationPrivateDomains returns successfully", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(
						[]ccv2.Domain{
							{
								Name: "some-domain.private",
								Type: constant.PrivateDomain,
							},
							{
								Name: "some-other-domain.private",
								Type: constant.PrivateDomain,
							},
						},
						ccv2.Warnings{"warning-3", "warning-4"},
						nil)
				})

				It("returns the shared and private domains and returns all warnings", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
					Expect(domains).To(ConsistOf(
						Domain{Name: "some-domain.com", Type: constant.SharedDomain},
						Domain{Name: "some-other-domain.com", Type: constant.SharedDomain},
						Domain{Name: "some-domain.private", Type: constant.PrivateDomain},
						Domain{Name: "some-other-domain.private", Type: constant.PrivateDomain},
					))
				})
			})

			When("GetOrganizationPrivateDomains returns an error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(nil, ccv2.Warnings{"warning-3", "warning-4"}, errors.New("boom"))
				})

				It("returns the error and all warnings", func() {
					Expect(executeErr).To(MatchError("boom"))
					Expect(warnings).To(ConsistOf("warning-1", "warning-2", "warning-3", "warning-4"))
				})
			})
		})

		When("GetSharedDomains returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSharedDomainsReturns(nil, ccv2.Warnings{"warning-1", "warning-2"}, errors.New("boom"))
			})

			It("returns the error and all warnings", func() {
				Expect(executeErr).To(MatchError("boom"))
				Expect(warnings).To(ConsistOf("warning-1", "warning-2"))
			})

			It("does not call GetOrganizationPrivateDomains", func() {
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetDomainsByNameAndOrganization", func() {
		var (
			domainNames []string
			orgGUID     string

			domains    []Domain
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			domainNames = []string{"domain-1", "domain-2", "domain-3"}
			orgGUID = "some-org-guid"
		})

		JustBeforeEach(func() {
			domains, warnings, executeErr = actor.GetDomainsByNameAndOrganization(domainNames, orgGUID)
		})

		When("looking up the shared domains is successful", func() {
			var sharedDomains []ccv2.Domain

			BeforeEach(func() {
				sharedDomains = []ccv2.Domain{
					{Name: "domain-1", GUID: "shared-domain-1"},
				}
				fakeCloudControllerClient.GetSharedDomainsReturns(sharedDomains, ccv2.Warnings{"shared-warning-1", "shared-warning-2"}, nil)
			})

			When("looking up the private domains is successful", func() {
				var privateDomains []ccv2.Domain

				BeforeEach(func() {
					privateDomains = []ccv2.Domain{
						{Name: "domain-2", GUID: "private-domain-2"},
						{Name: "domain-3", GUID: "private-domain-3"},
					}
					fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(privateDomains, ccv2.Warnings{"private-warning-1", "private-warning-2"}, nil)
				})

				It("returns the domains and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("shared-warning-1", "shared-warning-2", "private-warning-1", "private-warning-2"))
					Expect(domains).To(ConsistOf(
						Domain{Name: "domain-1", GUID: "shared-domain-1"},
						Domain{Name: "domain-2", GUID: "private-domain-2"},
						Domain{Name: "domain-3", GUID: "private-domain-3"},
					))

					Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSharedDomainsArgsForCall(0)).To(ConsistOf(ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.InOperator,
						Values:   domainNames,
					}))

					Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
					passedOrgGUID, filters := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
					Expect(filters).To(ConsistOf(ccv2.Filter{
						Type:     constant.NameFilter,
						Operator: constant.InOperator,
						Values:   domainNames,
					}))
					Expect(passedOrgGUID).To(Equal(orgGUID))
				})
			})

			When("looking up the private domains errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("foobar")
					fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(nil, ccv2.Warnings{"private-warning-1", "private-warning-2"}, expectedErr)
				})

				It("returns errors and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("shared-warning-1", "shared-warning-2", "private-warning-1", "private-warning-2"))
				})
			})
		})

		When("no domains are provided", func() {
			BeforeEach(func() {
				domainNames = nil
			})

			It("immediately returns", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(BeEmpty())
				Expect(domains).To(BeEmpty())

				Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(0))
				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(0))
			})
		})

		When("looking up the shared domains errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("foobar")
				fakeCloudControllerClient.GetSharedDomainsReturns(nil, ccv2.Warnings{"shared-warning-1", "shared-warning-2"}, expectedErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("shared-warning-1", "shared-warning-2"))
			})
		})
	})

	Describe("GetSharedDomain", func() {
		When("the shared domain exists", func() {
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

				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("shared-domain-guid"))
			})

			When("the domain has been looked up multiple times", func() {
				It("caches the domain", func() {
					domain, warnings, err := actor.GetSharedDomain("shared-domain-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(domain).To(Equal(Domain(expectedDomain)))
					Expect(warnings).To(ConsistOf("shared domain warning"))

					domain, warnings, err = actor.GetSharedDomain("shared-domain-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(domain).To(Equal(Domain(expectedDomain)))
					Expect(warnings).To(BeEmpty())

					Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
				})
			})
		})

		When("the API returns a not found error", func() {
			var expectedErr actionerror.DomainNotFoundError

			BeforeEach(func() {
				expectedErr = actionerror.DomainNotFoundError{GUID: "shared-domain-guid"}
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"shared domain warning"}, ccerror.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError and all warnings", func() {
				domain, warnings, err := actor.GetSharedDomain("shared-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("shared domain warning"))
			})
		})

		When("the API returns any other error", func() {
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
		When("the private domain exists", func() {
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

				Expect(fakeCloudControllerClient.GetPrivateDomainCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetPrivateDomainArgsForCall(0)).To(Equal("private-domain-guid"))
			})

			When("the domain has been looked up multiple times", func() {
				It("caches the domain", func() {
					domain, warnings, err := actor.GetPrivateDomain("private-domain-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(domain).To(Equal(Domain(expectedDomain)))
					Expect(warnings).To(ConsistOf("private domain warning"))

					domain, warnings, err = actor.GetPrivateDomain("private-domain-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(domain).To(Equal(Domain(expectedDomain)))
					Expect(warnings).To(BeEmpty())

					Expect(fakeCloudControllerClient.GetPrivateDomainCallCount()).To(Equal(1))
				})
			})
		})

		When("the API returns a not found error", func() {
			var expectedErr actionerror.DomainNotFoundError

			BeforeEach(func() {
				expectedErr = actionerror.DomainNotFoundError{GUID: "private-domain-guid"}
				fakeCloudControllerClient.GetPrivateDomainReturns(ccv2.Domain{}, ccv2.Warnings{"private domain warning"}, ccerror.ResourceNotFoundError{})
			})

			It("returns a DomainNotFoundError and all warnings", func() {
				domain, warnings, err := actor.GetPrivateDomain("private-domain-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(domain).To(Equal(Domain{}))
				Expect(warnings).To(ConsistOf("private domain warning"))
			})
		})

		When("the API returns any other error", func() {
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
		When("the organization has both shared and private domains", func() {
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
					{Name: "some-shared-domain"},
					{Name: "some-private-domain"},
					{Name: "some-other-private-domain"},
				}))
				Expect(warnings).To(ConsistOf("shared domains warning", "private domains warning"))

				Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))

				Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
				orgGUID, filters := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
				Expect(orgGUID).To(Equal("some-org-guid"))
				Expect(filters).To(BeEmpty())
			})
		})

		When("get shared domains returns an error", func() {
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

		When("get organization private domains returns an error", func() {
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
