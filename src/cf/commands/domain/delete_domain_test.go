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
	It("TestGetRequirements", func() {
		domainRepo := &testapi.FakeDomainRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: false}
		callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())

		reqFactory = &testreq.FakeReqFactory{LoginSuccess: false, TargetedOrgSuccess: true}
		callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestDeleteDomainSuccess", func() {

		domain := models.DomainFields{}
		domain.Name = "foo.com"
		domain.Guid = "foo-guid"
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

		domain := models.DomainFields{}
		domain.Name = "foo.com"
		domain.Guid = "foo-guid"
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"no"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Prompts, testassert.Lines{
			{"delete", "foo.com"},
		})

		Expect(len(ui.Outputs)).To(Equal(1))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
		})
	})
	It("TestDeleteDomainNotFound", func() {

		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgApiResponse: errors.NewModelNotFoundError("Domain", "foo.com"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"OK"},
			{"foo.com", "not found"},
		})
	})
	It("TestDeleteDomainFindError", func() {

		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgApiResponse: errors.New("failed badly"),
		}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		ui := callDeleteDomain([]string{"foo.com"}, []string{"y"}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal(""))

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Deleting domain", "foo.com"},
			{"FAILED"},
			{"foo.com"},
			{"failed badly"},
		})
	})
	It("TestDeleteDomainDeleteError", func() {

		domain := models.DomainFields{}
		domain.Name = "foo.com"
		domain.Guid = "foo-guid"
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
	It("TestDeleteDomainForceFlagSkipsConfirmation", func() {

		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedOrgSuccess: true}

		domain := models.DomainFields{}
		domain.Name = "foo.com"
		domain.Guid = "foo-guid"
		domainRepo := &testapi.FakeDomainRepository{
			FindByNameInOrgDomain: domain,
		}
		ui := callDeleteDomain([]string{"-f", "foo.com"}, []string{}, reqFactory, domainRepo)

		Expect(domainRepo.DeleteDomainGuid).To(Equal("foo-guid"))
		Expect(len(ui.Prompts)).To(Equal(0))
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
