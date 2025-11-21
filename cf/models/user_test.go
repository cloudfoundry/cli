package models_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User", func() {
	Describe("UserFields", func() {
		It("stores user information", func() {
			user := models.UserFields{
				Guid:     "user-guid",
				Username: "john.doe@example.com",
				Password: "secret-password",
				IsAdmin:  true,
			}

			Expect(user.Guid).To(Equal("user-guid"))
			Expect(user.Username).To(Equal("john.doe@example.com"))
			Expect(user.Password).To(Equal("secret-password"))
			Expect(user.IsAdmin).To(BeTrue())
		})

		It("handles non-admin user", func() {
			user := models.UserFields{
				Guid:     "user-guid",
				Username: "regular.user@example.com",
				IsAdmin:  false,
			}

			Expect(user.IsAdmin).To(BeFalse())
		})

		It("handles empty password", func() {
			user := models.UserFields{
				Guid:     "user-guid",
				Username: "user@example.com",
				Password: "",
			}

			Expect(user.Password).To(BeEmpty())
		})

		It("handles empty values", func() {
			user := models.UserFields{}

			Expect(user.Guid).To(BeEmpty())
			Expect(user.Username).To(BeEmpty())
			Expect(user.Password).To(BeEmpty())
			Expect(user.IsAdmin).To(BeFalse())
		})

		It("stores different username formats", func() {
			user1 := models.UserFields{
				Username: "email@example.com",
			}
			user2 := models.UserFields{
				Username: "username",
			}
			user3 := models.UserFields{
				Username: "user-with-dashes",
			}

			Expect(user1.Username).To(Equal("email@example.com"))
			Expect(user2.Username).To(Equal("username"))
			Expect(user3.Username).To(Equal("user-with-dashes"))
		})

		It("can change admin status", func() {
			user := models.UserFields{
				Guid:    "user-guid",
				IsAdmin: false,
			}

			Expect(user.IsAdmin).To(BeFalse())

			user.IsAdmin = true
			Expect(user.IsAdmin).To(BeTrue())
		})
	})
})
