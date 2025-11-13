package v7action

import (
	"errors"
	"fmt"
	"sort"

	"code.cloudfoundry.org/cli/v8/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/v8/api/uaa/constant"
	"code.cloudfoundry.org/cli/v8/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/v8/resources"
	"code.cloudfoundry.org/cli/v8/util/configv3"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . KubernetesConfigGetter

type KubernetesConfigGetter interface {
	Get() (*clientcmdapi.Config, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . WhoAmIer

type WhoAmIer interface {
	WhoAmI() (resources.K8sUser, ccv3.Warnings, error)
}

type DefaultKubernetesConfigGetter struct{}

func NewDefaultKubernetesConfigGetter() DefaultKubernetesConfigGetter {
	return DefaultKubernetesConfigGetter{}
}

func (c DefaultKubernetesConfigGetter) Get() (*clientcmdapi.Config, error) {
	pathOpts := clientcmd.NewDefaultPathOptions()
	return pathOpts.GetStartingConfig()
}

type kubernetesAuthActor struct {
	config          Config
	k8sConfigGetter KubernetesConfigGetter
	whoAmIer        WhoAmIer
}

func NewKubernetesAuthActor(config Config, k8sConfigGetter KubernetesConfigGetter, whoAmIer WhoAmIer) AuthActor {
	return &kubernetesAuthActor{
		config:          config,
		k8sConfigGetter: k8sConfigGetter,
		whoAmIer:        whoAmIer,
	}
}

func (actor kubernetesAuthActor) Authenticate(credentials map[string]string, origin string, grantType constant.GrantType) error {
	username := credentials["username"]
	availableUsernames, err := actor.getAvailableUsernames()
	if err != nil {
		return err
	}

	for _, u := range availableUsernames {
		if u == username {
			actor.config.SetKubernetesAuthInfo(username)
			return nil
		}
	}

	return errors.New("kubernetes user not found in configuration: " + username)
}

func (actor kubernetesAuthActor) GetLoginPrompts() (map[string]coreconfig.AuthPrompt, error) {
	availableUsernames, err := actor.getAvailableUsernames()
	if err != nil {
		return nil, err
	}
	sort.Strings(availableUsernames)

	return map[string]coreconfig.AuthPrompt{"username": {
		Type:        coreconfig.AuthPromptTypeMenu,
		Entries:     availableUsernames,
		DisplayName: "Choose your Kubernetes authentication info",
	}}, nil
}

func (actor kubernetesAuthActor) getAvailableUsernames() ([]string, error) {
	conf, err := actor.k8sConfigGetter.Get()
	if err != nil {
		return nil, err
	}

	if len(conf.AuthInfos) == 0 {
		return nil, errors.New("no kubernetes authentication infos configured")
	}

	var usernames []string
	for authInfo := range conf.AuthInfos {
		usernames = append(usernames, authInfo)
	}
	return usernames, nil
}

func (actor kubernetesAuthActor) GetCurrentUser() (configv3.User, error) {
	user, _, err := actor.whoAmIer.WhoAmI()
	if err != nil {
		return configv3.User{}, fmt.Errorf("calling /whoami endpoint failed: %w", err)
	}

	return configv3.User{Name: user.Name}, nil
}
