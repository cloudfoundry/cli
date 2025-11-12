package shared

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/transport"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/command"

	// imported for the side effects
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 net/http.RoundTripper

func WrapForCFOnK8sAuth(config command.Config, k8sConfigGetter v7action.KubernetesConfigGetter, roundTripper http.RoundTripper) (http.RoundTripper, error) {
	username, err := config.CurrentUserName()
	if err != nil {
		return nil, err
	}
	if username == "" {
		return nil, errors.New("current user not set")
	}

	k8sConfig, err := k8sConfigGetter.Get()
	if err != nil {
		return nil, err
	}

	restConfig, err := clientcmd.NewDefaultClientConfig(
		*k8sConfig,
		&clientcmd.ConfigOverrides{
			Context: api.Context{AuthInfo: username},
		},
	).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Special case for certs, since we don't want mtls
	cert, err := getCert(restConfig)
	if err != nil {
		return nil, err
	}

	transportConfig, err := restConfig.TransportConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get transport config: %w", err)
	}

	if cert != nil {
		return certRoundTripper{
			cert:         cert,
			roundTripper: roundTripper,
		}, nil
	}

	if transportConfig.WrapTransport == nil {
		// i.e. not auth-provider or exec plugin
		return transport.HTTPWrappersForConfig(transportConfig, roundTripper)
	}

	// using auth provider to generate token
	return transportConfig.WrapTransport(roundTripper), nil
}

func getCert(restConfig *rest.Config) (*tls.Certificate, error) {
	tlsConfig, err := rest.TLSConfigFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get tls config: %w", err)
	}

	if tlsConfig != nil && tlsConfig.GetClientCertificate != nil {
		cert, err := tlsConfig.GetClientCertificate(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get client certificate: %w", err)
		}

		if len(cert.Certificate) > 0 && cert.PrivateKey != nil {
			return cert, nil
		}
	}
	return nil, nil
}

type certRoundTripper struct {
	cert         *tls.Certificate
	roundTripper http.RoundTripper
}

func (rt certRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var buf bytes.Buffer

	if err := pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: rt.cert.Certificate[0]}); err != nil {
		return nil, fmt.Errorf("could not convert certificate to PEM format: %w", err)
	}

	key, err := x509.MarshalPKCS8PrivateKey(rt.cert.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("could not marshal private key: %w", err)
	}

	if err := pem.Encode(&buf, &pem.Block{Type: "PRIVATE KEY", Bytes: key}); err != nil {
		return nil, fmt.Errorf("could not convert key to PEM format: %w", err)
	}

	auth := "ClientCert " + base64.StdEncoding.EncodeToString(buf.Bytes())
	req.Header.Set("Authorization", auth)

	return rt.roundTripper.RoundTrip(req)
}
