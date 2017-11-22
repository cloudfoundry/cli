package authenticators_test

import (
	"encoding/json"
	"net"

	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/diego-ssh/authenticators"
	"code.cloudfoundry.org/diego-ssh/routes"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_ssh"
	"code.cloudfoundry.org/lager/lagertest"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PermissionsBuilder", func() {
	Describe("Build", func() {
		var (
			logger         *lagertest.TestLogger
			expectedRoute  routes.SSHRoute
			desiredLRP     *models.DesiredLRP
			actualLRPGroup *models.ActualLRPGroup
			bbsClient      *fake_bbs.FakeInternalClient
			metadata       *fake_ssh.FakeConnMetadata

			permissionsBuilder authenticators.PermissionsBuilder
			permissions        *ssh.Permissions
			buildErr           error
			processGuid        string
			index              int
		)

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("test")

			expectedRoute = routes.SSHRoute{
				ContainerPort:   1111,
				PrivateKey:      "fake-pem-encoded-key",
				HostFingerprint: "host-fingerprint",
				User:            "user",
				Password:        "password",
			}

			diegoSSHRoutePayload, err := json.Marshal(expectedRoute)
			Expect(err).NotTo(HaveOccurred())

			diegoSSHRouteMessage := json.RawMessage(diegoSSHRoutePayload)

			desiredLRP = &models.DesiredLRP{
				ProcessGuid: "some-guid",
				Instances:   2,
				Routes: &models.Routes{
					routes.DIEGO_SSH: &diegoSSHRouteMessage,
				},
				LogGuid: "log-guid",
			}

			actualLRPGroup = &models.ActualLRPGroup{
				Instance: &models.ActualLRP{
					ActualLRPKey:         models.NewActualLRPKey("some-guid", 1, "some-domain"),
					ActualLRPInstanceKey: models.NewActualLRPInstanceKey("some-instance-guid", "some-cell-id"),
					ActualLRPNetInfo:     models.NewActualLRPNetInfo("1.2.3.4", models.NewPortMapping(3333, 1111)),
				},
			}

			bbsClient = new(fake_bbs.FakeInternalClient)
			bbsClient.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			bbsClient.DesiredLRPByProcessGuidReturns(desiredLRP, nil)

			permissionsBuilder = authenticators.NewPermissionsBuilder(bbsClient)

			remoteAddr, err := net.ResolveIPAddr("ip", "1.1.1.1")
			Expect(err).NotTo(HaveOccurred())
			metadata = &fake_ssh.FakeConnMetadata{}
			metadata.RemoteAddrReturns(remoteAddr)

			processGuid = "some-guid"
			index = 1
		})

		JustBeforeEach(func() {
			permissions, buildErr = permissionsBuilder.Build(logger, processGuid, index, metadata)
		})

		It("gets information about the desired lrp referenced in the username", func() {
			Expect(bbsClient.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
			_, guid := bbsClient.DesiredLRPByProcessGuidArgsForCall(0)
			Expect(guid).To(Equal("some-guid"))
		})

		It("gets information about the the actual lrp from the username", func() {
			Expect(bbsClient.ActualLRPGroupByProcessGuidAndIndexCallCount()).To(Equal(1))

			_, guid, index := bbsClient.ActualLRPGroupByProcessGuidAndIndexArgsForCall(0)
			Expect(guid).To(Equal("some-guid"))
			Expect(index).To(Equal(1))
		})

		It("saves container information in the critical options of the permissions", func() {
			expectedConfig := `{
				"address": "1.2.3.4:3333",
				"host_fingerprint": "host-fingerprint",
				"private_key": "fake-pem-encoded-key",
				"user": "user",
				"password": "password"
			}`

			Expect(permissions).NotTo(BeNil())
			Expect(permissions.CriticalOptions).NotTo(BeNil())
			Expect(permissions.CriticalOptions["proxy-target-config"]).To(MatchJSON(expectedConfig))
		})

		It("saves log message information in the critical options of the permissions", func() {
			expectedConfig := `{
				"guid": "log-guid",
				"message": "Successful remote access by 1.1.1.1",
				"index": 1
			}`

			Expect(permissions).NotTo(BeNil())
			Expect(permissions.CriticalOptions).NotTo(BeNil())
			Expect(permissions.CriticalOptions["log-message"]).To(MatchJSON(expectedConfig))
		})

		Context("when getting the desired LRP information fails", func() {
			BeforeEach(func() {
				bbsClient.DesiredLRPByProcessGuidReturns(nil, &models.Error{})
			})

			It("returns the error", func() {
				Expect(buildErr).To(Equal(&models.Error{}))
			})
		})

		Context("when getting the actual LRP information fails", func() {
			BeforeEach(func() {
				bbsClient.ActualLRPGroupByProcessGuidAndIndexReturns(nil, &models.Error{})
			})

			It("returns the error", func() {
				Expect(buildErr).To(Equal(&models.Error{}))
			})
		})

		Context("when the container port cannot be found", func() {
			BeforeEach(func() {
				actualLRPGroup.Instance.Ports = []*models.PortMapping{}
				bbsClient.ActualLRPGroupByProcessGuidAndIndexReturns(actualLRPGroup, nil)
			})

			It("returns an empty permission reference", func() {
				Expect(permissions).To(Equal(&ssh.Permissions{}))
			})
		})

		Context("when the desired LRP does not include routes", func() {
			BeforeEach(func() {
				desiredLRP.Routes = nil
				bbsClient.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
			})

			It("fails the authentication", func() {
				Expect(buildErr).To(Equal(authenticators.RouteNotFoundErr))
			})
		})

		Context("when the desired LRP does not include an SSH route", func() {
			BeforeEach(func() {
				r := *desiredLRP.Routes
				delete(r, routes.DIEGO_SSH)
				bbsClient.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
			})

			It("fails the authentication", func() {
				Expect(buildErr).To(Equal(authenticators.RouteNotFoundErr))
			})
		})

		Context("when the ssh route fails to unmarshal", func() {
			BeforeEach(func() {
				message := json.RawMessage([]byte(`{,:`))
				(*desiredLRP.Routes)[routes.DIEGO_SSH] = &message
				bbsClient.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
			})

			It("fails the authentication", func() {
				Expect(buildErr).To(HaveOccurred())
			})
		})
	})
})
