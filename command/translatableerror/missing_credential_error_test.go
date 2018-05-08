package translatableerror_test

import (
	. "code.cloudfoundry.org/cli/command/translatableerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MissingCredentialsError", func() {
	Describe("Error", func() {
		Context("when username was not provided", func() {
			It("prints a message asking for username", func() {
				err := MissingCredentialsError{
					MissingUsername: true,
				}

				Expect(err.Error()).To(Equal("Username not provided."))
			})
		})

		Context("when password was not provided", func() {
			It("prints a message asking for username", func() {
				err := MissingCredentialsError{
					MissingPassword: true,
				}

				Expect(err.Error()).To(Equal("Password not provided."))
			})
		})

		Context("when neither username nor password was provided", func() {
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
