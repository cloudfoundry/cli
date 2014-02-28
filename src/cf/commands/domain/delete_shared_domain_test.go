package domain_test

import (
	"cf/commands/domain"
	"cf/configuration"
	"cf/errors"
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

var _ = Describe("delete-shared-domain command", func() {
	It("TestGetDeleteSharedDomainRequirements", func() {
		deps := getDeleteSharedDomainDeps()
		deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		deps.requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestDeleteSharedDomainNotFound", func() {

		deps := getDeleteSharedDomainDeps()
		deps.domainRepo.FindByNameInOrgApiResponse = errors.NewNotFoundError("%s %s not found", "Domain", "foo.com")
		ui := callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)

		Expect(deps.domainRepo.DeleteDomainGuid).To(Equal(""))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"OK"},
			{"foo.com", "not found"},
		})
	})
	It("TestDeleteSharedDomainFindError", func() {

		deps := getDeleteSharedDomainDeps()
		deps.domainRepo.FindByNameInOrgApiResponse = errors.NewErrorWithMessage("couldn't find the droids you're lookin for")
		ui := callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)

		Expect(deps.domainRepo.DeleteDomainGuid).To(Equal(""))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"FAILED"},
			{"foo.com"},
			{"couldn't find the droids you're lookin for"},
		})
	})
	It("TestDeleteSharedDomainDeleteError", func() {

		deps := getDeleteSharedDomainDeps()
		deps.domainRepo.DeleteSharedApiResponse = errors.NewErrorWithMessage("failed badly")
		ui := callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)

		Expect(deps.domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"FAILED"},
			{"foo.com"},
			{"failed badly"},
		})
	})
	It("TestDeleteSharedDomainHasConfirmation", func() {

		deps := getDeleteSharedDomainDeps()
		ui := callDeleteSharedDomain([]string{"foo.com"}, []string{"y"}, deps)

		Expect(deps.domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"shared", "foo.com"},
		})
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"OK"},
		})
	})
	It("TestDeleteSharedDomainForceFlagSkipsConfirmation", func() {

		deps := getDeleteSharedDomainDeps()
		ui := callDeleteSharedDomain([]string{"-f", "foo.com"}, []string{}, deps)

		Expect(deps.domainRepo.DeleteSharedDomainGuid).To(Equal("foo-guid"))
		Expect(len(ui.Prompts)).To(Equal(0))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"OK"},
		})
	})
})

func fakeDomainRepo() *testapi.FakeDomainRepository {
	domain := models.DomainFields{}
	domain.Name = "foo.com"
	domain.Guid = "foo-guid"
	domain.Shared = true

	return &testapi.FakeDomainRepository{
		FindByNameInOrgDomain: domain,
	}
}

type deleteSharedDomainDependencies struct {
	requirementsFactory *testreq.FakeReqFactory
	domainRepo          *testapi.FakeDomainRepository
}

func getDeleteSharedDomainDeps() deleteSharedDomainDependencies {
	return deleteSharedDomainDependencies{
		requirementsFactory: &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true},
		domainRepo:          fakeDomainRepo(),
	}
}

func callDeleteSharedDomain(args []string, inputs []string, deps deleteSharedDomainDependencies) (ui *testterm.FakeUI) {
	ctxt := testcmd.NewContext("delete-domain", args)
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

	configRepo := testconfig.NewRepositoryWithAccessToken(configuration.TokenInfo{Username: "my-user"})

	spaceFields := models.SpaceFields{}
	spaceFields.Name = "my-space"

	orgFields := models.OrganizationFields{}
	orgFields.Name = "my-org"
	configRepo.SetSpaceFields(spaceFields)
	configRepo.SetOrganizationFields(orgFields)

	cmd := domain.NewDeleteSharedDomain(ui, configRepo, deps.domainRepo)
	testcmd.RunCommand(cmd, ctxt, deps.requirementsFactory)
	return
}
