package domain_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		domainRepo          *testapi.FakeDomainRepository
		routingApiRepo      *testapi.FakeRoutingApiRepository
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetDomainRepository(domainRepo)
		deps.RepoLocator = deps.RepoLocator.SetRoutingApiRepository(routingApiRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("create-shared-domain").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		domainRepo = &testapi.FakeDomainRepository{}
		routingApiRepo = &testapi.FakeRoutingApiRepository{}
		configRepo = testconfig.NewRepositoryWithAccessToken(core_config.TokenInfo{Username: "my-user"})
	})

	runCommand := func(args ...string) bool {
		ui = new(testterm.FakeUI)
		return testcmd.RunCliCommand("create-shared-domain", args, requirementsFactory, updateCommandDependency, false)
	}

	It("TestShareDomainRequirements", func() {
		Expect(runCommand("example.com")).To(BeTrue())

		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false}

		Expect(runCommand("example.com")).To(BeFalse())
	})

	It("TestShareDomainFailsWithUsage", func() {
		runCommand()
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires an argument"},
		))

		runCommand("example.com")
		Expect(ui.Outputs).ToNot(ContainSubstrings(
			[]string{"Incorrect Usage", "Requires an argument"},
		))
	})

	It("TestShareDomain", func() {
		runCommand("example.com")

		Expect(domainRepo.CreateSharedDomainName).To(Equal("example.com"))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Creating shared domain", "example.com", "my-user"},
			[]string{"OK"},
		))
	})

	Context("when the cli is passed -r", func() {
		Context("when the router group exists", func() {
			var (
				routerGroupGuid string
				routerGroupName string
				routerGroupType string
			)

			BeforeEach(func() {
				routerGroupGuid = "router-group-guid"
				routerGroupName = "router-group-name"
				routerGroupType = "tcp"
				routerGroups := models.RouterGroups{
					models.RouterGroup{
						Name: routerGroupName,
						Guid: routerGroupGuid,
						Type: routerGroupType,
					},
				}

				routingApiRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) (apiErr error) {
					for _, r := range routerGroups {
						if !cb(r) {
							break
						}
					}
					return nil
				}
			})

			It("creates the domain", func() {
				runCommand("example.com", "-r", routerGroupName)

				Expect(routingApiRepo.ListRouterGroupsCallCount()).To(Equal(1))

				Expect(domainRepo.CreateSharedDomainName).To(Equal("example.com"))
				Expect(domainRepo.CreateSharedDomainRouterGroupGuid).To(Equal(routerGroupGuid))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Creating shared domain", "example.com", "my-user"},
					[]string{"OK"},
				))
			})
		})

		Context("when the router group does not exist", func() {
			It("fails with not found message", func() {
				runCommand("example.com", "-r", "does-not-exist")

				Expect(routingApiRepo.ListRouterGroupsCallCount()).To(Equal(1))

				// Expect it does not call CC
				Expect(domainRepo.CreateSharedDomainName).To(Equal(""))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"not found"},
				))
			})
		})

		Context("when an apiErr is received from ListRouterGroups", func() {
			BeforeEach(func() {
				routingApiRepo.ListRouterGroupsReturns(errors.New("BOOM"))
			})

			It("fails with the api err message", func() {
				runCommand("example.com", "-r", "does-not-exist")

				Expect(routingApiRepo.ListRouterGroupsCallCount()).To(Equal(1))

				// Expect it does not call CC
				Expect(domainRepo.CreateSharedDomainName).To(Equal(""))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"BOOM"},
				))
			})
		})
	})
})
