package commands_test

import (
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("target command", func() {
	var (
		orgRepo    *testapi.FakeOrgRepository
		spaceRepo  *testapi.FakeSpaceRepository
		config     configuration.ReadWriter
		reqFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		orgRepo, spaceRepo, config, reqFactory = getTargetDependencies()
	})

	It("TestTargetFailsWithUsage", func() {
		ui := callTarget([]string{"bad-foo"}, reqFactory, config, orgRepo, spaceRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when targeting a space or org", func() {
		callTarget([]string{"-o", "some-crazy-org-im-not-in"}, reqFactory, config, orgRepo, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		callTarget([]string{"-s", "i-love-space"}, reqFactory, config, orgRepo, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("passes requirements when not attempting to target a space or org", func() {
		callTarget([]string{}, reqFactory, config, orgRepo, spaceRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
	})

	Context("when the user logs in successfully", func() {
		BeforeEach(func() {
			reqFactory.LoginSuccess = true
		})

		It("passes requirements when targeting a space or org", func() {
			callTarget([]string{"-s", "i-love-space"}, reqFactory, config, orgRepo, spaceRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

			callTarget([]string{"-o", "orgs-are-delightful"}, reqFactory, config, orgRepo, spaceRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("TestTargetOrganizationWhenUserHasAccess", func() {
			org := models.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"

			orgRepo.Organizations = []models.Organization{org}
			orgRepo.FindByNameOrganization = org

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)

			Expect(orgRepo.FindByNameName).To(Equal("my-organization"))
			Expect(ui.ShowConfigurationCalled).To(BeTrue())

			Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))
		})

		It("TestTargetOrganizationWhenUserDoesNotHaveAccess", func() {
			orgs := []models.Organization{}
			orgRepo.Organizations = orgs
			orgRepo.FindByNameErr = true

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)
			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{{"FAILED"}})
		})

		It("TestTargetOrganizationWhenOrgNotFound", func() {
			orgRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-o", "my-organization"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-organization", "not found"},
			})
		})

		It("TestTargetSpaceWhenNoOrganizationIsSelected", func() {
			config.SetOrganizationFields(models.OrganizationFields{})

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"An org must be targeted before targeting a space"},
			})
			Expect(config.OrganizationFields().Guid).To(Equal(""))
		})

		It("TestTargetSpaceWhenUserHasAccess", func() {
			space := models.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"

			spaceRepo.Spaces = []models.Space{space}
			spaceRepo.FindByNameSpace = space

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
			Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))
			Expect(ui.ShowConfigurationCalled).To(BeTrue())
		})

		It("TestTargetSpaceWhenUserDoesNotHaveAccess", func() {
			config.SetSpaceFields(models.SpaceFields{})
			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})

			Expect(config.SpaceFields().Guid).To(Equal(""))
			Expect(ui.ShowConfigurationCalled).To(BeFalse())
		})

		It("TestTargetSpaceWhenSpaceNotFound", func() {
			spaceRepo.FindByNameNotFound = true

			ui := callTarget([]string{"-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"my-space", "not found"},
			})
		})

		It("TestTargetOrganizationAndSpace", func() {
			org := models.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.Organizations = []models.Organization{org}

			space := models.Space{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			spaceRepo.Spaces = []models.Space{space}

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			Expect(ui.ShowConfigurationCalled).To(BeTrue())
			Expect(orgRepo.FindByNameName).To(Equal("my-organization"))
			Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))
			Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
			Expect(config.SpaceFields().Guid).To(Equal("my-space-guid"))
		})

		It("TestTargetOrganizationAndSpaceWhenSpaceFails", func() {
			config.SetSpaceFields(models.SpaceFields{})

			org := models.Organization{}
			org.Name = "my-organization"
			org.Guid = "my-organization-guid"
			orgRepo.Organizations = []models.Organization{org}

			spaceRepo.FindByNameErr = true

			ui := callTarget([]string{"-o", "my-organization", "-s", "my-space"}, reqFactory, config, orgRepo, spaceRepo)

			Expect(ui.ShowConfigurationCalled).To(BeFalse())
			Expect(orgRepo.FindByNameName).To(Equal("my-organization"))
			Expect(config.OrganizationFields().Guid).To(Equal("my-organization-guid"))
			Expect(spaceRepo.FindByNameName).To(Equal("my-space"))
			Expect(config.SpaceFields().Guid).To(Equal(""))
			testassert.SliceContains(GinkgoT(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"Unable to access space", "my-space"},
			})
		})
	})
})

func getTargetDependencies() (
	orgRepo *testapi.FakeOrgRepository,
	spaceRepo *testapi.FakeSpaceRepository,
	config configuration.ReadWriter,
	reqFactory *testreq.FakeReqFactory) {

	orgRepo = &testapi.FakeOrgRepository{}
	spaceRepo = &testapi.FakeSpaceRepository{}
	config = testconfig.NewRepositoryWithDefaults()

	reqFactory = &testreq.FakeReqFactory{}
	return
}

func callTarget(args []string,
	reqFactory *testreq.FakeReqFactory,
	config configuration.ReadWriter,
	orgRepo api.OrganizationRepository,
	spaceRepo api.SpaceRepository) (ui *testterm.FakeUI) {

	ui = new(testterm.FakeUI)
	cmd := NewTarget(ui, config, orgRepo, spaceRepo)
	ctxt := testcmd.NewContext("target", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
