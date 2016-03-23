package space_quotas_test

import (
	"net/http"

	"github.com/cloudfoundry/cli/cf/api/space_quotas"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/onsi/gomega/ghttp"

	"github.com/cloudfoundry/cli/testhelpers/cloud_controller_gateway"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudControllerQuotaRepository", func() {
	var (
		ccServer   *ghttp.Server
		configRepo core_config.ReadWriter
		repo       space_quotas.CloudControllerSpaceQuotaRepository
	)

	BeforeEach(func() {
		ccServer = ghttp.NewServer()
		configRepo = testconfig.NewRepositoryWithDefaults()
		configRepo.SetApiEndpoint(ccServer.URL())
		gateway := cloud_controller_gateway.NewTestCloudControllerGateway(configRepo)
		repo = space_quotas.NewCloudControllerSpaceQuotaRepository(configRepo, gateway)
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
									"app_instance_limit": 333
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
				Guid:                    "my-quota-guid",
				Name:                    "my-remote-quota",
				MemoryLimit:             1024,
				RoutesLimit:             123,
				ServicesLimit:           321,
				NonBasicServicesAllowed: true,
				OrgGuid:                 "my-org-guid",
				AppInstanceLimit:        333,
			}))
		})

		It("Returns an error if the quota cannot be found", func() {
			_, err := repo.FindByName("totally-not-a-quota")
			Expect(err.(*errors.ModelNotFoundError)).NotTo(BeNil())
		})
	})

	Describe("FindByNameAndOrgGuid", func() {
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
									"app_instance_limit": 333
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
					quota, err := repo.FindByNameAndOrgGuid("my-remote-quota", "other-org-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					Expect(quota).To(Equal(models.SpaceQuota{
						Guid:                    "my-quota-guid",
						Name:                    "my-remote-quota",
						MemoryLimit:             1024,
						RoutesLimit:             123,
						ServicesLimit:           321,
						NonBasicServicesAllowed: true,
						OrgGuid:                 "other-org-guid",
						AppInstanceLimit:        333,
					}))
				})

				It("Returns an error if the quota cannot be found", func() {
					_, err := repo.FindByNameAndOrgGuid("totally-not-a-quota", "other-org-guid")
					Expect(err.(*errors.ModelNotFoundError)).To(HaveOccurred())
				})
			})

			Context("when the app_instance_limit is not provided", func() {
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

				It("sets app instance limit to -1", func() {
					quota, err := repo.FindByNameAndOrgGuid("my-remote-quota", "other-org-guid")
					Expect(err).NotTo(HaveOccurred())
					Expect(ccServer.ReceivedRequests()).To(HaveLen(2))
					Expect(quota).To(Equal(models.SpaceQuota{
						Guid:                    "my-quota-guid",
						Name:                    "my-remote-quota",
						MemoryLimit:             1024,
						RoutesLimit:             123,
						ServicesLimit:           321,
						NonBasicServicesAllowed: true,
						OrgGuid:                 "other-org-guid",
						AppInstanceLimit:        -1,
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
				_, err := repo.FindByNameAndOrgGuid("my-remote-quota", "totally-not-an-org")
				Expect(err.(*errors.HttpNotFoundError)).To(HaveOccurred())
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
									"organization_guid": "my-org-guid"
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

			Expect(quotas[0].Guid).To(Equal("my-quota-guid"))
			Expect(quotas[0].Name).To(Equal("my-remote-quota"))
			Expect(quotas[0].MemoryLimit).To(Equal(int64(1024)))
			Expect(quotas[0].RoutesLimit).To(Equal(123))
			Expect(quotas[0].ServicesLimit).To(Equal(321))
			Expect(quotas[0].OrgGuid).To(Equal("my-org-guid"))

			Expect(quotas[1].Guid).To(Equal("my-quota-guid2"))
			Expect(quotas[1].OrgGuid).To(Equal("my-org-guid"))
			Expect(quotas[2].Guid).To(Equal("my-quota-guid3"))
			Expect(quotas[2].OrgGuid).To(Equal("my-org-guid"))
		})
	})

	Describe("FindByGuid", func() {
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
									"organization_guid": "my-org-guid"
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

		It("Finds Quota definitions by Guid", func() {
			quota, err := repo.FindByGuid("my-quota-guid")
			Expect(err).NotTo(HaveOccurred())
			Expect(quota).To(Equal(models.SpaceQuota{
				Guid:                    "my-quota-guid",
				Name:                    "my-remote-quota",
				MemoryLimit:             1024,
				RoutesLimit:             123,
				ServicesLimit:           321,
				NonBasicServicesAllowed: true,
				OrgGuid:                 "my-org-guid",
				AppInstanceLimit:        -1,
			}))
		})

		It("Returns an error if the quota cannot be found", func() {
			_, err := repo.FindByGuid("totally-not-a-quota-guid")
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
						"app_instance_limit": 10
					}`),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)
		})

		It("creates a new quota with the given name", func() {
			quota := models.SpaceQuota{
				Name:             "not-so-strict",
				ServicesLimit:    1,
				RoutesLimit:      12,
				MemoryLimit:      123,
				OrgGuid:          "my-org-guid",
				AppInstanceLimit: 10,
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
						"app_instance_limit": 23
					}`),
					ghttp.RespondWith(http.StatusOK, nil),
				),
			)
		})

		It("updates an existing quota", func() {
			quota := models.SpaceQuota{
				Guid: "my-quota-guid",
				Name: "amazing-quota",
				NonBasicServicesAllowed: false,
				ServicesLimit:           1,
				RoutesLimit:             12,
				MemoryLimit:             123,
				InstanceMemoryLimit:     1234,
				AppInstanceLimit:        23,
				OrgGuid:                 "myorgguid",
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
