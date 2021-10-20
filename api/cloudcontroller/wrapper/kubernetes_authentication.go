package wrapper

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type KubernetesAuthentication struct {
	connection      cloudcontroller.Connection
	config          command.Config
	k8sConfigGetter v7action.KubernetesConfigGetter
}

func NewKubernetesAuthentication(
	config command.Config,
	k8sConfigGetter v7action.KubernetesConfigGetter,
) *KubernetesAuthentication {

	return &KubernetesAuthentication{
		config:          config,
		k8sConfigGetter: k8sConfigGetter,
	}
}

func (a *KubernetesAuthentication) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	username, err := a.config.CurrentUserName()
	if err != nil {
		return err
	}
	if username == "" {
		return errors.New("current user not set")
	}

	k8sConfig, err := a.k8sConfigGetter.Get()
	if err != nil {
		return err
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(
		*k8sConfig,
		&clientcmd.ConfigOverrides{
			Context: api.Context{AuthInfo: username},
		}).ClientConfig()
	if err != nil {
		return err
	}

	tlsConfig, err := rest.TLSConfigFor(restConfig)
	if err != nil {
		return fmt.Errorf("failed to get tls config: %w", err)
	}

	if tlsConfig != nil && tlsConfig.GetClientCertificate != nil {
		cert, err := tlsConfig.GetClientCertificate(nil)
		if err != nil {
			return fmt.Errorf("failed to get client certificate: %w", err)
		}

		if len(cert.Certificate) > 0 && cert.PrivateKey != nil {
			var buf bytes.Buffer

			if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}); err != nil {
				return fmt.Errorf("could not convert certificate to PEM format: %w", err)
			}

			key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
			if err != nil {
				return fmt.Errorf("could not marshal private key: %w", err)
			}

			if err := pem.Encode(&buf, &pem.Block{Type: "PRIVATE KEY", Bytes: key}); err != nil {
				return fmt.Errorf("could not convert key to PEM format: %w", err)
			}

			auth := "ClientCert " + base64.StdEncoding.EncodeToString(buf.Bytes())
			request.Header.Set("Authorization", auth)

			return a.connection.Make(request, passedResponse)
		}
	}

	transportConfig, err := restConfig.TransportConfig()
	if err != nil {
		return fmt.Errorf("failed to get transport config: %w", err)
	}

	if transportConfig.WrapTransport == nil {
		return fmt.Errorf("authentication method not supported")
	}

	_, err = transportConfig.WrapTransport(connectionRoundTripper{
		connection: a.connection,
		ccRequest:  request,
		ccResponse: passedResponse,
	}).RoundTrip(request.Request)

	return err
}

func (a *KubernetesAuthentication) Wrap(innerconnection cloudcontroller.Connection) cloudcontroller.Connection {
	a.connection = innerconnection

	return a
}

type connectionRoundTripper struct {
	connection cloudcontroller.Connection
	ccRequest  *cloudcontroller.Request
	ccResponse *cloudcontroller.Response
}

func (rt connectionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// The passed `*req` is a shallow clone of the original `*req` with the auth header added.
	// So we need to reset it on the `ccRequest`.
	rt.ccRequest.Request = req

	err := rt.connection.Make(rt.ccRequest, rt.ccResponse)
	if err != nil {
		return nil, err
	}

	return rt.ccResponse.HTTPResponse, nil
}
