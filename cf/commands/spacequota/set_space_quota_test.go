package spacequota_test

import (
	"github.com/cloudfoundry/cli/cf/api/spacequotas/spacequotasfakes"
	"github.com/cloudfoundry/cli/cf/api/spaces/spacesfakes"
	"github.com/cloudfoundry/cli/cf/commands/spacequota"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig/coreconfigfakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/flags"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/commandregistry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-space-quota command", func() {
	var (
		ui                  *testterm.FakeUI
		spaceRepo           *spacesfakes.FakeSpaceRepository
		quotaRepo           *spacequotasfakes.FakeSpaceQuotaRepository
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          *coreconfigfakes.FakeRepository
		deps                commandregistry.Dependency
		cmd                 spacequota.SetSpaceQuota
		flagContext         flags.FlagContext
		loginReq            requirements.Requirement
		orgReq              *requirementsfakes.FakeTargetedOrgRequirement
	)

	BeforeEach(func() {
		requirementsFactory = new(requirementsfakes.FakeFactory)

		loginReq = requirements.Passing{Type: "login"}
		requirementsFactory.NewLoginRequirementReturns(loginReq)
		orgReq = new(requirementsfakes.FakeTargetedOrgRequirement)
		requirementsFactory.NewTargetedOrgRequirementReturns(orgReq)

		ui = new(testterm.FakeUI)
		configRepo = new(coreconfigfakes.FakeRepository)
		deps = commandregistry.Dependency{
			UI:     ui,
			Config: configRepo,
		}
		quotaRepo = new(spacequotasfakes.FakeSpaceQuotaRepository)
		deps.RepoLocator = deps.RepoLocator.SetSpaceQuotaRepository(quotaRepo)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)

		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)

		cmd = spacequota.SetSpaceQuota{}
		cmd.SetDependency(deps, false)

		configRepo.UsernameReturns("my-user")
	})

	Describe("Requirements", func() {
		Context("when provided a quota and space", func() {
			var reqs []requirements.Requirement

			BeforeEach(func() {
				flagContext.Parse("space", "space-quota")
				reqs = cmd.Requirements(requirementsFactory, flagContext)
			})

			It("returns a LoginRequirement", func() {
				Expect(reqs).To(ContainElement(loginReq))
			})

			It("requires the user to target an org", func() {
				Expect(reqs).To(ContainElement(orgReq))
			})
		})

		Context("when not provided a quota and space", func() {
			BeforeEach(func() {
				flagContext.Parse("")
			})

			It("fails with usage", func() {
				Expect(func() { cmd.Requirements(requirementsFactory, flagContext) }).To(Panic())
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Incorrect Usage. Requires", "as arguments"},
				))
			})
		})
	})

	Describe("Execute", func() {
		var executeErr error

		JustBeforeEach(func() {
			flagContext.Parse("my-space", "quota-name")
			executeErr = cmd.Execute(flagContext)
		})

		Context("when the space and quota both exist", func() {
			BeforeEach(func() {
				quotaRepo.FindByNameReturns(
					models.SpaceQuota{
						Name:                    "quota-name",
						GUID:                    "quota-guid",
						MemoryLimit:             1024,
						InstanceMemoryLimit:     512,
						RoutesLimit:             111,
						ServicesLimit:           222,
						NonBasicServicesAllowed: true,
						OrgGUID:                 "my-org-guid",
					}, nil)

				spaceRepo.FindByNameReturns(
					models.Space{
						SpaceFields: models.SpaceFields{
							Name: "my-space",
							GUID: "my-space-guid",
						},
						SpaceQuotaGUID: "",
					}, nil)
			})

			Context("when the space quota was not previously assigned to a space", func() {
				It("associates the provided space with the provided space quota", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					spaceGUID, quotaGUID := quotaRepo.AssociateSpaceWithQuotaArgsForCall(0)

					Expect(spaceGUID).To(Equal("my-space-guid"))
					Expect(quotaGUID).To(Equal("quota-guid"))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Assigning space quota", "to space", "my-user"},
						[]string{"OK"},
					))
				})
			})

			Context("when the space quota was previously assigned to a space", func() {
				BeforeEach(func() {
					spaceRepo.FindByNameReturns(
						models.Space{
							SpaceFields: models.SpaceFields{
								Name: "my-space",
								GUID: "my-space-guid",
							},
							SpaceQuotaGUID: "another-quota",
						}, nil)
				})

				It("warns the user that the operation was not performed", func() {
					Expect(quotaRepo.UpdateCallCount()).To(Equal(0))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Assigning space quota", "to space", "my-user"},
					))
					Expect(executeErr).To(HaveOccurred())
					Expect(executeErr.Error()).To(Equal("This space already has an assigned space quota."))
				})
			})
		})

		Context("when an error occurs fetching the space", func() {
			var spaceError error

			BeforeEach(func() {
				spaceError = errors.New("space-repo-err")
				spaceRepo.FindByNameReturns(models.Space{}, spaceError)
			})

			It("prints an error", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Assigning space quota", "to space", "my-user"},
				))
				Expect(executeErr).To(Equal(spaceError))
			})
		})

		Context("when an error occurs fetching the quota", func() {
			var quotaErr error

			BeforeEach(func() {
				spaceRepo.FindByNameReturns(
					models.Space{
						SpaceFields: models.SpaceFields{
							Name: "my-space",
							GUID: "my-space-guid",
						},
						SpaceQuotaGUID: "",
					}, nil)

				quotaErr = errors.New("I can't find my quota name!")
				quotaRepo.FindByNameReturns(models.SpaceQuota{}, quotaErr)
			})

			It("prints an error", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Assigning space quota", "to space", "my-user"},
				))
				Expect(executeErr).To(Equal(quotaErr))
			})
		})
	})
})
