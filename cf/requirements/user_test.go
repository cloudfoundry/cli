package requirements_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"

	fakeapi "github.com/cloudfoundry/cli/cf/api/fakes"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserRequirement", func() {
	var (
		ui              *testterm.FakeUI
		userRepo        *fakeapi.FakeUserRepository
		userRequirement requirements.UserRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		userRepo = &fakeapi.FakeUserRepository{}
	})

	Describe("Execute", func() {
		Context("when wantGuid is true", func() {
			BeforeEach(func() {
				userRequirement = requirements.NewUserRequirement("the-username", ui, userRepo, true)
			})

			It("tries to find the user in CC", func() {
				userRequirement.Execute()
				Expect(userRepo.FindByUsernameCallCount()).To(Equal(1))
				Expect(userRepo.FindByUsernameArgsForCall(0)).To(Equal("the-username"))
			})

			Context("when the call to find the user succeeds", func() {
				var user models.UserFields
				BeforeEach(func() {
					user = models.UserFields{Username: "the-username", Guid: "the-guid"}
					userRepo.FindByUsernameReturns(user, nil)
				})

				It("stores the user that was found", func() {
					userRequirement.Execute()
					Expect(userRequirement.GetUser()).To(Equal(user))
				})

				It("returns true", func() {
					ok := userRequirement.Execute()
					Expect(ok).To(BeTrue())
				})
			})

			Context("when the call to find the user fails", func() {
				BeforeEach(func() {
					userRepo.FindByUsernameReturns(models.UserFields{}, errors.New("some error"))
				})

				It("panics and prints a failure message", func() {
					Expect(func() { userRequirement.Execute() }).To(Panic())
					Expect(ui.Outputs).To(ConsistOf([]string{"FAILED", "some error"}))
				})
			})
		})

		Context("when wantGuid is false", func() {
			BeforeEach(func() {
				userRequirement = requirements.NewUserRequirement("the-username", ui, userRepo, false)
			})

			It("does not try to find the user in CC", func() {
				userRequirement.Execute()
				Expect(userRepo.FindByUsernameCallCount()).To(Equal(0))
			})

			It("stores a user with just Username set", func() {
				userRequirement.Execute()
				expectedUser := models.UserFields{Username: "the-username"}
				Expect(userRequirement.GetUser()).To(Equal(expectedUser))
			})
		})
	})
})
