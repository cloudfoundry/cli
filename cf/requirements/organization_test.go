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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestOrgReqExecute", func() {

		org := models.Organization{}
		org.Name = "my-org-name"
		org.Guid = "my-org-guid"
		orgRepo := &testapi.FakeOrgRepository{Organizations: []models.Organization{org}}
		ui := new(testterm.FakeUI)

		orgReq := NewOrganizationRequirement("my-org-name", ui, orgRepo)
		success := orgReq.Execute()

		Expect(success).To(BeTrue())
		Expect(orgRepo.FindByNameName).To(Equal("my-org-name"))
		Expect(orgReq.GetOrganization()).To(Equal(org))
	})

	It("TestOrgReqWhenOrgDoesNotExist", func() {

		orgRepo := &testapi.FakeOrgRepository{FindByNameNotFound: true}
		ui := new(testterm.FakeUI)

		orgReq := NewOrganizationRequirement("foo", ui, orgRepo)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			orgReq.Execute()
		})
	})
})
