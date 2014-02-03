package requirements

import (
	"cf"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestUserReqExecute", func() {

			user := cf.UserFields{}
			user.Username = "my-user"
			user.Guid = "my-user-guid"

			userRepo := &testapi.FakeUserRepository{FindByUsernameUserFields: user}
			ui := new(testterm.FakeUI)

			userReq := newUserRequirement("foo", ui, userRepo)
			success := userReq.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), userRepo.FindByUsernameUsername, "foo")
			assert.Equal(mr.T(), userReq.GetUser(), user)
		})
		It("TestUserReqWhenUserDoesNotExist", func() {

			userRepo := &testapi.FakeUserRepository{FindByUsernameNotFound: true}
			ui := new(testterm.FakeUI)

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				newUserRequirement("foo", ui, userRepo).Execute()
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"UserFields not found"},
			})
		})
	})
}
