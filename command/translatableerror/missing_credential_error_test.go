package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MissingCredentialsError", func() {
	Describe("Error", func() {
		When("username was not provided", func() {
			It("prints a message asking for username", func() {
				err := MissingCredentialsError{
					MissingUsername: true,
				}

				Expect(err.Error()).To(Equal("Username not provided."))
			})
		})

		When("password was not provided", func() {
			It("prints a message asking for username", func() {
				err := MissingCredentialsError{
					MissingPassword: true,
				}

				Expect(err.Error()).To(Equal("Password not provided."))
			})
		})

		When("neither username nor password was provided", func() {
			It("prints a message asking for username", func() {
				err := MissingCredentialsError{
					MissingUsername: true,
					MissingPassword: true,
				}

				Expect(err.Error()).To(Equal("Username and password not provided."))
			})
		})
	})
})
