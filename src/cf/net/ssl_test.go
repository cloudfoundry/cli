package net_test

import (
	"cf/errors"
	. "cf/net"
	"crypto/x509"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/url"
)

var _ = Describe("SSL", func() {
	Describe("WrapSSLErrors", func() {
		var (
			err  error
			host string
		)

		BeforeEach(func() {
			host = "example.com"
		})

		Context("when provided url Error", func() {
			Context("and with x509 Unknown Authority error", func() {
				BeforeEach(func() {
					err = &url.Error{Err: x509.UnknownAuthorityError{}}
				})

				It("returns an InvalidSSLCert", func() {
					_, ok := WrapSSLErrors(host, err).(*errors.InvalidSSLCert)
					Expect(ok).To(BeTrue())
				})
			})
		})
		Context("when provided Websocket DialError", func() {
			Context("and with x509 hostname error", func() {
				It("returns an InvalidSSLCert", func() {

				})
			})
		})
	})
})
