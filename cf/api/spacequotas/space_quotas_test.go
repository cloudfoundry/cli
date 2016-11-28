package spacequotas_test

import (
	"encoding/json"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/cf/api/spacequotas"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/net"
	"github.com/onsi/gomega/ghttp"

	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerQuotaRepository", func() {
	var (
		ccServer   *ghttp.Server
		configRepo coreconfig.ReadWriter
		repo       spacequotas.CloudControllerSpaceQuotaRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetAPIEndpoint(ccServer.URL())
		gateway := net.NewCloudControllerGateway(configRepo, time.Now, new(terminalfakes.FakeUI), new(tracefakes.FakePrinter), "")
		repo = spacequotas.NewCloudControllerSpaceQuotaRepository(configRepo, gateway)
	})

	AfterEach(func() {
		ccServer.Close()
	})

	Describe("FindByName", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions"),
					ghttp.RespondWith(http.StatusOK, `{
						"next_url": "/v2/organizations/my-org-guid/space_quota_definitions?page=2",
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid" },
								"entity": {
									"name": "my-remote-quota",
									"memory_limit": 1024,
									"total_routes": 123,
									"total_services": 321,
									"non_basic_services_allowed": true,
									"organization_guid": "my-org-guid",
									"app_instance_limit": 333,
									"total_reserved_route_ports": 14
								}
							}
						]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions", "page=2"),
					ghttp.RespondWith(http.StatusOK, `{
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid2" },
								"entity": { "name": "my-remote-quota2", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							},
							{
								"metadata": { "guid": "my-quota-guid3" },
								"entity": { "name": "my-remote-quota3", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							}
						]
				  }`),
				),
			)
		})

		It("Finds Quota definitions by name", func() {
			quota, err := repo.FindByName("my-remote-quota")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
			Expect(quota).To(Equal(models.SpaceQuota{
				GUID:                    "my-quota-guid",
				Name:                    "my-remote-quota",
				MemoryLimit:             1024,
				RoutesLimit:             123,
				ServicesLimit:           321,
				NonBasicServicesAllowed: true,
				OrgGUID:                 "my-org-guid",
				AppInstanceLimit:        333,
				ReservedRoutePortsLimit: "14",
			}))
		})

		It("Returns an error if the quota cannot be found", func() {
			_, err := repo.FindByName("totally-not-a-quota")
			Expect(err.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("FindByNameAndOrgGUID", func() {
		Context("when the org exists", func() {
			Context("when the app_instance_limit is provided", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/organizations/other-org-guid/space_quota_definitions"),
							ghttp.RespondWith(http.StatusOK, `{
						"next_url": "/v2/organizations/other-org-guid/space_quota_definitions?page=2",
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid" },
								"entity": {
									"name": "my-remote-quota",
									"memory_limit": 1024,
									"total_routes": 123,
									"total_services": 321,
									"non_basic_services_allowed": true,
									"organization_guid": "other-org-guid",
									"app_instance_limit": 333,
									"total_reserved_route_ports": 14
								}
							}
						]
					}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/organizations/other-org-guid/space_quota_definitions", "page=2"),
							ghttp.RespondWith(http.StatusOK, `{
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid2" },
								"entity": { "name": "my-remote-quota2", "memory_limit": 1024, "organization_guid": "other-org-guid" }
							},
							{
								"metadata": { "guid": "my-quota-guid3" },
								"entity": { "name": "my-remote-quota3", "memory_limit": 1024, "organization_guid": "other-org-guid" }
							}
						]
				  }`),
						),
					)
				})

				It("Finds Quota definitions by name and org guid", func() {
					quota, err := repo.FindByNameAndOrgGUID("my-remote-quota", "other-org-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					Expect(quota).To(Equal(models.SpaceQuota{
						GUID:                    "my-quota-guid",
						Name:                    "my-remote-quota",
						MemoryLimit:             1024,
						RoutesLimit:             123,
						ServicesLimit:           321,
						NonBasicServicesAllowed: true,
						OrgGUID:                 "other-org-guid",
						AppInstanceLimit:        333,
						ReservedRoutePortsLimit: "14",
					}))
				})

				It("Returns an error if the quota cannot be found", func() {
					_, err := repo.FindByNameAndOrgGUID("totally-not-a-quota", "other-org-guid")
					Expect(err.(*errors.ModelNotFoundError)).To(HaveOccurred())
				})
			})

			Context("when the app_instance_limit and total_reserved_route_ports are not provided", func() {
				BeforeEach(func() {
					ccServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/organizations/other-org-guid/space_quota_definitions"),
							ghttp.RespondWith(http.StatusOK, `{
						"next_url": "/v2/organizations/other-org-guid/space_quota_definitions?page=2",
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid" },
								"entity": {
									"name": "my-remote-quota",
									"memory_limit": 1024,
									"total_routes": 123,
									"total_services": 321,
									"non_basic_services_allowed": true,
									"organization_guid": "other-org-guid"
								}
							}
						]
					}`),
						),
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/v2/organizations/other-org-guid/space_quota_definitions", "page=2"),
							ghttp.RespondWith(http.StatusOK, `{
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid2" },
								"entity": { "name": "my-remote-quota2", "memory_limit": 1024, "organization_guid": "other-org-guid" }
							},
							{
								"metadata": { "guid": "my-quota-guid3" },
								"entity": { "name": "my-remote-quota3", "memory_limit": 1024, "organization_guid": "other-org-guid" }
							}
						]
				  }`),
						),
					)
				})

				It("sets app instance limit to -1 and ReservedRoutePortsLimit is left blank", func() {
					quota, err := repo.FindByNameAndOrgGUID("my-remote-quota", "other-org-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					Expect(quota).To(Equal(models.SpaceQuota{
						GUID:                    "my-quota-guid",
						Name:                    "my-remote-quota",
						MemoryLimit:             1024,
						RoutesLimit:             123,
						ServicesLimit:           321,
						NonBasicServicesAllowed: true,
						OrgGUID:                 "other-org-guid",
						AppInstanceLimit:        -1,
						ReservedRoutePortsLimit: "",
					}))
				})
			})
		})

		Context("when the org does not exist", func() {
			BeforeEach(func() {
				ccServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v2/organizations/totally-not-an-org/space_quota_definitions"),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)
			})

			It("returns an error", func() {
				_, err := repo.FindByNameAndOrgGUID("my-remote-quota", "totally-not-an-org")
				Expect(err.(*errors.HTTPNotFoundError)).To(HaveOccurred())
			})
		})
	})

	Describe("FindByOrg", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions"),
					ghttp.RespondWith(http.StatusOK, `{
						"next_url": "/v2/organizations/my-org-guid/space_quota_definitions?page=2",
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid" },
								"entity": {
									"name": "my-remote-quota",
									"memory_limit": 1024,
									"total_routes": 123,
									"total_services": 321,
									"non_basic_services_allowed": true,
									"organization_guid": "my-org-guid",
									"total_reserved_route_ports": 14
								}
							}
						]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions", "page=2"),
					ghttp.RespondWith(http.StatusOK, `{
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid2" },
								"entity": { "name": "my-remote-quota2", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							},
							{
								"metadata": { "guid": "my-quota-guid3" },
								"entity": { "name": "my-remote-quota3", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							}
						]
				  }`),
				),
			)
		})

		It("finds all quota definitions by org guid", func() {
			quotas, err := repo.FindByOrg("my-org-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(quotas).To(HaveLen(3))

			Expect(quotas[0].GUID).To(Equal("my-quota-guid"))
			Expect(quotas[0].Name).To(Equal("my-remote-quota"))
			Expect(quotas[0].MemoryLimit).To(Equal(int64(1024)))
			Expect(quotas[0].RoutesLimit).To(Equal(123))
			Expect(quotas[0].ServicesLimit).To(Equal(321))
			Expect(quotas[0].OrgGUID).To(Equal("my-org-guid"))
			Expect(quotas[0].ReservedRoutePortsLimit).To(Equal(json.Number("14")))

			Expect(quotas[1].GUID).To(Equal("my-quota-guid2"))
			Expect(quotas[1].OrgGUID).To(Equal("my-org-guid"))
			Expect(quotas[2].GUID).To(Equal("my-quota-guid3"))
			Expect(quotas[2].OrgGUID).To(Equal("my-org-guid"))
		})
	})

	Describe("FindByGUID", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions"),
					ghttp.RespondWith(http.StatusOK, `{
						"next_url": "/v2/organizations/my-org-guid/space_quota_definitions?page=2",
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid" },
								"entity": {
									"name": "my-remote-quota",
									"memory_limit": 1024,
									"total_routes": 123,
									"total_services": 321,
									"non_basic_services_allowed": true,
									"organization_guid": "my-org-guid",
									"total_reserved_route_ports": 14
								}
							}
						]
					}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v2/organizations/my-org-guid/space_quota_definitions", "page=2"),
					ghttp.RespondWith(http.StatusOK, `{
						"resources": [
							{
								"metadata": { "guid": "my-quota-guid2" },
								"entity": { "name": "my-remote-quota2", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							},
							{
								"metadata": { "guid": "my-quota-guid3" },
								"entity": { "name": "my-remote-quota3", "memory_limit": 1024, "organization_guid": "my-org-guid" }
							}
						]
				  }`),
				),
			)
		})

		It("Finds Quota definitions by GUID", func() {
			quota, err := repo.FindByGUID("my-quota-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(quota).To(Equal(models.SpaceQuota{
				GUID:                    "my-quota-guid",
				Name:                    "my-remote-quota",
				MemoryLimit:             1024,
				RoutesLimit:             123,
				ServicesLimit:           321,
				NonBasicServicesAllowed: true,
				OrgGUID:                 "my-org-guid",
				AppInstanceLimit:        -1,
				ReservedRoutePortsLimit: "14",
			}))
		})

		It("Returns an error if the quota cannot be found", func() {
			_, err := repo.FindByGUID("totally-not-a-quota-guid")
			Expect(err.(*errors.ModelNotFoundError)).NotTo(BeNil())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
		})
	})

	Describe("AssociateSpaceWithQuota", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.VerifyRequest("PUT", "/v2/space_quota_definitions/my-quota-guid/spaces/my-space-guid"),
				ghttp.RespondWith(http.StatusCreated, nil),
			)
		})

		It("sets the quota for a space", func() {
			err := repo.AssociateSpaceWithQuota("my-space-guid", "my-quota-guid")
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("UnassignQuotaFromSpace", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/v2/space_quota_definitions/my-quota-guid/spaces/my-space-guid"),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
		})

		It("deletes the association between the quota and the space", func() {
			err := repo.UnassignQuotaFromSpace("my-space-guid", "my-quota-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("Create", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v2/space_quota_definitions"),
					ghttp.VerifyJSON(`{
						"name": "not-so-strict",
						"non_basic_services_allowed": false,
						"total_services": 1,
						"total_routes": 12,
						"memory_limit": 123,
						"instance_memory_limit": 0,
						"organization_guid": "my-org-guid",
						"app_instance_limit": 10,
						"total_reserved_route_ports": 5
					}`),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
		})

		It("creates a new quota with the given name", func() {
			quota := models.SpaceQuota{
				Name:                    "not-so-strict",
				ServicesLimit:           1,
				RoutesLimit:             12,
				MemoryLimit:             123,
				OrgGUID:                 "my-org-guid",
				AppInstanceLimit:        10,
				ReservedRoutePortsLimit: "5",
			}
			err := repo.Create(quota)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("Update", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/v2/space_quota_definitions/my-quota-guid"),
					ghttp.VerifyJSON(`{
						"guid": "my-quota-guid",
						"non_basic_services_allowed": false,
						"name": "amazing-quota",
						"total_services": 1,
						"total_routes": 12,
						"memory_limit": 123,
						"instance_memory_limit": 1234,
						"organization_guid": "myorgguid",
						"app_instance_limit": 23,
						"total_reserved_route_ports": 5
					}`),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("updates an existing quota", func() {
			quota := models.SpaceQuota{
				GUID: "my-quota-guid",
				Name: "amazing-quota",
				NonBasicServicesAllowed: false,
				ServicesLimit:           1,
				RoutesLimit:             12,
				MemoryLimit:             123,
				InstanceMemoryLimit:     1234,
				AppInstanceLimit:        23,
				OrgGUID:                 "myorgguid",
				ReservedRoutePortsLimit: "5",
			}

			err := repo.Update(quota)
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			ccServer.AppendHandlers(
				ghttp.VerifyRequest("DELETE", "/v2/space_quota_definitions/my-quota-guid"),
				ghttp.RespondWith(http.StatusNoContent, nil),
			)
		})

		It("deletes the quota with the given name", func() {
			err := repo.Delete("my-quota-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(ccServer.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
