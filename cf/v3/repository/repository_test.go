package repository_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/v3/models"
	"code.cloudfoundry.org/cli/cf/v3/repository"
	"code.cloudfoundry.org/cli/util/testhelpers/configuration"

	ccClientFakes "github.com/cloudfoundry/go-ccapi/v3/client/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Repository", func() {
	var (
		r        repository.Repository
		ccClient *ccClientFakes.FakeClient
		config   coreconfig.ReadWriter
	)

	BeforeEach(func() {
		ccClient = &ccClientFakes.FakeClient{}
		config = configuration.NewRepositoryWithDefaults()
		r = repository.NewRepository(config, ccClient)
	})

	Describe("GetApplications", func() {
		It("tries to get applications from CC with a token handler", func() {
			r.GetApplications()
			Expect(ccClient.GetApplicationsCallCount()).To(Equal(1))
		})

		Context("when the client has updated tokens", func() {
			BeforeEach(func() {
				ccClient.TokensUpdatedReturns(true)
				ccClient.GetUpdatedTokensReturns("updated-access-token", "updated-refresh-token")
			})

			It("stores the new tokens in the config", func() {
				r.GetApplications()
				Expect(config.AccessToken()).To(Equal("updated-access-token"))
				Expect(config.RefreshToken()).To(Equal("updated-refresh-token"))
			})
		})

		Context("when getting the applications succeeds", func() {
			BeforeEach(func() {
				ccClient.GetApplicationsReturns(getApplicationsJSON, nil)
			})

			It("returns a slice of application model objects", func() {
				applications, err := r.GetApplications()
				Expect(err).NotTo(HaveOccurred())
				Expect(applications).To(Equal([]models.V3Application{
					{
						Name:                  "app-1-name",
						DesiredState:          "STOPPED",
						TotalDesiredInstances: 1,
						Links: models.Links{
							Processes: models.Link{
								Href: "/v3/apps/app-1-guid/processes",
							},
							Routes: models.Link{
								Href: "/v3/apps/app-1-guid/routes",
							},
						},
					},
					{
						Name:                  "app-2-name",
						DesiredState:          "RUNNING",
						TotalDesiredInstances: 2,
						Links: models.Links{
							Processes: models.Link{
								Href: "/v3/apps/app-2-guid/processes",
							},
							Routes: models.Link{
								Href: "/v3/apps/app-2-guid/routes",
							},
						},
					},
				}))
			})
		})

		Context("when getting the applications returns JSON that cannot be parsed", func() {
			BeforeEach(func() {
				ccClient.GetApplicationsReturns([]byte(`:bad_json:`), nil)
			})

			It("returns a slice of application model objects", func() {
				_, err := r.GetApplications()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetProcesses", func() {
		It("tries to get processes from CC with a token handler", func() {
			r.GetProcesses("/the-path")
			Expect(ccClient.GetResourcesCallCount()).To(Equal(1))
			Expect(ccClient.GetResourcesArgsForCall(0)).To(Equal("/the-path"))
		})

		Context("when the client has updated tokens", func() {
			BeforeEach(func() {
				ccClient.TokensUpdatedReturns(true)
				ccClient.GetUpdatedTokensReturns("updated-access-token", "updated-refresh-token")
			})

			It("stores the new tokens in the config", func() {
				r.GetProcesses("/the-path")
				Expect(config.AccessToken()).To(Equal("updated-access-token"))
				Expect(config.RefreshToken()).To(Equal("updated-refresh-token"))
			})
		})

		Context("when getting the processes fails", func() {
			BeforeEach(func() {
				ccClient.GetResourcesReturns([]byte{}, errors.New("get-processes-err"))
			})

			It("returns an error", func() {
				_, err := r.GetProcesses("/the-path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("get-processes-err"))
			})
		})

		Context("when getting the processes succeeds", func() {
			BeforeEach(func() {
				ccClient.GetResourcesReturns(getProcessesJSON, nil)
			})

			It("returns a slice of procees model objects", func() {
				processes, err := r.GetProcesses("/the-path")
				Expect(err).NotTo(HaveOccurred())
				Expect(processes).To(Equal([]models.V3Process{
					{
						Type:       "web",
						Instances:  1,
						MemoryInMB: 1024,
						DiskInMB:   1024,
					},
					{
						Type:       "web",
						Instances:  2,
						MemoryInMB: 512,
						DiskInMB:   512,
					},
				}))
			})
		})
	})

	Describe("GetRoutes", func() {
		It("tries to get routes from CC with a token handler", func() {
			r.GetRoutes("/the-path")
			Expect(ccClient.GetResourcesCallCount()).To(Equal(1))
			Expect(ccClient.GetResourcesArgsForCall(0)).To(Equal("/the-path"))
		})

		Context("when the client has updated tokens", func() {
			BeforeEach(func() {
				ccClient.TokensUpdatedReturns(true)
				ccClient.GetUpdatedTokensReturns("updated-access-token", "updated-refresh-token")
			})

			It("stores the new tokens in the config", func() {
				r.GetRoutes("/the-path")
				Expect(config.AccessToken()).To(Equal("updated-access-token"))
				Expect(config.RefreshToken()).To(Equal("updated-refresh-token"))
			})
		})

		Context("when getting the routes fails", func() {
			BeforeEach(func() {
				ccClient.GetResourcesReturns([]byte{}, errors.New("get-routes-err"))
			})

			It("returns an error", func() {
				_, err := r.GetRoutes("/the-path")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("get-routes-err"))
			})
		})

		Context("when getting the routes succeeds", func() {
			BeforeEach(func() {
				ccClient.GetResourcesReturns(getRoutesJSON, nil)
			})

			It("returns a slice of routes model objects", func() {
				routes, err := r.GetRoutes("/the-path")
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(Equal([]models.V3Route{
					{
						Host: "route-1-host",
						Path: "/route-1-path",
					},
					{
						Host: "route-2-host",
						Path: "",
					},
				}))
			})
		})
	})
})

var getApplicationsJSON = []byte(`[
{
	"guid": "app-1-guid",
	"name": "app-1-name",
	"desired_state": "STOPPED",
	"total_desired_instances": 1,
	"created_at": "1970-01-01T00:00:03Z",
	"lifecycle": {
		"type": "buildpack",
		"data": {
			"buildpack": "app-1-buildpack-name",
			"stack": "app-1-buildpack-stack"
		}
	},
	"environment_variables": {
		"key": "value"
	},
	"links": {
			"self": {
				"href": "/v3/apps/app-1-guid"
			},
			"space": {
				"href": "/v2/spaces/space-guid"
			},
			"processes": {
				"href": "/v3/apps/app-1-guid/processes"
			},
			"routes": {
				"href": "/v3/apps/app-1-guid/routes"
			},
			"packages": {
				"href": "/v3/apps/app-1-guid/packages"
			},
			"droplets": {
				"href": "/v3/apps/app-1-guid/droplets"
			},
			"start": {
				"href": "/v3/apps/app-1-guid/start",
				"method": "PUT"
			},
			"stop": {
				"href": "/v3/apps/app-1-guid/stop",
				"method": "PUT"
			},
			"assign_current_droplet": {
				"href": "/v3/apps/app-1-guid/current_droplet",
				"method": "PUT"
			}
		}
},
{
	"guid": "app-2-guid",
	"name": "app-2-name",
	"desired_state": "RUNNING",
	"total_desired_instances": 2,
	"created_at": "1970-01-01T00:00:03Z",
	"lifecycle": {
		"type": "buildpack",
		"data": {
			"buildpack": "app-2-buildpack-name",
			"stack": "app-2-buildpack-stack"
		}
	},
	"environment_variables": {},
	"links": {
			"self": {
				"href": "/v3/apps/app-2-guid"
			},
			"space": {
				"href": "/v2/spaces/space-guid"
			},
			"processes": {
				"href": "/v3/apps/app-2-guid/processes"
			},
			"routes": {
				"href": "/v3/apps/app-2-guid/routes"
			},
			"packages": {
				"href": "/v3/apps/app-2-guid/packages"
			},
			"droplets": {
				"href": "/v3/apps/app-2-guid/droplets"
			},
			"start": {
				"href": "/v3/apps/app-2-guid/start",
				"method": "PUT"
			},
			"stop": {
				"href": "/v3/apps/app-2-guid/stop",
				"method": "PUT"
			},
			"assign_current_droplet": {
				"href": "/v3/apps/app-2-guid/current_droplet",
				"method": "PUT"
			}
		}
	}
]`)

var getProcessesJSON = []byte(`[
	{
	  "guid": "process-1-guid",
	  "type": "web",
	  "command": null,
	  "instances": 1,
	  "memory_in_mb": 1024,
	  "disk_in_mb": 1024,
	  "created_at": "2015-12-22T18:28:11Z",
	  "updated_at": "2015-12-22T18:28:11Z",
	  "links": {
	    "self": {
	      "href": "/v3/processes/process-1-guid"
	    },
	    "scale": {
	      "href": "/v3/processes/process-1-guid/scale",
	      "method": "PUT"
	    },
	    "app": {
	      "href": "/v3/apps/app-1-guid"
	    },
	    "space": {
	      "href": "/v2/spaces/process-1-guid"
	    }
	  }
	},
	{
		"guid": "process-2-guid",
		"type": "web",
		"command": null,
		"instances": 2,
		"memory_in_mb": 512,
		"disk_in_mb": 512,
		"created_at": "2015-12-22T18:28:11Z",
		"updated_at": "2015-12-22T18:28:11Z",
		"links": {
			"self": {
				"href": "/v3/processes/process-2-guid"
			},
			"scale": {
				"href": "/v3/processes/process-2-guid/scale",
				"method": "PUT"
			},
			"app": {
				"href": "/v3/apps/app-2-guid"
			},
			"space": {
				"href": "/v2/spaces/process-2-guid"
			}
		}
	}
]`)

var getRoutesJSON = []byte(`
[
  {
    "guid": "8e1e3d10-5c77-48e7-9c15-38d4c0946a1c",
    "host": "route-1-host",
    "path": "/route-1-path",
    "created_at": "2015-12-22T18:28:01Z",
    "updated_at": null,
    "links": {
      "space": {
        "href": "/v2/spaces/be76f146-2139-4589-96a3-3489bfc888d0"
      },
      "domain": {
        "href": "/v2/domains/235d9128-b3d9-4b6b-b1c3-911159590e3c"
      }
    }
  },
  {
    "guid": "efab3642-71b1-4ce9-ab7b-28515241e090",
    "host": "route-2-host",
    "path": "",
    "created_at": "2015-12-22T18:28:01Z",
    "updated_at": null,
    "links": {
      "space": {
        "href": "/v2/spaces/be76f146-2139-4589-96a3-3489bfc888d0"
      },
      "domain": {
        "href": "/v2/domains/3cda59fc-ce99-463e-9fe9-e9a325fb8797"
      }
    }
  }
]`)
