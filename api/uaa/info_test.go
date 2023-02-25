package uaa_test

import (
	. "code.cloudfoundry.org/cli/api/uaa"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	var info Info

	Describe("LoginLink", func() {
		BeforeEach(func() {
			info.Links.Login = "login-something"
		})

		It("returns the Login Link", func() {
			Expect(info.LoginLink()).To(Equal("login-something"))
		})
	})

	Describe("UAALink", func() {
		BeforeEach(func() {
			info.Links.UAA = "uaa-something"
		})

		It("returns the UAA Link", func() {
			Expect(info.UAALink()).To(Equal("uaa-something"))
		})
	})

	Describe("NewInfo", func() {
		When("provided a default link", func() {
			It("sets the links to the provided link", func() {
				info = NewInfo("uaa-url", "auth-url")
				Expect(info.LoginLink()).To(Equal("auth-url"))
				Expect(info.UAALink()).To(Equal("uaa-url"))
			})
		})
	})
})
