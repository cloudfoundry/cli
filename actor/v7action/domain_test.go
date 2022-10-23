package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Domain Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		fakeRoutingClient         *v7actionfakes.FakeRoutingClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, fakeRoutingClient, _ = NewTestActor()
	})

	Describe("CheckRoute", func() {
		var (
			domainName string
			hostname   string
			path       string
			port       int

			matches    bool
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			domainName = "domain-name"
			hostname = "host"
			path = "/path"

			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{{GUID: "domain-guid"}},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.CheckRouteReturns(
				true,
				ccv3.Warnings{"check-route-warning-1", "check-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			matches, warnings, executeErr = actor.CheckRoute(domainName, hostname, path, port)
		})

		It("delegates to the cloud controller client", func() {
			Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
			givenQuery := fakeCloudControllerClient.GetDomainsArgsForCall(0)
			Expect(givenQuery).To(Equal([]ccv3.Query{
				{Key: ccv3.NameFilter, Values: []string{domainName}},
			}))

			Expect(fakeCloudControllerClient.CheckRouteCallCount()).To(Equal(1))
			givenDomainGUID, givenHostname, givenPath, givenPort := fakeCloudControllerClient.CheckRouteArgsForCall(0)
			Expect(givenDomainGUID).To(Equal("domain-guid"))
			Expect(givenHostname).To(Equal(hostname))
			Expect(givenPath).To(Equal(path))
			Expect(givenPort).To(Equal(0))

			Expect(matches).To(BeTrue())
			Expect(warnings).To(ConsistOf("get-domains-warning", "check-route-warning-1", "check-route-warning-2"))
			Expect(executeErr).NotTo(HaveOccurred())
		})

		When("getting the domain by name errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{GUID: "domain-guid"}},
					ccv3.Warnings{"get-domains-warning"},
					errors.New("domain not found"),
				)
			})

			It("returns the error and any warnings", func() {
				Expect(warnings).To(ConsistOf("get-domains-warning"))
				Expect(executeErr).To(MatchError("domain not found"))
			})
		})

		When("checking the route errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CheckRouteReturns(
					true,
					ccv3.Warnings{"check-route-warning-1", "check-route-warning-2"},
					errors.New("failed to check route"),
				)
			})

			It("returns the error and any warnings", func() {
				Expect(warnings).To(ConsistOf("get-domains-warning", "check-route-warning-1", "check-route-warning-2"))
				Expect(executeErr).To(MatchError("failed to check route"))
			})
		})
	})

	Describe("CreateSharedDomain", func() {
		var (
			warnings    Warnings
			executeErr  error
			routerGroup string
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateSharedDomain("the-domain-name", true, routerGroup)
		})

		BeforeEach(func() {
			routerGroup = ""
			fakeCloudControllerClient.CreateDomainReturns(resources.Domain{}, ccv3.Warnings{"create-warning-1", "create-warning-2"}, errors.New("create-error"))
		})

		It("delegates to the cloud controller client", func() {
			Expect(executeErr).To(MatchError("create-error"))
			Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2"))

			Expect(fakeRoutingClient.GetRouterGroupsCallCount()).To(Equal(0))

			Expect(fakeCloudControllerClient.CreateDomainCallCount()).To(Equal(1))
			passedDomain := fakeCloudControllerClient.CreateDomainArgsForCall(0)
			Expect(passedDomain).To(Equal(
				resources.Domain{
					Name:        "the-domain-name",
					Internal:    types.NullBool{IsSet: true, Value: true},
					RouterGroup: "",
				},
			))
		})

		Context("when a router group name is provided", func() {
			BeforeEach(func() {
				routerGroup = "router-group"
				fakeRoutingClient.GetRouterGroupByNameReturns(
					router.RouterGroup{Name: routerGroup, GUID: "router-group-guid"}, nil,
				)
			})

			It("delegates to the cloud controller client", func() {
				Expect(executeErr).To(MatchError("create-error"))
				Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2"))

				Expect(fakeRoutingClient.GetRouterGroupByNameCallCount()).To(Equal(1))
				Expect(fakeRoutingClient.GetRouterGroupByNameArgsForCall(0)).To(Equal(routerGroup))

				Expect(fakeCloudControllerClient.CreateDomainCallCount()).To(Equal(1))
				passedDomain := fakeCloudControllerClient.CreateDomainArgsForCall(0)

				Expect(passedDomain).To(Equal(
					resources.Domain{
						Name:        "the-domain-name",
						Internal:    types.NullBool{IsSet: true, Value: true},
						RouterGroup: "router-group-guid",
					},
				))
			})
		})
	})

	Describe("CreatePrivateDomain", func() {

		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]resources.Organization{
					{GUID: "org-guid"},
				},
				ccv3.Warnings{"get-orgs-warning"},
				nil,
			)

			fakeCloudControllerClient.CreateDomainReturns(
				resources.Domain{},
				ccv3.Warnings{"create-warning-1", "create-warning-2"},
				errors.New("create-error"),
			)
		})

		It("delegates to the cloud controller client", func() {
			warnings, executeErr := actor.CreatePrivateDomain("private-domain-name", "org-name")
			Expect(executeErr).To(MatchError("create-error"))
			Expect(warnings).To(ConsistOf("get-orgs-warning", "create-warning-1", "create-warning-2"))

			Expect(fakeCloudControllerClient.CreateDomainCallCount()).To(Equal(1))
			passedDomain := fakeCloudControllerClient.CreateDomainArgsForCall(0)

			Expect(passedDomain).To(Equal(
				resources.Domain{
					Name:             "private-domain-name",
					OrganizationGUID: "org-guid",
				},
			))
		})
	})

	Describe("delete domain", func() {
		var (
			domain resources.Domain
		)
		BeforeEach(func() {
			domain = resources.Domain{Name: "the-domain.com", GUID: "domain-guid"}
		})

		It("delegates to the cloud controller client", func() {
			fakeCloudControllerClient.DeleteDomainReturns(
				ccv3.JobURL("https://jobs/job_guid"),
				ccv3.Warnings{"delete-warning"},
				nil)

			warnings, executeErr := actor.DeleteDomain(domain)
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("delete-warning"))

			Expect(fakeCloudControllerClient.DeleteDomainCallCount()).To(Equal(1))
			passedDomainGuid := fakeCloudControllerClient.DeleteDomainArgsForCall(0)

			Expect(passedDomainGuid).To(Equal("domain-guid"))

			Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
			responseJobUrl := fakeCloudControllerClient.PollJobArgsForCall(0)
			Expect(responseJobUrl).To(Equal(ccv3.JobURL("https://jobs/job_guid")))
		})

		When("polling the job fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.PollJobReturns(
					ccv3.Warnings{"poll-job-warning"},
					errors.New("async-domain-delete-error"),
				)
			})

			It("returns the error", func() {
				warnings, err := actor.DeleteDomain(domain)
				Expect(err).To(MatchError("async-domain-delete-error"))
				Expect(warnings).To(ConsistOf(
					"poll-job-warning",
				))
			})
		})
	})

	Describe("list domains for org", func() {
		var (
			ccv3Domains []resources.Domain
			domains     []resources.Domain

			domain1Name string
			domain1Guid string

			domain2Name string
			domain2Guid string

			domain3Name string
			domain3Guid string

			org1Guid          = "some-org-guid"
			sharedFromOrgGuid = "another-org-guid"
			warnings          Warnings
			executeErr        error
			labelSelector     string
		)

		BeforeEach(func() {
			ccv3Domains = []resources.Domain{
				{Name: domain1Name, GUID: domain1Guid, OrganizationGUID: org1Guid},
				{Name: domain2Name, GUID: domain2Guid, OrganizationGUID: org1Guid},
				{Name: domain3Name, GUID: domain3Guid, OrganizationGUID: sharedFromOrgGuid},
			}
			labelSelector = "foo=bar"
		})

		JustBeforeEach(func() {
			domains, warnings, executeErr = actor.GetOrganizationDomains(org1Guid, labelSelector)
		})

		When("the API layer call is successful", func() {
			expectedOrgGUID := "some-org-guid"

			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationDomainsReturns(
					ccv3Domains,
					ccv3.Warnings{"some-domain-warning"},
					nil,
				)
			})

			It("returns back the domains and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
				actualOrgGuid, _ := fakeCloudControllerClient.GetOrganizationDomainsArgsForCall(0)
				Expect(actualOrgGuid).To(Equal(expectedOrgGUID))

				Expect(domains).To(ConsistOf(
					resources.Domain{Name: domain1Name, GUID: domain1Guid, OrganizationGUID: org1Guid},
					resources.Domain{Name: domain2Name, GUID: domain2Guid, OrganizationGUID: org1Guid},
					resources.Domain{Name: domain3Name, GUID: domain3Guid, OrganizationGUID: sharedFromOrgGuid},
				))
				Expect(warnings).To(ConsistOf("some-domain-warning"))

			})

			It("uses the label selector", func() {
				_, actualQuery := fakeCloudControllerClient.GetOrganizationDomainsArgsForCall(0)
				Expect(actualQuery).To(ContainElement(ccv3.Query{Key: ccv3.LabelSelectorFilter, Values: []string{"foo=bar"}}))
			})

			When("the label selector isn't specified", func() {
				BeforeEach(func() {
					labelSelector = ""
				})

				It("calls the API with no label selectors", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
					_, actualQuery := fakeCloudControllerClient.GetOrganizationDomainsArgsForCall(0)
					Expect(len(actualQuery)).To(Equal(0))
				})
			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationDomainsReturns(
					[]resources.Domain{},
					ccv3.Warnings{"some-domain-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-domain-warning"))
				Expect(domains).To(ConsistOf([]resources.Domain{}))

				Expect(fakeCloudControllerClient.GetOrganizationDomainsCallCount()).To(Equal(1))
			})

		})
	})

	Describe("get domain by guid", func() {
		var (
			domain      resources.Domain
			ccv3Domain  resources.Domain
			domain1Guid string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			domain1Guid = "domain-1-guid"
			ccv3Domain = resources.Domain{Name: "domain-1-name", GUID: domain1Guid}
		})

		JustBeforeEach(func() {

			domain, warnings, executeErr = actor.GetDomain(domain1Guid)
		})

		When("the API layer call is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainReturns(
					ccv3Domain,
					ccv3.Warnings{"some-domain-warning"},
					nil,
				)
			})

			It("returns back the domains and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainCallCount()).To(Equal(1))
				actualGUID := fakeCloudControllerClient.GetDomainArgsForCall(0)
				Expect(actualGUID).To(Equal(domain1Guid))

				Expect(domain).To(Equal(
					resources.Domain{Name: "domain-1-name", GUID: domain1Guid},
				))
				Expect(warnings).To(ConsistOf("some-domain-warning"))

			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainReturns(
					resources.Domain{},
					ccv3.Warnings{"some-domain-warning"},
					errors.New("get-domain-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("get-domain-error"))
				Expect(warnings).To(ConsistOf("some-domain-warning"))
				Expect(domain).To(Equal(resources.Domain{}))

				Expect(fakeCloudControllerClient.GetDomainCallCount()).To(Equal(1))
			})
		})
	})

	Describe("get domain by name", func() {
		var (
			ccv3Domains []resources.Domain
			domain      resources.Domain

			domain1Name string
			domain1Guid string

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			ccv3Domains = []resources.Domain{
				{Name: domain1Name, GUID: domain1Guid},
			}
		})

		JustBeforeEach(func() {
			domain, warnings, executeErr = actor.GetDomainByName(domain1Name)
		})

		When("the API layer call is successful", func() {
			expectedQuery := []ccv3.Query{{Key: ccv3.NameFilter, Values: []string{domain1Name}}}
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					ccv3Domains,
					ccv3.Warnings{"some-domain-warning"},
					nil,
				)
			})

			It("returns back the domains and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				actualQuery := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(actualQuery).To(Equal(expectedQuery))

				Expect(domain).To(Equal(
					resources.Domain{Name: domain1Name, GUID: domain1Guid},
				))
				Expect(warnings).To(ConsistOf("some-domain-warning"))

			})
		})

		When("when the API layer call returns an error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{},
					ccv3.Warnings{"some-domain-warning"},
					errors.New("list-error"),
				)
			})

			It("returns the error and prints warnings", func() {
				Expect(executeErr).To(MatchError("list-error"))
				Expect(warnings).To(ConsistOf("some-domain-warning"))
				Expect(domain).To(Equal(resources.Domain{}))

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
			})
		})
	})

	Describe("share private domain to org", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetOrganizationsReturns(
				[]resources.Organization{
					{GUID: "org-guid"},
				},
				ccv3.Warnings{"get-orgs-warning"},
				nil,
			)

			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{
					{Name: "private-domain.com", GUID: "private-domain-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.SharePrivateDomainToOrgsReturns(
				ccv3.Warnings{"share-domain-warning"},
				nil,
			)
		})

		It("delegates to the cloud controller client", func() {
			warnings, executeErr := actor.SharePrivateDomain("private-domain.com", "org-name")
			Expect(executeErr).To(BeNil())
			Expect(warnings).To(ConsistOf("share-domain-warning", "get-orgs-warning", "get-domains-warning"))

			Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
			actualQuery := fakeCloudControllerClient.GetDomainsArgsForCall(0)
			Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{"private-domain.com"}}}))

			Expect(fakeCloudControllerClient.SharePrivateDomainToOrgsCallCount()).To(Equal(1))
			domainGuid, sharedOrgs := fakeCloudControllerClient.SharePrivateDomainToOrgsArgsForCall(0)
			Expect(domainGuid).To(Equal("private-domain-guid"))
			Expect(sharedOrgs).To(Equal(ccv3.SharedOrgs{GUIDs: []string{"org-guid"}}))
		})
	})

	Describe("unshare private domain from org", func() {
		var (
			domainName string
			orgName    string
			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnsharePrivateDomain(domainName, orgName)
		})

		When("getting the org or domain errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					nil,
					ccv3.Warnings{"get-orgs-warning"},
					errors.New("get-orgs-error"),
				)
			})
			It("returns an error and the warnings", func() {
				Expect(fakeCloudControllerClient.UnsharePrivateDomainFromOrgCallCount()).To(Equal(0))

				Expect(executeErr).To(MatchError(errors.New("get-orgs-error")))
				Expect(warnings).To(ConsistOf("get-orgs-warning"))
			})
		})

		When("getting the guids succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{Name: domainName, GUID: "private-domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{
						{Name: orgName, GUID: "org-guid"},
					},
					ccv3.Warnings{"get-orgs-warning"},
					nil,
				)
			})

			When("Unsharing the domain errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnsharePrivateDomainFromOrgReturns(
						ccv3.Warnings{"unshare-domain-warning"},
						errors.New("unshare-domain-error"),
					)
				})

				It("returns any warnings and errors", func() {
					Expect(fakeCloudControllerClient.UnsharePrivateDomainFromOrgCallCount()).To(Equal(1))
					actualDomainGUID, actualOrgGUID := fakeCloudControllerClient.UnsharePrivateDomainFromOrgArgsForCall(0)
					Expect(actualDomainGUID).To(Equal("private-domain-guid"))
					Expect(actualOrgGUID).To(Equal("org-guid"))

					Expect(executeErr).To(MatchError(errors.New("unshare-domain-error")))
					Expect(warnings).To(ConsistOf("get-orgs-warning", "get-domains-warning", "unshare-domain-warning"))
				})
			})

			When("everything succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UnsharePrivateDomainFromOrgReturns(
						ccv3.Warnings{"unshare-domain-warning"},
						nil,
					)
				})

				It("returns any warnings and no error", func() {
					Expect(fakeCloudControllerClient.UnsharePrivateDomainFromOrgCallCount()).To(Equal(1))
					actualDomainGUID, actualOrgGUID := fakeCloudControllerClient.UnsharePrivateDomainFromOrgArgsForCall(0)
					Expect(actualDomainGUID).To(Equal("private-domain-guid"))
					Expect(actualOrgGUID).To(Equal("org-guid"))

					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-orgs-warning", "get-domains-warning", "unshare-domain-warning"))
				})
			})
		})
	})

	Describe("GetDomainAndOrgByNames", func() {
		var (
			orgName    = "my-org"
			domainName = "domain.com"

			orgGUID    string
			domainGUID string
			warnings   Warnings
			executeErr error
		)
		JustBeforeEach(func() {
			orgGUID, domainGUID, warnings, executeErr = actor.GetDomainAndOrgGUIDsByName(domainName, orgName)
		})

		When("Getting the organization is not successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					nil,
					ccv3.Warnings{"get-orgs-warning"},
					errors.New("get-orgs-error"),
				)
			})

			It("returns the error and doesnt get the domain", func() {
				Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
				actualQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
				Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{orgName}}}))

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(0))

				Expect(executeErr).To(MatchError(errors.New("get-orgs-error")))
				Expect(warnings).To(ConsistOf("get-orgs-warning"))
			})
		})

		When("getting the orgs succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetOrganizationsReturns(
					[]resources.Organization{
						{Name: orgName, GUID: "org-guid"},
					},
					ccv3.Warnings{"get-orgs-warning"},
					nil,
				)
			})
			When("Getting the domain is not successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						nil,
						ccv3.Warnings{"get-domains-warning"},
						errors.New("get-domains-error"),
					)
				})

				It("returns the error and doesnt get the domain", func() {
					Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
					actualQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
					Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{orgName}}}))

					Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
					actualQuery = fakeCloudControllerClient.GetDomainsArgsForCall(0)
					Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{domainName}}}))

					Expect(executeErr).To(MatchError(errors.New("get-domains-error")))
					Expect(warnings).To(ConsistOf("get-orgs-warning", "get-domains-warning"))
				})
			})

			When("the api call are successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{
							{Name: domainName, GUID: "private-domain-guid"},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)
				})

				It("returns the GUIDs", func() {
					Expect(fakeCloudControllerClient.GetOrganizationsCallCount()).To(Equal(1))
					actualQuery := fakeCloudControllerClient.GetOrganizationsArgsForCall(0)
					Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{orgName}}}))

					Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
					actualQuery = fakeCloudControllerClient.GetDomainsArgsForCall(0)
					Expect(actualQuery).To(Equal([]ccv3.Query{{Key: ccv3.NameFilter, Values: []string{domainName}}}))

					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-orgs-warning", "get-domains-warning"))
					Expect(orgGUID).To(Equal("org-guid"))
					Expect(domainGUID).To(Equal("private-domain-guid"))
				})
			})
		})
	})
})
