package authenticators_test

import (
	"regexp"

	"code.cloudfoundry.org/diego-ssh/authenticators"
	"code.cloudfoundry.org/diego-ssh/authenticators/fake_authenticators"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"code.cloudfoundry.org/lager/lagertest"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiegoProxyAuthenticator", func() {
	var (
		logger             *lagertest.TestLogger
		credentials        []byte
		permissionsBuilder *fake_authenticators.FakePermissionsBuilder
		authenticator      *authenticators.DiegoProxyAuthenticator
		metadata           *fake_ssh.FakeConnMetadata
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		credentials = []byte("some-user:some-password")
		permissionsBuilder = &fake_authenticators.FakePermissionsBuilder{}
		permissionsBuilder.BuildReturns(&ssh.Permissions{}, nil)
		authenticator = authenticators.NewDiegoProxyAuthenticator(logger, credentials, permissionsBuilder)

		metadata = &fake_ssh.FakeConnMetadata{}
	})

	Describe("Authenticate", func() {
		var (
			password []byte
			authErr  error
		)

		BeforeEach(func() {
			password = []byte{}
		})

		JustBeforeEach(func() {
			_, authErr = authenticator.Authenticate(metadata, password)
		})

		Context("when the user name matches the user regex and valid credentials are provided", func() {
			BeforeEach(func() {
				metadata.UserReturns("diego:some-guid/0")
				password = []byte("some-user:some-password")
			})

			It("authenticates the password against the provided user:password", func() {
				Expect(authErr).NotTo(HaveOccurred())
			})

			It("builds permissions for the requested process", func() {
				Expect(permissionsBuilder.BuildCallCount()).To(Equal(1))
				_, guid, index, metadata := permissionsBuilder.BuildArgsForCall(0)
				Expect(guid).To(Equal("some-guid"))
				Expect(index).To(Equal(0))
				Expect(metadata).To(Equal(metadata))
			})
		})

		Context("when the user name doesn't match the user regex", func() {
			BeforeEach(func() {
				metadata.UserReturns("dora:some-guid")
			})

			It("fails the authentication", func() {
				Expect(authErr).To(MatchError("Invalid authentication domain"))
			})
		})

		Context("when the password doesn't match the provided credentials", func() {
			BeforeEach(func() {
				metadata.UserReturns("diego:some-guid/0")
				password = []byte("cf-user:cf-password")
			})

			It("fails the authentication", func() {
				Expect(authErr).To(MatchError("Invalid credentials"))
			})
		})
	})

	Describe("UserRegexp", func() {
		var regexp *regexp.Regexp

		BeforeEach(func() {
			regexp = authenticator.UserRegexp()
		})

		It("matches diego patterns", func() {
			Expect(regexp.MatchString("diego:guid/0")).To(BeTrue())
			Expect(regexp.MatchString("diego:123-abc-def/00")).To(BeTrue())
			Expect(regexp.MatchString("diego:guid/99")).To(BeTrue())
		})

		It("does not match other patterns", func() {
			Expect(regexp.MatchString("diego:some+guid/99")).To(BeFalse())
			Expect(regexp.MatchString("diego:..\\/something/99")).To(BeFalse())
			Expect(regexp.MatchString("diego:guid/")).To(BeFalse())
			Expect(regexp.MatchString("diego:00")).To(BeFalse())
			Expect(regexp.MatchString("diego:/00")).To(BeFalse())
			Expect(regexp.MatchString("cf:guid/0")).To(BeFalse())
			Expect(regexp.MatchString("cf:guid/99")).To(BeFalse())
			Expect(regexp.MatchString("user@guid/0")).To(BeFalse())
		})
	})
})
