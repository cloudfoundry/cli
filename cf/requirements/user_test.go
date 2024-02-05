package requirements_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"

	"code.cloudfoundry.org/cli/cf/api/apifakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserRequirement", func() {
	var (
		userRepo        *apifakes.FakeUserRepository
		userRequirement requirements.UserRequirement
	)

	BeforeEach(func() {
		userRepo = new(apifakes.FakeUserRepository)
	})

	Describe("Execute", func() {
		Context("when wantGUID is true", func() {
			BeforeEach(func() {
				userRequirement = requirements.NewUserRequirement("the-username", userRepo, true)
			})

			It("tries to find the user in CC", func() {
				userRequirement.Execute()
				Expect(userRepo.FindByUsernameCallCount()).To(Equal(1))
				Expect(userRepo.FindByUsernameArgsForCall(0)).To(Equal("the-username"))
			})

			Context("when the call to find the user succeeds", func() {
				var user models.UserFields
				BeforeEach(func() {
					user = models.UserFields{Username: "the-username", GUID: "the-guid"}
					userRepo.FindByUsernameReturns(user, nil)
				})

				It("stores the user that was found", func() {
					userRequirement.Execute()
					Expect(userRequirement.GetUser()).To(Equal(user))
				})

				It("does not error", func() {
					err := userRequirement.Execute()
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the call to find the user fails", func() {
				userError := errors.New("some error")
				BeforeEach(func() {
					userRepo.FindByUsernameReturns(models.UserFields{}, userError)
				})

				It("errors", func() {
					err := userRequirement.Execute()

					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(userError))
				})
			})
		})

		Context("when wantGUID is false", func() {
			BeforeEach(func() {
				userRequirement = requirements.NewUserRequirement("the-username", userRepo, false)
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

	Describe("NewClientRequirement", func() {
		var (
			clientRepo *apifakes.FakeClientRepository
		)

		BeforeEach(func() {
			clientRepo = new(apifakes.FakeClientRepository)
		})
		Context("Execute", func() {
			It("returns a user model with the client ID as its GUID and Username", func() {
				userReq := requirements.NewClientRequirement("the-client-id", clientRepo)

				userReq.Execute()
				user := userReq.GetUser()

				Expect(user.GUID).To(Equal("the-client-id"))
				Expect(user.Username).To(Equal("the-client-id"))
			})

			When("The client is not found by the clientRepo", func() {
				var (
					clientReq requirements.UserRequirement
					err       error
				)

				BeforeEach(func() {
					err = errors.New("Client not found")
					clientRepo.ClientExistsReturns(false, err)
				})

				It("returns an error", func() {
					clientReq = requirements.NewClientRequirement("the-client-id", clientRepo)
					output := clientReq.Execute()

					Expect(clientRepo.ClientExistsCallCount()).To(Equal(1))
					Expect(clientRepo.ClientExistsArgsForCall(0)).To(Equal("the-client-id"))
					Expect(output).To(Equal(err))
				})
			})
		})
	})
})
