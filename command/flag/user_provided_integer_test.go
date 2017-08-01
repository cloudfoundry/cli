package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserProvidedInteger", func() {
	var userProvidedInt UserProvidedInteger

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {

		})

		Context("when the user provides a non-zero int", func() {
			It("stores the non-zero integer", func() {

			})
		})

		Context("when the user provides 0", func() {
			It("stores 0", func() {

			})
		})

		Context("when the user does not provide a value", func() {
			It("does not store a value, and sets UserGiven to false", func() {

			})
		})
	})
})
