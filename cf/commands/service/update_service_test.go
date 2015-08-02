package service_test

import (
	"errors"
	"io/ioutil"
	"os"

	testplanbuilder "github.com/cloudfoundry/cli/cf/actors/plan_builder/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("update-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		serviceRepo         *testapi.FakeServiceRepo
		planBuilder         *testplanbuilder.FakePlanBuilder
		offering1           models.ServiceOffering
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = config
		deps.PlanBuilder = planBuilder
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("update-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		config.SetApiVersion("2.26.0")
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}
		serviceRepo = &testapi.FakeServiceRepo{}
		planBuilder = &testplanbuilder.FakePlanBuilder{}

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			Guid: "cleardb-spark-guid",
		}, {
			Name: "flare",
			Guid: "cleardb-flare-guid",
		},
		}

	})

	var callUpdateService = func(args []string) bool {
		return testcmd.RunCliCommand("update-service", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callUpdateService([]string{"cleardb"})).To(BeTrue())
		})

		It("fails with usage when not provided exactly one arg", func() {
			Expect(callUpdateService([]string{})).To(BeFalse())
		})

		It("fails when not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})
	})

	Context("when no flags are passed", func() {

		Context("when the instance exists", func() {
			It("prints a user indicating it is a no-op", func() {
				callUpdateService([]string{"my-service"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
					[]string{"No changes were made"},
				))
			})
		})
	})

	Context("when passing arbitrary params", func() {
		BeforeEach(func() {
			serviceInstance := models.ServiceInstance{
				ServiceInstanceFields: models.ServiceInstanceFields{
					Name: "my-service-instance",
					Guid: "my-service-instance-guid",
					LastOperation: models.LastOperationFields{
						Type:        "update",
						State:       "in progress",
						Description: "fake service instance description",
					},
				},
				ServiceOffering: models.ServiceOfferingFields{
					Label: "murkydb",
					Guid:  "murkydb-guid",
				},
			}

			servicePlans := []models.ServicePlanFields{{
				Name: "spark",
				Guid: "murkydb-spark-guid",
			}, {
				Name: "flare",
				Guid: "murkydb-flare-guid",
			},
			}
			serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
			planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)
		})

		Context("as a json string", func() {
			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "-c", `{"foo": "bar"}`, "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))
				Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				It("returns an error to the UI", func() {
					callUpdateService([]string{"-p", "flare", "-c", `bad-json`, "my-service-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
					))
				})
			})
		})

		Context("as a file that contains json", func() {
			var jsonFile *os.File
			var params string

			BeforeEach(func() {
				params = "{\"foo\": \"bar\"}"
			})

			AfterEach(func() {
				if jsonFile != nil {
					jsonFile.Close()
					os.Remove(jsonFile.Name())
				}
			})

			JustBeforeEach(func() {
				var err error
				jsonFile, err = ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				err = ioutil.WriteFile(jsonFile.Name(), []byte(params), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("successfully updates a service and passes the params as a json", func() {
				callUpdateService([]string{"-p", "flare", "-c", jsonFile.Name(), "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))
				Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.Params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				BeforeEach(func() {
					params = "bad-json"
				})

				It("returns an error to the UI", func() {
					callUpdateService([]string{"-p", "flare", "-c", jsonFile.Name(), "my-service-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."},
					))
				})
			})
		})
	})

	Context("when passing in tags", func() {
		It("successfully updates a service and passes the tags as json", func() {
			callUpdateService([]string{"-t", "tag1, tag2,tag3,  tag4", "my-service-instance"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating service instance", "my-service-instance"},
				[]string{"OK"},
			))
			Expect(serviceRepo.UpdateServiceInstanceArgs.Tags).To(ConsistOf("tag1", "tag2", "tag3", "tag4"))
		})

		It("successfully updates a service and passes the tags as json", func() {
			callUpdateService([]string{"-t", "tag1", "my-service-instance"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Updating service instance", "my-service-instance"},
				[]string{"OK"},
			))
			Expect(serviceRepo.UpdateServiceInstanceArgs.Tags).To(ConsistOf("tag1"))
		})

		Context("and the tags string is passed with an empty string", func() {
			It("successfully updates the service", func() {
				callUpdateService([]string{"-t", "", "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating service instance", "my-service-instance"},
					[]string{"OK"},
				))
				Expect(serviceRepo.UpdateServiceInstanceArgs.Tags).To(Equal([]string{}))
			})
		})
	})

	Context("when service update is asynchronous", func() {
		Context("when the plan flag is passed", func() {
			BeforeEach(func() {
				serviceInstance := models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Name: "my-service-instance",
						Guid: "my-service-instance-guid",
						LastOperation: models.LastOperationFields{
							Type:        "update",
							State:       "in progress",
							Description: "fake service instance description",
						},
					},
					ServiceOffering: models.ServiceOfferingFields{
						Label: "murkydb",
						Guid:  "murkydb-guid",
					},
				}

				servicePlans := []models.ServicePlanFields{{
					Name: "spark",
					Guid: "murkydb-spark-guid",
				}, {
					Name: "flare",
					Guid: "murkydb-flare-guid",
				},
				}
				serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
				planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)
			})

			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))
				Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
			})

			Context("and the CC API Version >= 2.16.0", func() {
				BeforeEach(func() {
					config.SetApiVersion("2.16.0")
				})

				It("successfully updates a service", func() {
					callUpdateService([]string{"-p", "flare", "my-service-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Updating service", "my-service", "as", "my-user", "..."},
						[]string{"OK"},
						[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
					))
					Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
					Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
					Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
				})

				Context("when there is an err finding the instance", func() {
					It("returns an error", func() {
						serviceRepo.FindInstanceByNameErr = true

						callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Error finding instance"},
							[]string{"FAILED"},
						))
					})
				})
				Context("when there is an err finding service plans", func() {
					It("returns an error", func() {
						planBuilder.GetPlansForServiceForOrgReturns(nil, errors.New("Error fetching plans"))

						callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Error fetching plans"},
							[]string{"FAILED"},
						))
					})
				})
				Context("when the plan specified does not exist in the service offering", func() {
					It("returns an error", func() {
						callUpdateService([]string{"-p", "not-a-real-plan", "instance-without-service-offering"})

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Plan does not exist for the murkydb service"},
							[]string{"FAILED"},
						))
					})
				})
				Context("when there is an error updating the service instance", func() {
					It("returns an error", func() {
						serviceRepo.UpdateServiceInstanceReturnsErr = true
						callUpdateService([]string{"-p", "flare", "my-service-instance"})

						Expect(ui.Outputs).To(ContainSubstrings(
							[]string{"Error updating service instance"},
							[]string{"FAILED"},
						))
					})
				})
			})

			Context("and the CC API Version < 2.16.0", func() {
				BeforeEach(func() {
					config.SetApiVersion("2.15.0")
				})

				It("returns an error", func() {
					callUpdateService([]string{"-p", "flare", "instance-without-service-offering"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Updating a plan requires API v2.16.0 or newer. Your current target is v2.15.0."},
					))
				})
			})
		})
	})

	Context("when service update is synchronous", func() {
		Context("when the plan flag is passed", func() {
			BeforeEach(func() {
				serviceInstance := models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Name: "my-service-instance",
						Guid: "my-service-instance-guid",
					},
					ServiceOffering: models.ServiceOfferingFields{
						Label: "murkydb",
						Guid:  "murkydb-guid",
					},
				}

				servicePlans := []models.ServicePlanFields{{
					Name: "spark",
					Guid: "murkydb-spark-guid",
				}, {
					Name: "flare",
					Guid: "murkydb-flare-guid",
				},
				}
				serviceRepo.FindInstanceByNameServiceInstance = serviceInstance
				planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)

			})
			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
				))
				Expect(serviceRepo.FindInstanceByNameName).To(Equal("my-service-instance"))
				serviceGuid, orgName := planBuilder.GetPlansForServiceForOrgArgsForCall(0)
				Expect(serviceGuid).To(Equal("murkydb-guid"))
				Expect(orgName).To(Equal("my-org"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.InstanceGuid).To(Equal("my-service-instance-guid"))
				Expect(serviceRepo.UpdateServiceInstanceArgs.PlanGuid).To(Equal("murkydb-flare-guid"))
			})

			Context("when there is an err finding the instance", func() {
				It("returns an error", func() {
					serviceRepo.FindInstanceByNameErr = true

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Error finding instance"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an err finding service plans", func() {
				It("returns an error", func() {
					planBuilder.GetPlansForServiceForOrgReturns(nil, errors.New("Error fetching plans"))

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Error fetching plans"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when the plan specified does not exist in the service offering", func() {
				It("returns an error", func() {
					callUpdateService([]string{"-p", "not-a-real-plan", "instance-without-service-offering"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Plan does not exist for the murkydb service"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an error updating the service instance", func() {
				It("returns an error", func() {
					serviceRepo.UpdateServiceInstanceReturnsErr = true
					callUpdateService([]string{"-p", "flare", "my-service-instance"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Error updating service instance"},
						[]string{"FAILED"},
					))
				})
			})
		})

	})
})
