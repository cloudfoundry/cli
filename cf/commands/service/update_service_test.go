package service_test

import (
	"errors"
	"io/ioutil"
	"os"

	planbuilderfakes "code.cloudfoundry.org/cli/cf/actors/planbuilder/planbuilderfakes"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/commands/service"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("update-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		serviceRepo         *apifakes.FakeServiceRepository
		planBuilder         *planbuilderfakes.FakePlanBuilder
		offering1           models.ServiceOffering
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = config
		deps.PlanBuilder = planBuilder
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("update-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}

		config = testconfig.NewRepositoryWithDefaults()

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		requirementsFactory.NewMinAPIVersionRequirementReturns(requirements.Passing{Type: "minAPIVersionReq"})

		serviceRepo = new(apifakes.FakeServiceRepository)
		planBuilder = new(planbuilderfakes.FakePlanBuilder)

		offering1 = models.ServiceOffering{}
		offering1.Label = "cleardb"
		offering1.Plans = []models.ServicePlanFields{{
			Name: "spark",
			GUID: "cleardb-spark-guid",
		}, {
			Name: "flare",
			GUID: "cleardb-flare-guid",
		},
		}

	})

	var callUpdateService = func(args []string) bool {
		return testcmd.RunCLICommand("update-service", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("passes when logged in and a space is targeted", func() {
			Expect(callUpdateService([]string{"cleardb"})).To(BeTrue())
		})

		It("fails with usage when not provided exactly one arg", func() {
			Expect(callUpdateService([]string{})).To(BeFalse())
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(callUpdateService([]string{"cleardb", "spark", "my-cleardb-service"})).To(BeFalse())
		})

		Context("-p", func() {
			It("when provided, requires a CC API version > cf.UpdateServicePlanMinimumAPIVersion", func() {
				cmd := &service.UpdateService{}

				fc := flags.NewFlagContext(cmd.MetaData().Flags)
				fc.Parse("potato", "-p", "plan-name")

				reqs, err := cmd.Requirements(requirementsFactory, fc)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqs).NotTo(BeEmpty())

				Expect(reqs).To(ContainElement(requirements.Passing{Type: "minAPIVersionReq"}))
			})

			It("does not requirue a CC Api Version if not provided", func() {
				cmd := &service.UpdateService{}

				fc := flags.NewFlagContext(cmd.MetaData().Flags)
				fc.Parse("potato")

				reqs, err := cmd.Requirements(requirementsFactory, fc)
				Expect(err).NotTo(HaveOccurred())
				Expect(reqs).NotTo(BeEmpty())

				Expect(reqs).NotTo(ContainElement(requirements.Passing{Type: "minAPIVersionReq"}))
			})
		})
	})

	Context("when no flags are passed", func() {

		Context("when the instance exists", func() {
			It("prints a user indicating it is a no-op", func() {
				callUpdateService([]string{"my-service"})

				Expect(ui.Outputs()).To(ContainSubstrings(
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
					GUID: "my-service-instance-guid",
					LastOperation: models.LastOperationFields{
						Type:        "update",
						State:       "in progress",
						Description: "fake service instance description",
					},
				},
				ServiceOffering: models.ServiceOfferingFields{
					Label: "murkydb",
					GUID:  "murkydb-guid",
				},
			}

			servicePlans := []models.ServicePlanFields{{
				Name: "spark",
				GUID: "murkydb-spark-guid",
			}, {
				Name: "flare",
				GUID: "murkydb-flare-guid",
			},
			}
			serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
			planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)
		})

		Context("as a json string", func() {
			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "-c", `{"foo": "bar"}`, "my-service-instance"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))
				Expect(serviceRepo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service-instance"))

				instanceGUID, planGUID, params, _ := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(instanceGUID).To(Equal("my-service-instance-guid"))
				Expect(planGUID).To(Equal("murkydb-flare-guid"))
				Expect(params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				It("returns an error to the UI", func() {
					callUpdateService([]string{"-p", "flare", "-c", `bad-json`, "my-service-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))

				Expect(serviceRepo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service-instance"))

				instanceGUID, planGUID, params, _ := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(instanceGUID).To(Equal("my-service-instance-guid"))
				Expect(planGUID).To(Equal("murkydb-flare-guid"))
				Expect(params).To(Equal(map[string]interface{}{"foo": "bar"}))
			})

			Context("that are not valid json", func() {
				BeforeEach(func() {
					params = "bad-json"
				})

				It("returns an error to the UI", func() {
					callUpdateService([]string{"-p", "flare", "-c", jsonFile.Name(), "my-service-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
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

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Updating service instance", "my-service-instance"},
				[]string{"OK"},
			))
			_, _, _, tags := serviceRepo.UpdateServiceInstanceArgsForCall(0)
			Expect(tags).To(ConsistOf("tag1", "tag2", "tag3", "tag4"))
		})

		It("successfully updates a service and passes the tags as json", func() {
			callUpdateService([]string{"-t", "tag1", "my-service-instance"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Updating service instance", "my-service-instance"},
				[]string{"OK"},
			))
			_, _, _, tags := serviceRepo.UpdateServiceInstanceArgsForCall(0)
			Expect(tags).To(ConsistOf("tag1"))
		})

		Context("and the tags string is passed with an empty string", func() {
			It("successfully updates the service", func() {
				callUpdateService([]string{"-t", "", "my-service-instance"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service instance", "my-service-instance"},
					[]string{"OK"},
				))
				_, _, _, tags := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(tags).To(Equal([]string{}))
			})
		})
	})

	Context("when service update is asynchronous", func() {
		Context("when the plan flag is passed", func() {
			BeforeEach(func() {
				serviceInstance := models.ServiceInstance{
					ServiceInstanceFields: models.ServiceInstanceFields{
						Name: "my-service-instance",
						GUID: "my-service-instance-guid",
						LastOperation: models.LastOperationFields{
							Type:        "update",
							State:       "in progress",
							Description: "fake service instance description",
						},
					},
					ServiceOffering: models.ServiceOfferingFields{
						Label: "murkydb",
						GUID:  "murkydb-guid",
					},
				}

				servicePlans := []models.ServicePlanFields{{
					Name: "spark",
					GUID: "murkydb-spark-guid",
				}, {
					Name: "flare",
					GUID: "murkydb-flare-guid",
				},
				}
				serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
				planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)
			})

			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))

				Expect(serviceRepo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service-instance"))

				instanceGUID, planGUID, _, _ := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(instanceGUID).To(Equal("my-service-instance-guid"))
				Expect(planGUID).To(Equal("murkydb-flare-guid"))
			})

			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
					[]string{"Update in progress. Use 'cf services' or 'cf service my-service-instance' to check operation status."},
				))

				Expect(serviceRepo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service-instance"))

				instanceGUID, planGUID, _, _ := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(instanceGUID).To(Equal("my-service-instance-guid"))
				Expect(planGUID).To(Equal("murkydb-flare-guid"))
			})

			Context("when there is an err finding the instance", func() {
				It("returns an error", func() {
					serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("Error finding instance"))

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error finding instance"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an err finding service plans", func() {
				It("returns an error", func() {
					planBuilder.GetPlansForServiceForOrgReturns(nil, errors.New("Error fetching plans"))

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error fetching plans"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when the plan specified does not exist in the service offering", func() {
				It("returns an error", func() {
					callUpdateService([]string{"-p", "not-a-real-plan", "instance-without-service-offering"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Plan does not exist for the murkydb service"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an error updating the service instance", func() {
				It("returns an error", func() {
					serviceRepo.UpdateServiceInstanceReturns(errors.New("Error updating service instance"))
					callUpdateService([]string{"-p", "flare", "my-service-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error updating service instance"},
						[]string{"FAILED"},
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
						GUID: "my-service-instance-guid",
					},
					ServiceOffering: models.ServiceOfferingFields{
						Label: "murkydb",
						GUID:  "murkydb-guid",
					},
				}

				servicePlans := []models.ServicePlanFields{{
					Name: "spark",
					GUID: "murkydb-spark-guid",
				}, {
					Name: "flare",
					GUID: "murkydb-flare-guid",
				},
				}
				serviceRepo.FindInstanceByNameReturns(serviceInstance, nil)
				planBuilder.GetPlansForServiceForOrgReturns(servicePlans, nil)

			})
			It("successfully updates a service", func() {
				callUpdateService([]string{"-p", "flare", "my-service-instance"})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Updating service", "my-service", "as", "my-user", "..."},
					[]string{"OK"},
				))
				Expect(serviceRepo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service-instance"))
				serviceGUID, orgName := planBuilder.GetPlansForServiceForOrgArgsForCall(0)
				Expect(serviceGUID).To(Equal("murkydb-guid"))
				Expect(orgName).To(Equal("my-org"))

				instanceGUID, planGUID, _, _ := serviceRepo.UpdateServiceInstanceArgsForCall(0)
				Expect(instanceGUID).To(Equal("my-service-instance-guid"))
				Expect(planGUID).To(Equal("murkydb-flare-guid"))
			})

			Context("when there is an err finding the instance", func() {
				It("returns an error", func() {
					serviceRepo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.New("Error finding instance"))

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error finding instance"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an err finding service plans", func() {
				It("returns an error", func() {
					planBuilder.GetPlansForServiceForOrgReturns(nil, errors.New("Error fetching plans"))

					callUpdateService([]string{"-p", "flare", "some-stupid-not-real-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error fetching plans"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when the plan specified does not exist in the service offering", func() {
				It("returns an error", func() {
					callUpdateService([]string{"-p", "not-a-real-plan", "instance-without-service-offering"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Plan does not exist for the murkydb service"},
						[]string{"FAILED"},
					))
				})
			})
			Context("when there is an error updating the service instance", func() {
				It("returns an error", func() {
					serviceRepo.UpdateServiceInstanceReturns(errors.New("Error updating service instance"))
					callUpdateService([]string{"-p", "flare", "my-service-instance"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Error updating service instance"},
						[]string{"FAILED"},
					))
				})
			})
		})

	})
})
