package wrapper

import (
	"net/http"

	"code.cloudfoundry.org/cli/v9/actor/v7action"
	"code.cloudfoundry.org/cli/v9/api/cloudcontroller"
	"code.cloudfoundry.org/cli/v9/api/shared"
	"code.cloudfoundry.org/cli/v9/command"
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
	roundTripper, err := shared.WrapForCFOnK8sAuth(a.config, a.k8sConfigGetter, connectionRoundTripper{
		connection: a.connection,
		ccRequest:  request,
		ccResponse: passedResponse,
	})
	if err != nil {
		return err
	}

	_, err = roundTripper.RoundTrip(request.Request)

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
