package net_test

import (
	"code.google.com/p/go.net/websocket"
	"crypto/x509"
	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
	"net/http"
	"net/url"
	"syscall"
)

var _ = Describe("HTTP Client", func() {

	Describe("PrepareRedirect", func() {
		It("transfers authorization headers", func() {
			originalReq, err := http.NewRequest("GET", "/foo", nil)
			Expect(err).NotTo(HaveOccurred())
			originalReq.Header.Set("Authorization", "my-auth-token")

			redirectReq, err := http.NewRequest("GET", "/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			via := []*http.Request{originalReq}

			err = PrepareRedirect(redirectReq, via)

			Expect(err).NotTo(HaveOccurred())
			Expect(redirectReq.Header.Get("Authorization")).To(Equal("my-auth-token"))
		})

		It("fails after one redirect", func() {
			firstReq, err := http.NewRequest("GET", "/foo", nil)
			Expect(err).NotTo(HaveOccurred())

			secondReq, err := http.NewRequest("GET", "/manchu", nil)
			Expect(err).NotTo(HaveOccurred())

			redirectReq, err := http.NewRequest("GET", "/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			via := []*http.Request{firstReq, secondReq}

			err = PrepareRedirect(redirectReq, via)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("WrapNetworkErrors", func() {
		It("replaces http unknown authority errors with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &url.Error{Err: x509.UnknownAuthorityError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("replaces http hostname errors with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &url.Error{Err: x509.HostnameError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("replaces http certificate invalid errors with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &url.Error{Err: x509.CertificateInvalidError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("replaces websocket unknown authority errors with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &websocket.DialError{Err: x509.UnknownAuthorityError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("replaces websocket hostname with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &websocket.DialError{Err: x509.HostnameError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("replaces http websocket certificate invalid errors with InvalidSSLCert errors", func() {
			err, ok := WrapNetworkErrors("example.com", &websocket.DialError{Err: x509.CertificateInvalidError{}}).(*errors.InvalidSSLCert)
			Expect(ok).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})

		It("provides a nice message for connection errors", func() {
			underlyingErr := syscall.Errno(61)
			err := WrapNetworkErrors("example.com", &url.Error{Err: &net.OpError{Err: underlyingErr}})
			Expect(err).To(Equal(underlyingErr))
		})

		It("wraps other errors in a generic error type", func() {
			err := WrapNetworkErrors("example.com", errors.New("whatever"))
			Expect(err).To(HaveOccurred())

			_, ok := err.(*errors.InvalidSSLCert)
			Expect(ok).To(BeFalse())
		})
	})
})
