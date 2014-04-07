/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

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

var _ = Describe("delete-domain command", func() {
	It("fails requirements when not targeting an org", func() {
		domainRepo := &testapi.FakeDomainRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("fails requirements when not logged in", func() {
		domainRepo := &testapi.FakeDomainRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}

		callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("deletes domains", func() {
		domain := models.DomainFields{Name: "foo.com", Guid: "foo-guid"}
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"delete", "foo.com"},
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com", "my-user"},
			{"OK"},
		})
	})

	It("TestDeleteDomainNoConfirmation", func() {
		domain := models.DomainFields{Name: "foo.com", Guid: "foo-guid"}
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"no"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"delete", "foo.com"},
		})

		Expect(ui.Outputs).To(BeEmpty())
	})

	It("fails when the domain is not found", func() {
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgApiResponse: errors.NewModelNotFoundError("Domain", "foo.com"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"OK"},
			{"foo.com", "not found"},
		})
	})

	It("shows an error to the user when finding the domain fails", func() {
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgApiResponse: errors.New("failed badly"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"foo.com"},
			{"failed badly"},
		})
	})

	It("show the user an error when deleting the domain fails", func() {
		domain := models.DomainFields{Name: "foo.com", Guid: "foo-guid"}
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
			DeleteApiResponse:     errors.New("failed badly"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"FAILED"},
			{"foo.com"},
			{"failed badly"},
		})
	})

	It("skips confirmation when the -f flag is given", func() {
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		domain := models.DomainFields{Name: "foo.com", Guid: "foo-guid"}
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
		}
		ui := callDeleteDomain([]string{"-f", "foo.com"}, []string{}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))
		Expect(ui.Prompts).To(BeEmpty())
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"OK"},
		})
	})
})

func callDeleteDomain(args []string, inputs []string, reqFactory *testreq.FakeReqFactory, domainRepo *testapi.FakeDomainRepository) (ui *testterm.FakeUI) {
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

	cmd := domain.NewDeleteDomain(ui, configRepo, domainRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
