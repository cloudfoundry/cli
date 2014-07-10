package requirements_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("UserRequirement", func() {
	Context("when a user with the given name can be found", func() {
		It("returns the user model", func() {
			user := models.UserFields{}
			user.Username = "my-user"
			user.Guid = "my-user-guid"

			userRepo := &testapi.FakeUserRepository{FindByUsernameUserFields: user}
			ui := new(testterm.FakeUI)

			userReq := NewUserRequirement("foo", ui, userRepo)
			success := userReq.Execute()

			Expect(success).To(BeTrue())
			Expect(userRepo.FindByUsernameUsername).To(Equal("foo"))
			Expect(userReq.GetUser()).To(Equal(user))
		})
	})

	Context("when a user with the given name cannot be found", func() {
		It("panics and prints a failure message", func() {
			userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
			ui := new(testterm.FakeUI)

			testassert.AssertPanic(testterm.QuietPanic, func() {
				NewUserRequirement("foo", ui, userRepo).Execute()
			})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"not found"},
			))
		})
	})
})
