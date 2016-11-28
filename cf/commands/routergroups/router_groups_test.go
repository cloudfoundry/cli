package routergroups_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commands/routergroups"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouterGroups", func() {

	var (
		ui             *testterm.FakeUI
		routingAPIRepo *apifakes.FakeRoutingAPIRepository
		deps           commandregistry.Dependency
		cmd            *routergroups.RouterGroups
		flagContext    flags.FlagContext
		repoLocator    api.RepositoryLocator
		config         coreconfig.Repository

		requirementsFactory           *requirementsfakes.FakeFactory
		minAPIVersionRequirement      *requirementsfakes.FakeRequirement
		loginRequirement              *requirementsfakes.FakeRequirement
		routingAPIEndpoingRequirement *requirementsfakes.FakeRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		routingAPIRepo = new(apifakes.FakeRoutingAPIRepository)
		config = testconfig.NewRepositoryWithDefaults()
		repoLocator = api.RepositoryLocator{}.SetRoutingAPIRepository(routingAPIRepo)
		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      config,
			RepoLocator: repoLocator,
		}

		minAPIVersionRequirement = new(requirementsfakes.FakeRequirement)
		loginRequirement = new(requirementsfakes.FakeRequirement)
		routingAPIEndpoingRequirement = new(requirementsfakes.FakeRequirement)

		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewMinAPIVersionRequirementReturns(minAPIVersionRequirement)
		requirementsFactory.NewLoginRequirementReturns(loginRequirement)
		requirementsFactory.NewRoutingAPIRequirementReturns(routingAPIEndpoingRequirement)

		cmd = new(routergroups.RouterGroups)
		cmd = cmd.SetDependency(deps, false).(*routergroups.RouterGroups)
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	Describe("Requirements", func() {
		It("fails if the user is not logged in", func() {
			cmd.Requirements(requirementsFactory, flagContext)

			Expect(requirementsFactory.NewLoginRequirementCallCount()).To(Equal(1))
		})

		It("fails when the routing API endpoint is not set", func() {
			cmd.Requirements(requirementsFactory, flagContext)

			Expect(requirementsFactory.NewRoutingAPIRequirementCallCount()).To(Equal(1))
		})

		It("should fail with usage", func() {
			flagContext.Parse("blahblah")
			cmd.Requirements(requirementsFactory, flagContext)

			Expect(requirementsFactory.NewUsageRequirementCallCount()).To(Equal(1))
		})
	})

	Describe("Execute", func() {
		var err error

		BeforeEach(func() {
			err := flagContext.Parse()
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			err = cmd.Execute(flagContext)
		})

		Context("when there are router groups", func() {
			BeforeEach(func() {
				routerGroups := models.RouterGroups{
					models.RouterGroup{
						GUID: "guid-0001",
						Name: "default-router-group",
						Type: "tcp",
					},
				}
				routingAPIRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) (apiErr error) {
					for _, r := range routerGroups {
						if !cb(r) {
							break
						}
					}
					return nil
				}
			})

			It("lists router groups", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting router groups", "my-user"},
					[]string{"name", "type"},
					[]string{"default-router-group", "tcp"},
				))
			})
		})

		Context("when there are no router groups", func() {
			It("tells the user when no router groups were found", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting router groups"},
					[]string{"No router groups found"},
				))
			})
		})

		Context("when there is an error listing router groups", func() {
			BeforeEach(func() {
				routingAPIRepo.ListRouterGroupsReturns(errors.New("BOOM"))
			})

			It("returns an error to the user", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Getting router groups"},
				))

				Expect(err).To(HaveOccurred())
				errStr := err.Error()
				Expect(errStr).To(ContainSubstring("BOOM"))
				Expect(errStr).To(ContainSubstring("Failed fetching router groups"))
			})
		})
	})
})
