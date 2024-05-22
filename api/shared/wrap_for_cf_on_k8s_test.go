package shared_test

import (
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/shared"
	"code.cloudfoundry.org/cli/api/shared/sharedfakes"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/integration/helpers"

	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientauthenticationv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/tools/clientcmd/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	clientCertData = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJVk9iMUFIckxNUjh3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRFd01EVXhOVEExTURsYUZ3MHlNakV3TURVeE5UQTFNVEZhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXJrdWxLbS9qTTJhZWZsdjkKK00zQk9Jc2QvVXZrRTBONGhWb3hSeWRBbE0xQXhWd3REYUdzL3dmUzRzb0xuNHJENTF3UE1SRlNJaitwSzdGYQprRGdaR0x4UFhrai96UkZOTzcvU3J2RHYwVGxjYjJENzNCS21qaXArQ2hBWkpQdWhMQlY2VnlTN0pXSWhOM1lOCktyamR5TnB5MHN3SjI1TW9CbW1saUpFc3V2dCtDaEhseERqWE9KenF1U2owa1hPQVVsWUFTN1dKK09JMU9HbzQKUjcvdHdHZlFTNW9oYXpRVVlDR2lZSllYcjVRNkVKTmJOVVI0RjdpRSthY1I5Rm9GNnNKSmkrQStET1VDUFFSKwptbjQ5Zm1pcFVHSGtMc3BicTNFZ0FEME40VW5jcmIyeUJEMFNVTmdLQmJjclY1S2hybFA2SzkwNkY5NEpubzNHCm1Id1JwUUlEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JUV2VNZ1ZBRkRhbWcraDRqS3hoRUh2Q1l5egp5akFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBUUxMWWFXQTRva1M2b3ZEWjQ1Z28ybkVZdUR4MklmRXZwYnh3CkNmYkFTNDY4M3lLT3FiYVBHOEpTVGhSbkh3TWlMcVBrbGFsdkJvV2R3aFB5Vkk0d2tVNHI4Y2c0UEpxNEZwWnQKVkNUQzFPZWVwRGpTMFpVQjRwSDVIZVlNQUxqSDBqcFV3RU96djFpaEtid05IMHFoZ2pGeUNTdld5TG9oZHdzbApJWXIvV1NEZm50NlBETC84TjFpcEJJbEN5Z1JHVGdoSFhPemhHUklPWG4rYWVOR29yWm9YWm0xbHErc1hyUnc5CktNdVZhRmdhaWVjSm0vbytyemFFSG9VZjRYOERKeVNubmVTa3ViaEx6ZERNc2o5eEs1cEJpdFgvaDlQMUQrMkcKeW5rcWdJVTJSWTM0SjBRcnU4Z0syNlJVT2pOcHIvRWJHQ0dUQUxiMXJnSDM0K2NFdlE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	clientKeyData  = "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRQ3VTNlVxYitNelpwNSsKVy8zNHpjRTRpeDM5UytRVFEzaUZXakZISjBDVXpVREZYQzBOb2F6L0I5TGl5Z3VmaXNQblhBOHhFVklpUDZrcgpzVnFRT0JrWXZFOWVTUC9ORVUwN3Y5S3U4Ty9ST1Z4dllQdmNFcWFPS240S0VCa2srNkVzRlhwWEpMc2xZaUUzCmRnMHF1TjNJMm5MU3pBbmJreWdHYWFXSWtTeTYrMzRLRWVYRU9OYzRuT3E1S1BTUmM0QlNWZ0JMdFluNDRqVTQKYWpoSHYrM0FaOUJMbWlGck5CUmdJYUpnbGhldmxEb1FrMXMxUkhnWHVJVDVweEgwV2dYcXdrbUw0RDRNNVFJOQpCSDZhZmoxK2FLbFFZZVF1eWx1cmNTQUFQUTNoU2R5dHZiSUVQUkpRMkFvRnR5dFhrcUd1VS9vcjNUb1gzZ21lCmpjYVlmQkdsQWdNQkFBRUNnZ0VBZG80WndLM3VteTM0TFBjaDM3VUU4eE1keVFkd0VmSlk3a3dWTE5MMFNNTDgKaGNKWEd1aVlKYmtLcHh6TG55L2laV0xuS25jZnFSQW9ZQUg1R2hRdWJmYlkvY2NseURVMmxhZTdCU2Y1MkJUdQpYUXhaQks3aS85ekRjdERVYWFXSFVkY2lLbGhmdStQdHVDM2ljdWJnWlJqQjljUzRCOVVtNm9XK0JSREtuandICkduQ0lEZlNNQWt4VXdTaUwwa2NXelNpZ1BYMVN3UHcxOEZvZWgzTmJEd1VXTHhxUWZLVThydVlSTUsxYUg5M3cKcjFtbjlDWUwvd0hiVWRqcmtZMlIxTjVUR21ab2Vldm5qUXgyQVc2NkYzdEg1cGg4RTF6TEFQVTl4TFdRTW9KcwpXM0gzSTdUaEYvRnJuNERQa3hQbThUUVVhQUdvQ09SSWFUQkN5VlgxQVFLQmdRREkxbkRmNWYySHdHaldrTStpCk9YbGE1R1VnRUtXaGZpeGhidE5OclNpMDU5VnZQUEJwNXdtbGQzMHJKUDhWem8vbnFnUW5ISmpmaEQ2Y3NSMTQKL2VlMHZ1Um0zYzZwZzMrODdwOHhWY3lLNHhDd0JmdFFuMGRZWWFLMkRMOEtYb0liYThpN09EQmFoNW5OZWQwcgpKa1RPcE5NRGRkL0p0bEpPZ25jRXBlUk9oUUtCZ1FEZUt1L0R1MXU1QVR3Y3p5STRXOWV1L1YwTXRwMHdqM3RpClF2MmpObW83QU1zS3BwK0ZKVDFqWFhUKzZCTm02OWpxUVJwdlAyd2RhVUdqV1dLa1lHVEVpbUZCc2ZuKzJDOFAKOEc3Uk50YWpRdEV2QlR1ZDZPN0tZUkFoTU56dm9RcDkrZmJKY1ZsRG13Nlk0bUYzUTJXS3NmZU51TGtpY3VqNQpYVFV1ekVMd29RS0JnUUREU0IvQTFYVEx4cjhwd3V6aHBGam5sQ1R3Skwrb1kzTHIya01EeUZkSWNCUU1jWWlpCnNNK2tZS2NJaUpTdnM0WWhrQ014bEpEZzVVbXNPbHVhQmVpQ3l3cHpLMEdEZWlWK285ZU90UXFLRVhkc2NLU0oKSkJiUFRVQlZHOWUyVVdiWkd0aTNrazhSOThBSkYzR0NQMWV3Um53WFpVb1FiSU5qYTJBbTJOZEJzUUtCZ0Q4eApOVXVXTWl1NE56SDJsTVExRTI4NXI4cmE4bkVLanN6UFF6ZTJWWmI4emNQMHl2RGpPOGZVb0YrVkFWZklBOFgxCnlLQVdDUm1BZytRRG03UW5tdUh3Zm1OaVRUcDRvVUpHWUM3d0N6TWE0VWNmbE9xQWc5TmFzbXpPYWpsYXRCSkwKRkRBT0pwYTlOdlN6aDRlVnl2OGRTYzJzMmpQN1BWc1ljUFVqc25LaEFvR0JBSy9kQjlnVEFpME5nczVmaVNtWQovWkp3Yk52MjcyTHdKbWV4Vit2eWtjN3J5LzRraTRQb2xRd1BHNzQ5eFZ0T2NNc2FhRlVNMVVkclN2NlIwbjlkCmpTbXhCeTl2YWdzc1FmVDNSc3BvUUJKM0w5YWxiNHM2V2ZtUEpzNkFrQkhIZHNpVXFaaElYT2J2WE1lQ0k2aVMKOTQ2R0toekFxMlVGbjhFUGxXaFVNeEFiCi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K"
)

var _ = Describe("WrapForCFOnK8sAuth", func() {
	var (
		config              *commandfakes.FakeConfig
		k8sConfigGetter     *v7actionfakes.FakeKubernetesConfigGetter
		req                 *http.Request
		res                 *http.Response
		actualRes           *http.Response
		kubeConfig          *api.Config
		wrapErr             error
		wrappedRoundTripper *sharedfakes.FakeRoundTripper
	)

	BeforeEach(func() {
		kubeConfig = &api.Config{
			Kind:       "Config",
			APIVersion: "v1",
			Clusters: map[string]*api.Cluster{
				"my-cluster": {
					Server: "https://example.org",
				},
			},
			Contexts: map[string]*api.Context{
				"my-context": {
					Cluster:   "my-cluster",
					AuthInfo:  "my-auth-info",
					Namespace: "my-namespace",
				},
			},
			CurrentContext: "my-context",
			AuthInfos:      map[string]*api.AuthInfo{},
		}

		k8sConfigGetter = new(v7actionfakes.FakeKubernetesConfigGetter)
		k8sConfigGetter.GetReturns(kubeConfig, nil)

		config = new(commandfakes.FakeConfig)
		config.CurrentUserNameReturns("auth-test", nil)

		var err error
		req, err = http.NewRequest(http.MethodPost, "", strings.NewReader("hello"))
		Expect(err).NotTo(HaveOccurred())

		wrappedRoundTripper = new(sharedfakes.FakeRoundTripper)
		res = &http.Response{StatusCode: http.StatusTeapot}

		wrappedRoundTripper.RoundTripReturns(res, nil)
		actualRes = nil
	})

	JustBeforeEach(func() {
		var roundTripper http.RoundTripper
		roundTripper, wrapErr = shared.WrapForCFOnK8sAuth(config, k8sConfigGetter, wrappedRoundTripper)

		if wrapErr == nil {
			actualRes, wrapErr = roundTripper.RoundTrip(req)
		}
	})

	When("getting the k8s config fails", func() {
		BeforeEach(func() {
			k8sConfigGetter.GetReturns(nil, errors.New("boom!"))
		})

		It("returns the error", func() {
			Expect(wrapErr).To(MatchError("boom!"))
		})
	})

	When("no user is set in the config", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("", nil)
		})

		It("errors", func() {
			Expect(wrapErr).To(MatchError(ContainSubstring("current user not set")))
		})
	})

	When("there is an error getting the current user from the config", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("", errors.New("boom"))
		})

		It("errors", func() {
			Expect(wrapErr).To(MatchError(ContainSubstring("boom")))
		})
	})

	When("the chosen kubernetes auth info is not present in kubeconfig", func() {
		BeforeEach(func() {
			config.CurrentUserNameReturns("not-present", nil)
		})

		It("errors", func() {
			Expect(wrapErr).To(MatchError(ContainSubstring(`auth info "not-present" does not exist`)))
		})
	})

	checkCalls := func() *http.Request {
		Expect(wrapErr).NotTo(HaveOccurred())
		Expect(wrappedRoundTripper.RoundTripCallCount()).To(Equal(1))

		actualReq := wrappedRoundTripper.RoundTripArgsForCall(0)

		body, err := ioutil.ReadAll(actualReq.Body)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(body)).To(Equal("hello"))

		Expect(actualRes).To(Equal(res))

		return actualReq
	}

	checkBearerTokenInAuthHeader := func() {
		actualReq := checkCalls()

		token, err := jws.ParseJWTFromRequest(actualReq)
		Expect(err).NotTo(HaveOccurred())
		Expect(token.Validate(keyPair.Public(), crypto.SigningMethodRS256)).To(Succeed())

		claims := token.Claims()
		Expect(claims).To(HaveKeyWithValue("another", "thing"))
	}

	checkClientCertInAuthHeader := func() {
		actualReq := checkCalls()

		Expect(actualReq.Header).To(HaveKeyWithValue("Authorization", ConsistOf(HavePrefix("ClientCert "))))

		certAndKeyPEMBase64 := actualReq.Header.Get("Authorization")[11:]
		certAndKeyPEM, err := base64.StdEncoding.DecodeString(certAndKeyPEMBase64)
		Expect(err).NotTo(HaveOccurred())

		cert, rest := pem.Decode(certAndKeyPEM)
		Expect(cert.Type).To(Equal(pemDecodeKubeConfigCertData(clientCertData).Type))
		Expect(cert.Bytes).To(Equal(pemDecodeKubeConfigCertData(clientCertData).Bytes))

		var key *pem.Block
		key, rest = pem.Decode(rest)
		Expect(key.Bytes).To(Equal(pemDecodeKubeConfigCertData(clientKeyData).Bytes))

		Expect(rest).To(BeEmpty())
	}

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

			kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
				AuthProvider: &api.AuthProviderConfig{
					Name: "oidc",
					Config: map[string]string{
						"id-token":       string(token),
						"idp-issuer-url": "-",
						"client-id":      "-",
					},
				},
			}
		})

		It("uses the auth-provider to generate the Bearer token", func() {
			checkBearerTokenInAuthHeader()
		})
	})

	Describe("client certs", func() {
		var (
			certFilePath string
			keyFilePath  string
		)

		BeforeEach(func() {
			certFilePath = writeToFile(clientCertData)
			keyFilePath = writeToFile(clientKeyData)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(certFilePath)).To(Succeed())
			Expect(os.RemoveAll(keyFilePath)).To(Succeed())
		})

		When("inline cert and key are provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificateData: []byte(base64Decode(clientCertData)),
					ClientKeyData:         []byte(base64Decode(clientKeyData)),
				}
			})

			It("puts concatenated client ceritificate and key data into the Authorization header", func() {
				checkClientCertInAuthHeader()
			})
		})

		When("cert and key are provided in files", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificate: certFilePath,
					ClientKey:         keyFilePath,
				}
			})

			It("puts concatenated client ceritificate and key data into the Authorization header", func() {
				checkClientCertInAuthHeader()
			})

			When("cert file cannot be read", func() {
				BeforeEach(func() {
					Expect(os.Remove(certFilePath)).To(Succeed())
				})

				It("returns an error", func() {
					Expect(wrapErr).To(MatchError(ContainSubstring(certFilePath)))
				})
			})

			When("key file cannot be read", func() {
				BeforeEach(func() {
					Expect(os.Remove(keyFilePath)).To(Succeed())
				})

				It("returns an error", func() {
					Expect(wrapErr).To(MatchError(ContainSubstring(keyFilePath)))
				})
			})
		})

		When("file and inline cert is provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificate:     certFilePath,
					ClientCertificateData: []byte(base64Decode(clientCertData)),
					ClientKeyData:         []byte(base64Decode(clientKeyData)),
				}
			})

			It("complains about invalid configuration", func() {
				Expect(wrapErr).To(MatchError(ContainSubstring("client-cert-data and client-cert are both specified")))
			})
		})

		When("file and inline key is provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificateData: []byte(base64Decode(clientCertData)),
					ClientKeyData:         []byte(base64Decode(clientKeyData)),
					ClientKey:             keyFilePath,
				}
			})

			It("complains about invalid configuration", func() {
				Expect(wrapErr).To(MatchError(ContainSubstring("client-key-data and client-key are both specified")))
			})
		})

		When("inline cert and key file are provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificateData: []byte(base64Decode(clientCertData)),
					ClientKey:             keyFilePath,
				}
			})

			It("uses the inline key", func() {
				checkClientCertInAuthHeader()
			})
		})

		When("cert file and inline key are provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					ClientCertificate: certFilePath,
					ClientKeyData:     []byte(base64Decode(clientKeyData)),
				}
			})

			It("uses the inline key", func() {
				checkClientCertInAuthHeader()
			})
		})
	})

	Describe("exec", func() {
		BeforeEach(func() {
			kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
				Exec: &api.ExecConfig{
					APIVersion:      "client.authentication.k8s.io/v1beta1",
					InteractiveMode: "Never",
					Command:         "echo",
				},
			}
		})

		When("the command returns a token", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"].Exec.Args = []string{execCredential(&clientauthenticationv1beta1.ExecCredentialStatus{
					Token: "a-token",
				})}
			})

			It("uses the exec command to generate the Bearer token", func() {
				helpers.SkipIfWindows() // We're getting "plugin returned version client.authentication.k8s.io/__internal" on Windows. This issue is unresolved in upstream library.

				Expect(wrapErr).NotTo(HaveOccurred())
				Expect(wrappedRoundTripper.RoundTripCallCount()).To(Equal(1))

				actualReq := wrappedRoundTripper.RoundTripArgsForCall(0)
				Expect(actualReq.Header.Get("Authorization")).To(Equal("Bearer a-token"))

				Expect(actualRes).To(Equal(res))
			})
		})

		When("the command returns a client cert and key", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"].Exec.Args = []string{execCredential(&clientauthenticationv1beta1.ExecCredentialStatus{
					ClientCertificateData: base64Decode(clientCertData),
					ClientKeyData:         base64Decode(clientKeyData),
				})}
			})

			It("uses the exec command to generate client certs", func() {
				checkClientCertInAuthHeader()
			})
		})
	})

	Describe("tokens provided in config", func() {
		var token []byte

		BeforeEach(func() {
			jwt := jws.NewJWT(jws.Claims{
				"exp":     time.Now().Add(time.Hour).Unix(),
				"another": "thing",
			}, crypto.SigningMethodRS256)
			var err error
			token, err = jwt.Serialize(keyPair)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("inline tokens", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					Token: string(token),
				}
			})

			It("inserts the token in the authorization header", func() {
				checkBearerTokenInAuthHeader()
			})
		})

		Context("token file paths", func() {
			var tokenFilePath string

			BeforeEach(func() {
				tokenFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				defer tokenFile.Close()
				_, err = tokenFile.Write(token)
				Expect(err).NotTo(HaveOccurred())
				tokenFilePath = tokenFile.Name()
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					TokenFile: tokenFilePath,
				}
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tokenFilePath)).To(Succeed())
			})

			It("inserts the token in the authorization header", func() {
				checkBearerTokenInAuthHeader()
			})
		})

		When("both file and inline token are provided", func() {
			BeforeEach(func() {
				kubeConfig.AuthInfos["auth-test"] = &api.AuthInfo{
					Token:     string(token),
					TokenFile: "some-path",
				}
			})

			It("the inline token takes precedence", func() {
				checkBearerTokenInAuthHeader()
			})
		})
	})
})

func pemDecodeKubeConfigCertData(data string) *pem.Block {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	Expect(err).NotTo(HaveOccurred())
	pemDecodedBlock, rest := pem.Decode(decodedData)
	Expect(rest).To(BeEmpty())
	return pemDecodedBlock
}

func base64Decode(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	Expect(err).NotTo(HaveOccurred())
	return string(decoded)
}

func writeToFile(base64Data string) string {
	file, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())
	file.WriteString(base64Decode(base64Data))
	Expect(file.Close()).To(Succeed())
	return file.Name()
}

func execCredential(status *clientauthenticationv1beta1.ExecCredentialStatus) string {
	execCred, err := json.Marshal(clientauthenticationv1beta1.ExecCredential{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "client.authentication.k8s.io/v1beta1",
		},
		Status: status,
	})
	Expect(err).NotTo(HaveOccurred())
	return string(execCred)
}
