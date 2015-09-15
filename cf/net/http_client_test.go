package net_test

import (
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"syscall"

	"github.com/cloudfoundry/cli/cf/errors"
	. "github.com/cloudfoundry/cli/cf/net"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/websocket"
)

var _ = Describe("HTTP Client", func() {

	Describe("PrepareRedirect", func() {
		It("transfers original headers", func() {
			originalReq, err := http.NewRequest("GET", "http://local.com/foo", nil)
			Expect(err).NotTo(HaveOccurred())
			originalReq.Header.Set("Authorization", "my-auth-token")
			originalReq.Header.Set("Accept", "application/json")

			redirectReq, err := http.NewRequest("GET", "http://local.com/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			via := []*http.Request{originalReq}

			err = PrepareRedirect(redirectReq, via)

			Expect(err).NotTo(HaveOccurred())
			Expect(redirectReq.Header.Get("Authorization")).To(Equal("my-auth-token"))
			Expect(redirectReq.Header.Get("Accept")).To(Equal("application/json"))
		})

		It("does not transfer 'Authorization' headers during a redirect to different Host", func() {
			originalReq, err := http.NewRequest("GET", "http://www.local.com/foo", nil)
			Expect(err).NotTo(HaveOccurred())
			originalReq.Header.Set("Authorization", "my-auth-token")
			originalReq.Header.Set("Accept", "application/json")

			redirectReq, err := http.NewRequest("GET", "http://www.remote.com/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			via := []*http.Request{originalReq}

			err = PrepareRedirect(redirectReq, via)

			Expect(err).NotTo(HaveOccurred())
			Expect(redirectReq.Header.Get("Authorization")).To(Equal(""))
			Expect(redirectReq.Header.Get("Accept")).To(Equal("application/json"))
		})

		It("does not transfer POST-specific headers", func() {
			originalReq, err := http.NewRequest("POST", "http://local.com/foo", nil)
			Expect(err).NotTo(HaveOccurred())
			originalReq.Header.Set("Content-Type", "application/json")
			originalReq.Header.Set("Content-Length", "100")

			redirectReq, err := http.NewRequest("GET", "http://local.com/bar", nil)
			Expect(err).NotTo(HaveOccurred())

			via := []*http.Request{originalReq}

			err = PrepareRedirect(redirectReq, via)

			Expect(err).NotTo(HaveOccurred())
			Expect(redirectReq.Header.Get("Content-Type")).To(Equal(""))
			Expect(redirectReq.Header.Get("Content-Length")).To(Equal(""))
		})

		It("fails after one redirect", func() {
			firstReq, err := http.NewRequest("GET", "http://local.com/foo", nil)
			Expect(err).NotTo(HaveOccurred())

			secondReq, err := http.NewRequest("GET", "http://local.com/manchu", nil)
			Expect(err).NotTo(HaveOccurred())

			redirectReq, err := http.NewRequest("GET", "http://local.com/bar", nil)
			redirectReq.Header["Referer"] = []string{"http://local.com"}
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
			Expect(err.Error()).To(ContainSubstring("Error performing request"))
		})

		It("wraps other errors in a generic error type", func() {
			err := WrapNetworkErrors("example.com", errors.New("whatever"))
			Expect(err).To(HaveOccurred())

			_, ok := err.(*errors.InvalidSSLCert)
			Expect(ok).To(BeFalse())
		})
	})
})
