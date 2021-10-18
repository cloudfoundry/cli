package selfcontained_test

import (
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/v7/selfcontained/fake"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/configv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	apiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

var _ = Describe("client-certificate/key-data", func() {
	const (
		// taken from a real kubeconfig
		certData = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURJVENDQWdtZ0F3SUJBZ0lJVk9iMUFIckxNUjh3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRFd01EVXhOVEExTURsYUZ3MHlNakV3TURVeE5UQTFNVEZhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXJrdWxLbS9qTTJhZWZsdjkKK00zQk9Jc2QvVXZrRTBONGhWb3hSeWRBbE0xQXhWd3REYUdzL3dmUzRzb0xuNHJENTF3UE1SRlNJaitwSzdGYQprRGdaR0x4UFhrai96UkZOTzcvU3J2RHYwVGxjYjJENzNCS21qaXArQ2hBWkpQdWhMQlY2VnlTN0pXSWhOM1lOCktyamR5TnB5MHN3SjI1TW9CbW1saUpFc3V2dCtDaEhseERqWE9KenF1U2owa1hPQVVsWUFTN1dKK09JMU9HbzQKUjcvdHdHZlFTNW9oYXpRVVlDR2lZSllYcjVRNkVKTmJOVVI0RjdpRSthY1I5Rm9GNnNKSmkrQStET1VDUFFSKwptbjQ5Zm1pcFVHSGtMc3BicTNFZ0FEME40VW5jcmIyeUJEMFNVTmdLQmJjclY1S2hybFA2SzkwNkY5NEpubzNHCm1Id1JwUUlEQVFBQm8xWXdWREFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RBWURWUjBUQVFIL0JBSXdBREFmQmdOVkhTTUVHREFXZ0JUV2VNZ1ZBRkRhbWcraDRqS3hoRUh2Q1l5egp5akFOQmdrcWhraUc5dzBCQVFzRkFBT0NBUUVBUUxMWWFXQTRva1M2b3ZEWjQ1Z28ybkVZdUR4MklmRXZwYnh3CkNmYkFTNDY4M3lLT3FiYVBHOEpTVGhSbkh3TWlMcVBrbGFsdkJvV2R3aFB5Vkk0d2tVNHI4Y2c0UEpxNEZwWnQKVkNUQzFPZWVwRGpTMFpVQjRwSDVIZVlNQUxqSDBqcFV3RU96djFpaEtid05IMHFoZ2pGeUNTdld5TG9oZHdzbApJWXIvV1NEZm50NlBETC84TjFpcEJJbEN5Z1JHVGdoSFhPemhHUklPWG4rYWVOR29yWm9YWm0xbHErc1hyUnc5CktNdVZhRmdhaWVjSm0vbytyemFFSG9VZjRYOERKeVNubmVTa3ViaEx6ZERNc2o5eEs1cEJpdFgvaDlQMUQrMkcKeW5rcWdJVTJSWTM0SjBRcnU4Z0syNlJVT2pOcHIvRWJHQ0dUQUxiMXJnSDM0K2NFdlE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="

		// taken from a real kubeconfig
		keyData = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBcmt1bEttL2pNMmFlZmx2OStNM0JPSXNkL1V2a0UwTjRoVm94UnlkQWxNMUF4Vnd0CkRhR3Mvd2ZTNHNvTG40ckQ1MXdQTVJGU0lqK3BLN0Zha0RnWkdMeFBYa2ovelJGTk83L1NydkR2MFRsY2IyRDcKM0JLbWppcCtDaEFaSlB1aExCVjZWeVM3SldJaE4zWU5LcmpkeU5weTBzd0oyNU1vQm1tbGlKRXN1dnQrQ2hIbAp4RGpYT0p6cXVTajBrWE9BVWxZQVM3V0orT0kxT0dvNFI3L3R3R2ZRUzVvaGF6UVVZQ0dpWUpZWHI1UTZFSk5iCk5VUjRGN2lFK2FjUjlGb0Y2c0pKaStBK0RPVUNQUVIrbW40OWZtaXBVR0hrTHNwYnEzRWdBRDBONFVuY3JiMnkKQkQwU1VOZ0tCYmNyVjVLaHJsUDZLOTA2Rjk0Sm5vM0dtSHdScFFJREFRQUJBb0lCQUhhT0djQ3Q3cHN0K0N6MwpJZCsxQlBNVEhja0hjQkh5V081TUZTelM5RWpDL0lYQ1Z4cm9tQ1c1Q3FjY3k1OHY0bVZpNXlwM0g2a1FLR0FCCitSb1VMbTMyMlAzSEpjZzFOcFdudXdVbitkZ1U3bDBNV1FTdTR2L2N3M0xRMUdtbGgxSFhJaXBZWDd2ajdiZ3QKNG5MbTRHVVl3ZlhFdUFmVkp1cUZ2Z1VReXA0OEJ4cHdpQTMwakFKTVZNRW9pOUpIRnMwb29EMTlVc0Q4TmZCYQpIb2R6V3c4RkZpOGFrSHlsUEs3bUVUQ3RXaC9kOEs5WnAvUW1DLzhCMjFIWTY1R05rZFRlVXhwbWFIbnI1NDBNCmRnRnV1aGQ3UithWWZCTmN5d0QxUGNTMWtES0NiRnR4OXlPMDRSZnhhNStBejVNVDV2RTBGR2dCcUFqa1NHa3cKUXNsVjlRRUNnWUVBeU5adzMrWDloOEJvMXBEUG9qbDVXdVJsSUJDbG9YNHNZVzdUVGEwb3RPZlZienp3YWVjSgpwWGQ5S3lUL0ZjNlA1Nm9FSnh5WTM0UStuTEVkZVAzbnRMN2tadDNPcVlOL3ZPNmZNVlhNaXVNUXNBWDdVSjlICldHR2l0Z3kvQ2w2Q0cydkl1emd3V29lWnpYbmRLeVpFenFUVEEzWGZ5YlpTVG9KM0JLWGtUb1VDZ1lFQTNpcnYKdzd0YnVRRThITThpT0Z2WHJ2MWRETGFkTUk5N1lrTDlvelpxT3dETENxYWZoU1U5WTExMC91Z1RadXZZNmtFYQpiejlzSFdsQm8xbGlwR0JreElwaFFiSDUvdGd2RC9CdTBUYldvMExSTHdVN25lanV5bUVRSVREYzc2RUtmZm4yCnlYRlpRNXNPbU9KaGQwTmxpckgzamJpNUluTG8rVjAxTHN4QzhLRUNnWUVBdzBnZndOVjB5OGEvS2NMczRhUlkKNTVRazhDUy9xR055NjlwREE4aFhTSEFVREhHSW9yRFBwR0NuQ0lpVXI3T0dJWkFqTVpTUTRPVkpyRHBibWdYbwpnc3NLY3l0Qmczb2xmcVBYanJVS2loRjNiSENraVNRV3owMUFWUnZYdGxGbTJScll0NUpQRWZmQUNSZHhnajlYCnNFWjhGMlZLRUd5RFkydGdKdGpYUWJFQ2dZQS9NVFZMbGpJcnVEY3g5cFRFTlJOdk9hL0sydkp4Q283TXowTTMKdGxXVy9NM0Q5TXJ3NHp2SDFLQmZsUUZYeUFQRjljaWdGZ2taZ0lQa0E1dTBKNXJoOEg1allrMDZlS0ZDUm1BdQo4QXN6R3VGSEg1VHFnSVBUV3JKc3ptbzVXclFTU3hRd0RpYVd2VGIwczRlSGxjci9IVW5Ock5veit6MWJHSEQxCkk3SnlvUUtCZ1FDdjNRZllFd0l0RFlMT1g0a3BtUDJTY0d6Yjl1OWk4Q1puc1ZmcjhwSE82OHYrSkl1RDZKVU0KRHh1K1BjVmJUbkRMR21oVkROVkhhMHIra2RKL1hZMHBzUWN2YjJvTExFSDA5MGJLYUVBU2R5L1dwVytMT2xuNQpqeWJPZ0pBUngzYklsS21ZU0Z6bTcxekhnaU9va3ZlT2hpb2N3S3RsQlovQkQ1Vm9WRE1RR3c9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
	)

	var (
		apiConfig  fake.CFAPIConfig
		kubeConfig apiv1.Config
	)

	BeforeEach(func() {
		apiConfig = fake.CFAPIConfig{
			Routes: map[string]fake.Response{
				"GET /v3/apps": {
					Code: http.StatusOK, Body: map[string]interface{}{
						"pagination": map[string]interface{}{},
						"resources":  []resources.Application{},
					},
				},
			},
		}
		apiServer.SetConfiguration(apiConfig)
		helpers.SetConfig(func(config *configv3.Config) {
			config.ConfigFile.Target = apiServer.URL()
			config.ConfigFile.CFOnK8s.Enabled = true
			config.ConfigFile.CFOnK8s.AuthInfo = "one"
			config.ConfigFile.TargetedOrganization = configv3.Organization{
				GUID: "my-org",
				Name: "My Org",
			}

			config.ConfigFile.TargetedSpace = configv3.Space{
				GUID: "my-space",
				Name: "My Space",
			}
		})

		kubeConfig = apiv1.Config{
			Kind:       "Config",
			APIVersion: "v1",
			AuthInfos: []apiv1.NamedAuthInfo{
				{
					Name: "one", AuthInfo: apiv1.AuthInfo{
						ClientCertificateData: []byte(certData),
						ClientKeyData:         []byte(keyData),
					},
				},
			},
		}
		kubeConfigPath := filepath.Join(homeDir, ".kube", "config")
		storeKubeConfig(kubeConfig, kubeConfigPath)

		env = helpers.CFEnv{
			EnvVars: map[string]string{
				"KUBECONFIG": kubeConfigPath,
			},
		}
	})

	JustBeforeEach(func() {
		Eventually(helpers.CustomCF(env, "apps")).Should(gexec.Exit(0))
	})

	It("sends the client certificate and key in the Authorization header", func() {
		reqs := apiServer.ReceivedRequests()["GET /v3/apps"]
		Expect(reqs).To(HaveLen(1))
		Expect(reqs[0].Header).To(HaveKeyWithValue("Authorization", ConsistOf(HavePrefix("ClientCert "))))

		certAndKeyPEMBase64 := reqs[0].Header.Get("Authorization")[11:]
		certAndKeyPEM, err := base64.StdEncoding.DecodeString(certAndKeyPEMBase64)
		Expect(err).NotTo(HaveOccurred())

		cert, rest := pem.Decode(certAndKeyPEM)
		Expect(cert.Type).To(Equal(pemDecodeKubeConfigCertData(certData).Type))
		Expect(cert.Bytes).To(Equal(pemDecodeKubeConfigCertData(certData).Bytes))

		var key *pem.Block
		key, rest = pem.Decode(rest)
		Expect(key.Type).To(Equal(pemDecodeKubeConfigCertData(keyData).Type))
		Expect(key.Bytes).To(Equal(pemDecodeKubeConfigCertData(keyData).Bytes))

		Expect(rest).To(BeEmpty())
	})
})

func pemDecodeKubeConfigCertData(data string) *pem.Block {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	Expect(err).NotTo(HaveOccurred())
	pemDecodedBlock, rest := pem.Decode(decodedData)
	Expect(rest).To(BeEmpty())
	return pemDecodedBlock
}
