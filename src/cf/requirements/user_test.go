package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestUserReqExecute", func() {
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

	It("TestUserReqWhenUserDoesNotExist", func() {
		userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
		ui := new(testterm.FakeUI)

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			NewUserRequirement("foo", ui, userRepo).Execute()
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"User not found"},
		})
	})
})
