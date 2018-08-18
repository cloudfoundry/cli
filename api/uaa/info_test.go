package uaa_test

import (
	. "code.cloudfoundry.org/cli/api/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Info", func() {
	var info Info

	Describe("APIVersion", func() {
		BeforeEach(func() {
			info.App.Version = "api-version"
		})

		It("returns the version", func() {
			Expect(info.APIVersion()).To(Equal("api-version"))
		})
	})

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
				link := "something-else-i-don't know"
				info = NewInfo(link)
				Expect(info.LoginLink()).To(Equal(link))
				Expect(info.UAALink()).To(Equal(link))
			})
		})
	})
})
