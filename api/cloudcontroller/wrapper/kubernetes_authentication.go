package wrapper

import (
	"errors"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/command"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

type KubernetesAuthentication struct {
	connection      cloudcontroller.Connection
	config          command.Config
	k8sConfigGetter v7action.KubernetesConfigGetter
	requiresAuth    bool
}

func NewKubernetesAuthentication(config command.Config, k8sConfigGetter v7action.KubernetesConfigGetter, requiresAuth bool) *KubernetesAuthentication {
	return &KubernetesAuthentication{
		config:          config,
		k8sConfigGetter: k8sConfigGetter,
		requiresAuth:    requiresAuth,
	}
}

func (a *KubernetesAuthentication) Make(request *cloudcontroller.Request, passedResponse *cloudcontroller.Response) error {
	if !a.requiresAuth {
		return a.connection.Make(request, passedResponse)
	}

	k8sConfig, err := a.k8sConfigGetter.Get()
	if err != nil {
		return err
	}

	username, err := a.config.CurrentUserName()
	if err != nil {
		return err
	}
	if username == "" {
		return errors.New("current user not set")
	}

	authInfo, ok := k8sConfig.AuthInfos[username]
	if !ok {
		return fmt.Errorf("auth info %q not present in kubeconfig", username)
	}

	pathOpts := clientcmd.NewDefaultPathOptions()
	persister := clientcmd.PersisterForUser(pathOpts, username)
	authProvider, err := rest.GetAuthProvider(a.config.Target(), authInfo.AuthProvider, persister)
	if err != nil {
		return err
	}

	wrappedRoundTripper := authProvider.WrapTransport(
		connectionRoundTripper{
			connection: a.connection,
			ccRequest:  request,
			ccResponse: passedResponse,
		})
	_, err = wrappedRoundTripper.RoundTrip(request.Request)
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
