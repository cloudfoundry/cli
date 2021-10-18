package wrapper_test

import (
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/ccv3fakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/wrapper"
	"code.cloudfoundry.org/cli/command/commandfakes"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd/api"
)

var _ = Describe("KubernetesAuthentication", func() {
	var (
		k8sAuthWrapper    *wrapper.KubernetesAuthentication
		config            *commandfakes.FakeConfig
		k8sConfigGetter   *v7actionfakes.FakeKubernetesConfigGetter
		wrappedConnection *ccv3fakes.FakeConnectionWrapper
		req               *cloudcontroller.Request
		resp              *cloudcontroller.Response
		err               error
	)

	BeforeEach(func() {
		k8sConfigGetter = new(v7actionfakes.FakeKubernetesConfigGetter)
		k8sConfigGetter.GetReturns(&api.Config{}, nil)

		config = new(commandfakes.FakeConfig)
		config.CurrentUserNameReturns("auth-test", nil)

		wrappedConnection = new(ccv3fakes.FakeConnectionWrapper)

		httpReq, err := http.NewRequest(http.MethodPost, "", strings.NewReader("hello"))
		Expect(err).NotTo(HaveOccurred())
		req = cloudcontroller.NewRequest(httpReq, nil)

		resp = &cloudcontroller.Response{
			HTTPResponse: &http.Response{
				StatusCode: http.StatusTeapot,
			},
		}
	})

	JustBeforeEach(func() {
		k8sAuthWrapper = wrapper.NewKubernetesAuthentication(config, k8sConfigGetter)
		k8sAuthWrapper.Wrap(wrappedConnection)

		err = k8sAuthWrapper.Make(req, resp)
	})

	When("getting the k8s config fails", func() {
		BeforeEach(func() {
			k8sConfigGetter.GetReturns(nil, errors.New("boom!"))
		})

		It("returns the error", func() {
			Expect(err).To(MatchError("boom!"))
		})
	})

	When("no user is set in the config", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("", nil)
		})

		It("errors", func() {
			Expect(err).To(MatchError(ContainSubstring("current user not set")))
		})
	})

	When("there is an error getting the current user from the config", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("", errors.New("boom"))
		})

		It("errors", func() {
			Expect(err).To(MatchError(ContainSubstring("boom")))
		})
	})

	When("the chosen kubeernetes auth info is not present in kubeconfig", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("not-present", nil)
		})

		It("errors", func() {
			Expect(err).To(MatchError(ContainSubstring(`auth info "not-present" not present in kubeconfig`)))
		})
	})

	Describe("auth-provider", func() {
		var token []byte

		BeforeEach(func() {
			jwt := jws.NewJWT(jws.Claims{
				"exp":     time.Now().Add(time.Hour).Unix(),
				"another": "thing",
			}, crypto.SigningMethodRS256)
			var err error
			token, err = jwt.Serialize(keyPair)
			Expect(err).NotTo(HaveOccurred())

			k8sConfigGetter.GetReturns(&api.Config{
				Kind:       "Config",
				APIVersion: "v1",
				AuthInfos: map[string]*api.AuthInfo{
					"auth-test": {
						AuthProvider: &api.AuthProviderConfig{
							Name: "oidc",
							Config: map[string]string{
								"id-token":       string(token),
								"idp-issuer-url": "-",
								"client-id":      "-",
							},
						},
					},
				},
			}, nil)
		})

		It("uses the auth-provider to generate the Bearer token", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(wrappedConnection.MakeCallCount()).To(Equal(1))

			actualReq, actualResp := wrappedConnection.MakeArgsForCall(0)
			Expect(actualResp.HTTPResponse).To(HaveHTTPStatus(http.StatusTeapot))

			token, err := jws.ParseJWTFromRequest(actualReq.Request)
			Expect(err).NotTo(HaveOccurred())
			Expect(token.Validate(keyPair.Public(), crypto.SigningMethodRS256)).To(Succeed())

			claims := token.Claims()
			Expect(claims).To(HaveKeyWithValue("another", "thing"))

			body, err := ioutil.ReadAll(actualReq.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(Equal("hello"))
		})
	})

	Describe("client-cert/key-data", func() {
		const (
			clientCertData = "client-cert-data"
			clientKeyData  = "client-key-data"
		)

		BeforeEach(func() {
			k8sConfigGetter.GetReturns(&api.Config{
				Kind:       "Config",
				APIVersion: "v1",
				AuthInfos: map[string]*api.AuthInfo{
					"auth-test": {
						ClientCertificateData: []byte(clientCertData),
						ClientKeyData:         []byte(clientKeyData),
					},
				},
			}, nil)
		})

		It("puts concatenated client ceritificate and key data into the Authorization header", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(wrappedConnection.MakeCallCount()).To(Equal(1))

			actualReq, actualResp := wrappedConnection.MakeArgsForCall(0)
			Expect(actualResp.HTTPResponse).To(HaveHTTPStatus(http.StatusTeapot))

			Expect(actualReq.Header).To(HaveKeyWithValue("Authorization", ConsistOf(HavePrefix("ClientCert "))))

			certAndKey := actualReq.Header.Get("Authorization")[11:]
			certAndKeyDecoded, err := base64.StdEncoding.DecodeString(string(certAndKey))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(certAndKeyDecoded)).To(Equal(clientCertData + clientKeyData))
		})
	})

	Describe("unsupported authentication method", func() {
		BeforeEach(func() {
			k8sConfigGetter.GetReturns(&api.Config{
				Kind:       "Config",
				APIVersion: "v1",
				AuthInfos: map[string]*api.AuthInfo{
					"auth-test": {},
				},
			}, nil)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError(ContainSubstring("authentication method not supported")))
		})
	})
})
